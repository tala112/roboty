//go:build linux

package linux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Roboty/internal/platform/types"
)

func (p *LinuxPlatform) Poll() (*types.ForegroundActivity, error) {
	pid, err := getActivePIDLinux()
	if err != nil || pid == 0 {
		return nil, err
	}

	execName := getProcessNameLinux(pid)
	windowTitle := getWindowTitleLinux()

	if execName == "" {
		return nil, nil
	}

	return &types.ForegroundActivity{
		AppName:     friendlyAppNameLinux(execName),
		ExecName:    execName,
		WindowTitle: windowTitle,
		PID:         pid,
		Timestamp:   time.Now(),
	}, nil
}

func getActivePIDLinux() (int, error) {
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowpid")
	out, err := cmd.Output()
	if err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil && pid > 0 {
			return pid, nil
		}
	}

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

func getProcessNameLinux(pid int) string {
	comm, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err == nil {
		return strings.TrimSpace(string(comm))
	}

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

func getWindowTitleLinux() string {
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

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
	parts := strings.SplitN(string(out), "=", 2)
	if len(parts) < 2 {
		return ""
	}
	title := strings.TrimSpace(parts[1])
	title = strings.Trim(title, `"`)
	return title
}

func friendlyAppNameLinux(execName string) string {
	known := map[string]string{
		"chrome":              "Google Chrome",
		"google chrome":       "Google Chrome",
		"firefox":             "Mozilla Firefox",
		"mozilla firefox":     "Mozilla Firefox",
		"msedge":              "Microsoft Edge",
		"microsoft edge":      "Microsoft Edge",
		"code":                "Visual Studio Code",
		"visual studio code":  "Visual Studio Code",
		"discord":             "Discord",
		"slack":               "Slack",
		"spotify":             "Spotify",
		"teams":               "Microsoft Teams",
		"microsoft teams":     "Microsoft Teams",
		"zoom":                "Zoom",
		"firefox-esr":         "Firefox ESR",
	}
	name := strings.ToLower(execName)
	if friendly, ok := known[name]; ok {
		return friendly
	}
	return execName
}
