//go:build darwin

package darwin

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"Roboty/internal/platform/types"
)

func (p *DarwinPlatform) Poll() (*types.ForegroundActivity, error) {
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

	windowTitle := getWindowTitleMacOS()
	execName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))

	activity := &types.ForegroundActivity{
		AppName:     appName,
		ExecName:    execName,
		WindowTitle: windowTitle,
		PID:         0,
		Timestamp:   time.Now(),
	}

	if bundleID != "" {
		url := getBrowserURLMacOS(appName, bundleID)
		if url != "" {
			activity.WindowTitle = url
		}
	}

	return activity, nil
}

func getWindowTitleMacOS() string {
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

func getBrowserURLMacOS(appName, bundleID string) string {
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
