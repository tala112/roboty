package modes

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// realNotificationManager is the production implementation.
type realNotificationManager struct{}

func NewRealNotificationManager() NotificationManager {
	return &realNotificationManager{}
}

func (m *realNotificationManager) Mute() error {
	if IsDevMode() {
		log.Println("[dev] WOULD mute notifications")
		return nil
	}
	switch runtime.GOOS {
	case "linux":
		return m.muteLinux()
	case "windows":
		return m.muteWindows()
	case "darwin":
		return m.muteMacOS()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (m *realNotificationManager) Restore() error {
	if IsDevMode() {
		log.Println("[dev] WOULD restore notifications")
		return nil
	}
	switch runtime.GOOS {
	case "linux":
		return m.restoreLinux()
	case "windows":
		return m.restoreWindows()
	case "darwin":
		return m.restoreMacOS()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (m *realNotificationManager) IsMuted() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		return m.isMutedWindows()
	case "linux":
		return m.isMutedLinux()
	case "darwin":
		return m.isMutedMacOS()
	}
	return false, nil
}

func (m *realNotificationManager) muteLinux() error {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibit", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("D-Bus inhibit: %w", err)
	}
	log.Println("[notifications] muted (D-Bus Inhibit)")
	return nil
}

func (m *realNotificationManager) restoreLinux() error {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Uninhibit", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("D-Bus uninhibit: %w", err)
	}
	log.Println("[notifications] restored (D-Bus Uninhibit)")
	return nil
}

func (m *realNotificationManager) isMutedLinux() (bool, error) {
	out, err := exec.Command("busctl", "--user", "get-property",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibited").Output()
	if err != nil {
		return false, err
	}
	return string(out) == "true\n", nil
}

func (m *realNotificationManager) muteWindows() error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 0 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast disable failed: %v", err)
	}

	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		`New-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications" -Name "QuietHours" -Force | Out-Null; Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 1 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist enable failed: %v", err)
	}
	log.Println("[notifications] Windows notifications muted")
	return nil
}

func (m *realNotificationManager) restoreWindows() error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast restore failed: %v", err)
	}

	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 0 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist disable failed: %v", err)
	}
	log.Println("[notifications] Windows notifications restored")
	return nil
}

func (m *realNotificationManager) isMutedWindows() (bool, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED").NOC_GLOBAL_SETTING_TOASTS_ENABLED`)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "0", nil
}

func (m *realNotificationManager) muteMacOS() error {
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to true
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("macOS DND enable: %w", err)
	}
	log.Println("[notifications] macOS Do Not Disturb enabled")
	return nil
}

func (m *realNotificationManager) restoreMacOS() error {
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to false
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("macOS DND disable: %w", err)
	}
	log.Println("[notifications] macOS Do Not Disturb disabled")
	return nil
}

func (m *realNotificationManager) isMutedMacOS() (bool, error) {
	script := `
tell application "System Events"
    tell expose preferences
        get do not disturb
    end tell
end tell`
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "true", nil
}
