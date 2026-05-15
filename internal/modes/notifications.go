package modes

import (
	"log"
	"os/exec"
	"runtime"
)

func MuteNotifications() {
	switch runtime.GOOS {
	case "linux":
		muteNotificationsLinux()
	case "windows":
		muteNotificationsWindows()
	case "darwin":
		muteNotificationsMacOS()
	}
}

func RestoreNotifications() {
	switch runtime.GOOS {
	case "linux":
		restoreNotificationsLinux()
	case "windows":
		restoreNotificationsWindows()
	case "darwin":
		restoreNotificationsMacOS()
	}
}

// Linux: D-Bus Inhibit (existing approach)
func muteNotificationsLinux() {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibit", "",
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] D-Bus inhibit failed: %v", err)
	}
	log.Println("[notifications] muted (D-Bus Inhibit)")
}

func restoreNotificationsLinux() {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Uninhibit", "",
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] D-Bus uninhibit failed: %v", err)
	}
	log.Println("[notifications] restored (D-Bus Uninhibit)")
}

// Windows: Registry Focus Assist approach
func muteNotificationsWindows() {
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 0 -Type DWord -Force`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast disable failed: %v", err)
	} else {
		log.Println("[notifications] Windows toasts disabled via registry")
	}

	// Also try Focus Assist (quiet hours) via registry
	quietHoursCmd := `New-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications" -Name "QuietHours" -Force | Out-Null; Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 1 -Type DWord -Force`
	cmd = exec.Command("powershell", "-NoProfile", "-Command", quietHoursCmd)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist enable failed: %v", err)
	} else {
		log.Println("[notifications] Windows Focus Assist enabled")
	}
}

func restoreNotificationsWindows() {
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast restore failed: %v", err)
	} else {
		log.Println("[notifications] Windows toasts restored via registry")
	}

	quietHoursCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 0 -Type DWord -Force`
	cmd = exec.Command("powershell", "-NoProfile", "-Command", quietHoursCmd)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist disable failed: %v", err)
	} else {
		log.Println("[notifications] Windows Focus Assist disabled")
	}
}

// macOS: via osascript notification mute
func muteNotificationsMacOS() {
	// Use AppleScript to temporarily disable notifications
	// macOS doesn't have a simple mute — this approach uses Do Not Disturb
	script := `
tell application "System Events"
    tell expose preferences
        set temp to do not disturb
        set do not disturb to true
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] macOS DND enable failed: %v", err)
	} else {
		log.Println("[notifications] macOS Do Not Disturb enabled")
	}
}

func restoreNotificationsMacOS() {
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to false
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] macOS DND disable failed: %v", err)
	} else {
		log.Println("[notifications] macOS Do Not Disturb disabled")
	}
}
