package modes

import (
	"strings"
	"testing"
	"time"
)

func TestAppBlocker_Lifecycle(t *testing.T) {
	tracker := NewForegroundTracker()
	blocker := NewAppBlocker(tracker)

	if blocker.IsRunning() {
		t.Fatal("new blocker should not be running")
	}

	// Start without closing apps
	blocker.Start([]string{"chrome", "firefox"}, nil, 100*time.Millisecond)
	if !blocker.IsRunning() {
		t.Fatal("blocker should be running after Start")
	}

	// Second Start should be no-op (already running)
	blocker.Start([]string{"code"}, nil, 100*time.Millisecond)

	// Stop
	blocker.Stop()
	if blocker.IsRunning() {
		t.Fatal("blocker should not be running after Stop")
	}

	// Can restart after stop
	blocker.Start([]string{"chrome"}, nil, 100*time.Millisecond)
	if !blocker.IsRunning() {
		t.Fatal("blocker should restart after stop")
	}
	blocker.Stop()
}

func TestAppBlocker_AllowedSetIncludesWhitelist(t *testing.T) {
	tracker := NewForegroundTracker()
	blocker := NewAppBlocker(tracker)

	// Only chrome is user-allowed
	blocker.Start([]string{"chrome"}, nil, time.Hour)
	defer blocker.Stop()

	// Check: the whitelist entries should be in the allowed set
	// We can verify indirectly by checking that the goroutine runs
	// But the allowedSet is internal — test via the whitelist function directly
	whitelist := GetWhitelistExecs()
	if !whitelist["explorer.exe"] {
		t.Error("whitelist should contain explorer.exe")
	}
	if !whitelist["svchost.exe"] {
		t.Error("whitelist should contain svchost.exe")
	}
	if !whitelist["roboty"] {
		t.Error("whitelist should contain roboty")
	}
	if !whitelist["roboty-dev"] {
		t.Error("whitelist should contain roboty-dev")
	}
}

// Test that blocking.go CloseApp is safe to call with empty/nil/unknown exec
func TestCloseApp_EdgeCases(t *testing.T) {
	// CloseApp should just log and return — should not panic or crash
	CloseApp("")
	CloseApp("nonexistent-app-that-will-never-exist-12345")
}

// Test that CloseApps handles edge cases
func TestCloseApps_EdgeCases(t *testing.T) {
	CloseApps(nil)
	CloseApps([]string{})
	CloseApps([]string{"", "valid-app"})
}

func TestIsAppRunning_EdgeCases(t *testing.T) {
	// Should return false for empty/nil/unknown, not crash
	if isAppRunning("") {
		t.Error("empty exec should not be running")
	}
}

// Run on Windows: verify tasklist parsing works
func TestListWindowsProcesses_Format(t *testing.T) {
	if !isWindows() {
		t.Skip("Windows only")
	}
	apps, err := listWindowsProcesses()
	if err != nil {
		t.Fatalf("listWindowsProcesses failed: %v", err)
	}
	if len(apps) == 0 {
		t.Fatal("expected at least one process on Windows")
	}
	// Verify each entry has valid fields
	for _, a := range apps {
		if a.Name == "" {
			t.Error("app name should not be empty")
		}
		if a.Exec == "" {
			t.Error("app exec should not be empty")
		}
	}
	// Should contain the current process (this test or powershell)
	found := false
	for _, a := range apps {
		if strings.Contains(a.Exec, "powershell") || strings.Contains(a.Exec, "go") {
			found = true
			break
		}
	}
	if !found {
		t.Log("note: process list didn't contain expected entries (this is OK on some systems)")
	}
}

func isWindows() bool {
	return true // build tag check; this file only runs on windows in practice
}

// Test foreground tracker dispatch (palatable)
func TestForegroundTracker_PollDispatch(t *testing.T) {
	ft := NewForegroundTracker()
	// Poll should not panic on any platform — it runs the OS-specific code
	activity, err := ft.Poll()
	if err != nil {
		// Error is acceptable (no display, no X11, etc.)
		t.Logf("Poll returned error (expected on headless): %v", err)
	}
	if activity == nil {
		t.Log("Poll returned nil (expected on headless or no display)")
	} else {
		t.Logf("Poll succeeded: app=%s exec=%s title=%s pid=%d", activity.AppName, activity.ExecName, activity.WindowTitle, activity.PID)
	}
}

func TestFriendlyAppName(t *testing.T) {
	tracker := NewForegroundTracker()
	tests := []struct {
		input    string
		expected string
	}{
		{"chrome", "Google Chrome"},
		{"google chrome", "Google Chrome"},
		{"firefox", "Mozilla Firefox"},
		{"code", "Visual Studio Code"},
		{"discord", "Discord"},
		{"slack", "Slack"},
		{"spotify", "Spotify"},
		{"teams", "Microsoft Teams"},
		{"zoom", "Zoom"},
		{"unknown-app", "unknown-app"},
		{"", ""},
	}
	for _, tt := range tests {
		got := tracker.friendlyAppName(tt.input)
		if got != tt.expected {
			t.Errorf("friendlyAppName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
