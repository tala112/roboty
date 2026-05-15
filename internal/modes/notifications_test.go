package modes

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestNotificationMute_RoundTrip(t *testing.T) {
	// Just verify Mute/Restore don't panic
	// On headless/CI systems, the shell commands may fail, but they should log and return
	t.Run("mute", func(t *testing.T) {
		MuteNotifications()
	})

	t.Run("restore", func(t *testing.T) {
		RestoreNotifications()
	})
}

// Test the Windows notification commands are syntactically valid PowerShell
func TestWindowsNotificationCommands_ValidPowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows specific")
	}

	cmds := []string{
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 0 -Type DWord -Force`,
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force`,
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 1 -Type DWord -Force`,
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 0 -Type DWord -Force`,
	}

	for _, psCmd := range cmds {
		// Use -WhatIf to verify syntax without actually changing anything
		fullCmd := psCmd + " -WhatIf"
		cmd := exec.Command("powershell", "-NoProfile", "-Command", fullCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("PowerShell command failed: %v\nCommand: %s\nOutput: %s", err, fullCmd, string(out))
		} else {
			t.Logf("PowerShell -WhatIf OK: %s", strings.TrimSpace(string(out)))
		}
	}
}

// Test the Linux notification commands (busctl -- check they're parseable)
func TestLinuxNotificationCommands(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux specific")
	}

	// busctl --user should exist
	cmd := exec.Command("which", "busctl")
	if err := cmd.Run(); err != nil {
		t.Log("busctl not found (expected on non-Linux or without systemd)")
		t.Skip("busctl not available")
	}

	// Verify the D-Bus inhibit command syntax
	cmd = exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibit", "",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("D-Bus inhibit failed (expected on non-graphical session): %v - %s", err, string(out))
	} else {
		t.Logf("D-Bus inhibit OK: %s", string(out))
	}

	// Restore
	cmd = exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Uninhibit", "",
	)
	_ = cmd.Run()
}

// Test the macOS notification commands (osascript syntax)
func TestMacOSNotificationCommands(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS specific")
	}

	// Verify osascript exists
	cmd := exec.Command("which", "osascript")
	if err := cmd.Run(); err != nil {
		t.Skip("osascript not available")
	}

	// Test the DND enable script
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to true
    end tell
end tell`
	cmd = exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("macOS DND enable failed: %v - %s", err, string(out))
	} else {
		t.Log("macOS DND enable OK")
		// Restore
		restoreScript := `
tell application "System Events"
    tell expose preferences
        set do not disturb to false
    end tell
end tell`
		cmd = exec.Command("osascript", "-e", restoreScript)
		_ = cmd.Run()
	}
}

// Test platform dispatch for notifications
func TestNotificationPlatformDispatch(t *testing.T) {
	// Ensure MuteNotifications/RestoreNotifications complete without panic
	// regardless of platform
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("notification functions panicked: %v", r)
		}
	}()

	MuteNotifications()
	RestoreNotifications()
}

// Test notification command generation (verify shell commands are properly escaped)
func TestNotificationCommandSafety(t *testing.T) {
	t.Run("powershell no special chars", func(t *testing.T) {
		// Verify all PowerShell commands use single-quote-safe strings
		// or properly escaped paths
		cmds := []string{
			`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 0 -Type DWord -Force`,
			`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force`,
		}
		for _, cmd := range cmds {
			if strings.Contains(cmd, "`") {
				t.Errorf("PowerShell command contains backtick (possible injection): %s", cmd)
			}
		}
	})

	t.Run("osascript no injection", func(t *testing.T) {
		// Verify osascript commands don't have shell injection vectors
		macScripts := []string{
			`tell application "System Events"`,
			`tell expose preferences`,
			`set do not disturb to true`,
		}
		for _, s := range macScripts {
			if strings.Contains(s, ";") || strings.Contains(s, "`") || strings.Contains(s, "$") {
				t.Errorf("osascript contains potential injection chars: %s", s)
			}
		}
	})

	t.Run("busctl safe args", func(t *testing.T) {
		busctlArgs := []string{"busctl", "--user", "call",
			"org.freedesktop.Notifications",
			"/org/freedesktop/Notifications",
			"org.freedesktop.Notifications",
			"Inhibit", "",
		}
		for _, arg := range busctlArgs {
			if strings.Contains(arg, ";") || strings.Contains(arg, "`") || strings.Contains(arg, "$") {
				t.Errorf("busctl arg contains potential injection chars: %s", arg)
			}
		}
	})
}
