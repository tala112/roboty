package modes

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ForegroundActivity struct {
	AppName     string    `json:"app_name"`
	ExecName    string    `json:"exec_name"`
	WindowTitle string    `json:"window_title"`
	PID         int       `json:"pid"`
	Timestamp   time.Time `json:"timestamp"`
}

type ForegroundTracker struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewForegroundTracker() *ForegroundTracker {
	return &ForegroundTracker{}
}

func (ft *ForegroundTracker) Start(ctx context.Context, interval time.Duration, callback func(ForegroundActivity)) {
	ft.mu.Lock()
	if ft.cancel != nil {
		ft.cancel()
	}
	ft.ctx, ft.cancel = context.WithCancel(ctx)
	ft.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ft.ctx.Done():
				return
			case <-ticker.C:
				activity, err := ft.Poll()
				if err != nil {
					log.Printf("[tracker] poll error: %v", err)
					continue
				}
				if activity != nil {
					callback(*activity)
				}
			}
		}
	}()
}

func (ft *ForegroundTracker) Stop() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if ft.cancel != nil {
		ft.cancel()
		ft.cancel = nil
	}
}

func (ft *ForegroundTracker) Poll() (*ForegroundActivity, error) {
	switch runtime.GOOS {
	case "linux":
		return ft.pollLinux()
	case "darwin":
		return ft.pollMacOS()
	case "windows":
		return ft.pollWindows()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// =============================================================================
// Linux — using xdotool + /proc (fallback to xprop)
// =============================================================================

func (ft *ForegroundTracker) pollLinux() (*ForegroundActivity, error) {
	pid, err := ft.getActivePIDLinux()
	if err != nil || pid == 0 {
		return nil, err
	}

	execName := ft.getProcessNameLinux(pid)
	windowTitle := ft.getWindowTitleLinux()

	if execName == "" {
		return nil, nil
	}

	return &ForegroundActivity{
		AppName:     ft.friendlyAppName(execName),
		ExecName:    execName,
		WindowTitle: windowTitle,
		PID:         pid,
		Timestamp:   time.Now(),
	}, nil
}

func (ft *ForegroundTracker) getActivePIDLinux() (int, error) {
	// Try xdotool first (most reliable on X11)
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowpid")
	out, err := cmd.Output()
	if err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil && pid > 0 {
			return pid, nil
		}
	}

	// Fallback: xprop + /proc
	cmd = exec.Command("xprop", "-root", "_NET_ACTIVE_WINDOW")
	out, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("xprop failed: %w", err)
	}

	fields := strings.Fields(string(out))
	if len(fields) < 5 {
		return 0, fmt.Errorf("unexpected xprop output: %s", string(out))
	}
	windowID := fields[4]

	cmd = exec.Command("xprop", "-id", windowID, "_NET_WM_PID")
	out, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("xprop pid failed: %w", err)
	}

	parts := strings.Split(string(out), "=")
	if len(parts) < 2 {
		return 0, fmt.Errorf("no pid in xprop output")
	}
	pidStr := strings.TrimSpace(parts[len(parts)-1])
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid pid: %s", pidStr)
	}
	return pid, nil
}

func (ft *ForegroundTracker) getProcessNameLinux(pid int) string {
	// Read /proc/<pid>/comm for process name
	comm, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err == nil {
		return strings.TrimSpace(string(comm))
	}
	// Fallback: /proc/<pid>/cmdline
	cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return ""
	}
	parts := strings.SplitN(string(cmdline), "\x00", 2)
	if len(parts) > 0 {
		return filepath.Base(parts[0])
	}
	return ""
}

func (ft *ForegroundTracker) getWindowTitleLinux() string {
	// Try xdotool first
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	// Fallback: xprop
	cmd = exec.Command("xprop", "-root", "_NET_ACTIVE_WINDOW")
	out, err = cmd.Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	if len(fields) < 5 {
		return ""
	}
	windowID := fields[4]

	cmd = exec.Command("xprop", "-id", windowID, "WM_NAME")
	out, err = cmd.Output()
	if err != nil {
		return ""
	}
	// xprop returns: WM_NAME(STRING) = "title"
	parts := strings.SplitN(string(out), "=", 2)
	if len(parts) < 2 {
		return ""
	}
	title := strings.TrimSpace(parts[1])
	title = strings.Trim(title, `"`)
	return title
}

// =============================================================================
// macOS — using osascript (same approach as Velosi tracker.rs)
// =============================================================================

func (ft *ForegroundTracker) pollMacOS() (*ForegroundActivity, error) {
	// Get frontmost app name + bundle ID
	appScript := `
tell application "System Events"
    set procs to (application processes whose frontmost is true and visible is true)
    if (count of procs) = 0 then return ""
    set frontApp to item 1 of procs
    set appName to name of frontApp
    try
        set bundleID to bundle identifier of frontApp
    on error
        set bundleID to ""
    end try
    return appName & "|" & bundleID
end tell`

	cmd := exec.Command("osascript", "-e", appScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("osascript failed: %w", err)
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return nil, nil
	}

	parts := strings.SplitN(result, "|", 2)
	appName := parts[0]
	var bundleID string
	if len(parts) > 1 {
		bundleID = parts[1]
	}

	windowTitle := ft.getWindowTitleMacOS()
	execName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))

	activity := &ForegroundActivity{
		AppName:     appName,
		ExecName:    execName,
		WindowTitle: windowTitle,
		PID:         0,
		Timestamp:   time.Now(),
	}

	// Try to get the app's URL if it's a browser (like Velosi does)
	if bundleID != "" {
		url := ft.getBrowserURLMacOS(appName, bundleID)
		if url != "" {
			activity.WindowTitle = url
		}
	}

	return activity, nil
}

func (ft *ForegroundTracker) getWindowTitleMacOS() string {
	script := `
tell application "System Events"
    set procs to (application processes whose frontmost is true and visible is true)
    if (count of procs) = 0 then return ""
    set frontApp to item 1 of procs
    try
        if (count of windows of frontApp) > 0 then
            set windowTitle to name of first window of frontApp
            if windowTitle is not missing value and windowTitle is not "" then
                return windowTitle
            end if
        end if
    end try
    return ""
end tell`

	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (ft *ForegroundTracker) getBrowserURLMacOS(appName, bundleID string) string {
	if strings.Contains(appName, "Chrome") || strings.Contains(bundleID, "com.google.Chrome") {
		script := `
tell application "Google Chrome"
    try
        if (count of windows) > 0 then
            set currentTab to active tab of first window
            return URL of currentTab
        end if
    end try
    return ""
end tell`
		cmd := exec.Command("osascript", "-e", script)
		out, _ := cmd.Output()
		return strings.TrimSpace(string(out))
	}

	if strings.Contains(appName, "Safari") || strings.Contains(bundleID, "com.apple.Safari") {
		script := `
tell application "Safari"
    try
        if (count of windows) > 0 then
            set currentTab to current tab of first window
            return URL of currentTab
        end if
    end try
    return ""
end tell`
		cmd := exec.Command("osascript", "-e", script)
		out, _ := cmd.Output()
		return strings.TrimSpace(string(out))
	}

	if strings.Contains(appName, "Firefox") || strings.Contains(bundleID, "org.mozilla.firefox") {
		return "Firefox"
	}

	return ""
}

// =============================================================================
// Windows — using WinAPI via syscall (same approach as Velosi tracker.rs)
// =============================================================================

func (ft *ForegroundTracker) pollWindows() (*ForegroundActivity, error) {
	// Use the syscall approach via a helper that calls the native Windows APIs
	activity := ft.pollWindowsWinAPI()
	return activity, nil
}

func (ft *ForegroundTracker) pollWindowsWinAPI() *ForegroundActivity {
	execName, windowTitle, pid, err := getWindowsForegroundInfo()
	if err != nil || execName == "" {
		return nil
	}

	return &ForegroundActivity{
		AppName:     ft.friendlyAppName(execName),
		ExecName:    strings.ToLower(strings.TrimSuffix(execName, ".exe")),
		WindowTitle: windowTitle,
		PID:         pid,
		Timestamp:   time.Now(),
	}
}

// =============================================================================
// Utilities
// =============================================================================

func (ft *ForegroundTracker) friendlyAppName(execName string) string {
	known := map[string]string{
		"chrome":           "Google Chrome",
		"google chrome":    "Google Chrome",
		"firefox":          "Mozilla Firefox",
		"mozilla firefox":  "Mozilla Firefox",
		"msedge":           "Microsoft Edge",
		"microsoft edge":   "Microsoft Edge",
		"code":             "Visual Studio Code",
		"visual studio code": "Visual Studio Code",
		"discord":          "Discord",
		"slack":            "Slack",
		"spotify":          "Spotify",
		"teams":            "Microsoft Teams",
		"microsoft teams":  "Microsoft Teams",
		"zoom":             "Zoom",
	}
	name := strings.ToLower(execName)
	if friendly, ok := known[name]; ok {
		return friendly
	}
	return execName
}

func (ft *ForegroundTracker) ListRunningProcesses() ([]InstalledApp, error) {
	switch runtime.GOOS {
	case "linux":
		return listLinuxProcesses()
	case "darwin":
		return listMacOSProcesses()
	case "windows":
		return listWindowsProcesses()
	default:
		return nil, nil
	}
}

func listLinuxProcesses() ([]InstalledApp, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var apps []InstalledApp

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid < 100 {
			continue
		}

		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		apps = append(apps, InstalledApp{
			Name: name,
			Exec: name,
		})
	}
	return apps, nil
}

func listMacOSProcesses() ([]InstalledApp, error) {
	cmd := exec.Command("ps", "-eo", "comm")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var apps []InstalledApp

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" || strings.HasPrefix(name, "/") && !strings.Contains(name, ".app") {
			continue
		}
		base := filepath.Base(name)
		if seen[base] {
			continue
		}
		seen[base] = true
		apps = append(apps, InstalledApp{
			Name: base,
			Exec: strings.ToLower(base),
		})
	}
	return apps, nil
}

func listWindowsProcesses() ([]InstalledApp, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var apps []InstalledApp

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// CSV format: "chrome.exe","1234","Console","1","123,456 K"
		parts := strings.Split(line, ",")
		if len(parts) < 1 {
			continue
		}
		name := strings.Trim(parts[0], `"`)
		execName := strings.ToLower(strings.TrimSuffix(name, ".exe"))
		if execName == "" || seen[execName] {
			continue
		}
		seen[execName] = true
		apps = append(apps, InstalledApp{
			Name: name,
			Exec: execName,
		})
	}
	return apps, nil
}
