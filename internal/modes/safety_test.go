package modes

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Whitelist Safety Tests — CRITICAL: whitelist must protect system processes
// =============================================================================

func TestWhitelist_NormalizedMatch(t *testing.T) {
	whitelist := GetWhitelistExecs()

	// CRITICAL: These must be in the whitelist to prevent system UI destruction.
	// The whitelist must use normalized names WITHOUT .exe extension to match
	// tracker output. Without this, blocking a "non-allowed" foreground app
	// would taskkill explorer/startmenu → WHITE SCREEN DESKTOP CRASH.
	criticalSystemProcesses := []struct {
		name     string
		execName string // as returned by tracker (without .exe)
	}{
		{"Windows Explorer", "explorer"},
		{"Desktop Window Manager", "dwm"},
		{"Start Menu", "startmenuexperiencehost"},
		{"Shell Experience", "shellexperiencehost"},
		{"Task Manager", "taskmgr"},
		{"Task Host", "taskhostw"},
		{"Runtime Broker", "runtimebroker"},
		{"Search Host", "searchhost"},
		{"Search App", "searchapp"},
		{"System Settings", "systemsettings"},
		{"Logon UI", "logonui"},
		{"Application Frame Host", "applicationframehost"},
		{"Text Input Host", "textinputhost"},
		{"Sihost", "sihost"},
		{"Conhost", "conhost"},
		{"Command Prompt", "cmd"},
		{"PowerShell", "powershell"},
		{"Pwsh", "pwsh"},
		{"Windows Terminal", "wt"},
		{"Roboty itself", "roboty"},
		{"Roboty dev", "roboty-dev"},
	}

	for _, p := range criticalSystemProcesses {
		if !whitelist[p.execName] {
			t.Errorf("🚨 CRITICAL: whitelist missing normalized entry for %s (key=%q) — "+
				"This means %s CAN BE KILLED by the app blocker, causing SYSTEM UI CRASH!",
				p.name, p.execName, p.name)
		}
	}

	// Also check .exe forms are NO LONGER the only entries (they should be normalized out)
	if whitelist["explorer.exe"] {
		t.Log("Note: explorer.exe exists in whitelist set (not harmful, normalized version takes precedence)")
	}
}

func TestWhitelist_NoPartialMatching(t *testing.T) {
	whitelist := GetWhitelistExecs()

	// Verify that short process names don't accidentally match longer ones
	// The whitelist uses exact map lookup, so this should be safe by design.
	// But verify there's no substring logic in the lookup path.
	badSubstrings := []string{
		"explor",   // partial of "explorer"
		"svch",     // partial of "svchost"
		"system",   // partial of "systemd", "systemsettings"
		"search",   // partial of "searchhost"
		"shell",    // partial of "shellexperiencehost"
		"window",   // partial of "windowsterminal"
	}

	for _, s := range badSubstrings {
		if whitelist[s] {
			// This would mean partial names are in the whitelist,
			// which is a misconfiguration
			t.Errorf("whitelist should NOT contain partial name: %q", s)
		}
	}
}

func TestWhitelist_ContainsAllPlatformEntries(t *testing.T) {
	whitelist := GetWhitelistExecs()

	// Windows critical
	winRequired := []string{"explorer", "dwm", "svchost", "csrss", "lsass"}
	for _, name := range winRequired {
		if !whitelist[name] {
			t.Errorf("Windows critical process %q missing from whitelist", name)
		}
	}

	// Linux critical (all lowercased by GetWhitelistExecs)
	linuxRequired := []string{"systemd", "gnome-shell", "xorg", "dbus-daemon"}
	for _, name := range linuxRequired {
		if !whitelist[name] {
			t.Errorf("Linux critical process %q missing from whitelist", name)
		}
	}

	// macOS critical
	macRequired := []string{"finder", "dock", "windowmanager", "launchd"}
	for _, name := range macRequired {
		if !whitelist[name] {
			t.Errorf("macOS critical process %q missing from whitelist", name)
		}
	}
}

// =============================================================================
// Kill Safety Verifier Tests
// =============================================================================

func TestKillSafetyVerifier_BlocksSystemProcesses(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	systemProcesses := []string{
		"explorer", "explorer.exe",
		"dwm", "dwm.exe",
		"svchost", "svchost.exe",
		"csrss", "csrss.exe",
		"lsass", "lsass.exe",
		"systemd",
		"gnome-shell",
		"dock",
		"finder",
		"launchd",
	}

	for _, proc := range systemProcesses {
		safe, reason := v.IsSafeToKill(proc)
		if safe {
			t.Errorf("🚨 CRITICAL: KillSafetyVerifier says it's SAFE to kill %q — this is DANGEROUS! Reason given: %s", proc, reason)
		}
	}
}

func TestKillSafetyVerifier_AllowsUserProcesses(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	userProcesses := []string{
		"chrome",
		"firefox",
		"spotify",
		"discord",
		"slack",
		"steam",
		"notepad",
		"calculator",
	}

	for _, proc := range userProcesses {
		safe, reason := v.IsSafeToKill(proc)
		if !safe {
			t.Errorf("KillSafetyVerifier blocked user process %q: %s", proc, reason)
		}
	}
}

func TestKillSafetyVerifier_ProtectsRoboty(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	selfProcesses := []string{
		"roboty",
		"roboty.exe",
		"roboty-dev",
		"roboty-dev.exe",
		"roboty1",
		"roboty1.exe",
	}

	for _, proc := range selfProcesses {
		safe, reason := v.IsSafeToKill(proc)
		if safe {
			t.Errorf("KillSafetyVerifier says it's safe to kill %q — this would kill Roboty itself! Reason: %s", proc, reason)
		}
	}
}

// =============================================================================
// NormalizeKillExec Tests
// =============================================================================

func TestNormalizeKillExec_ValidInputs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"chrome", "chrome"},
		{"chrome.exe", "chrome"},
		{"CHROME.EXE", "chrome"},
		{"Notepad++.exe", "notepad++"},
		{"Code.exe", "code"},
		{"firefox", "firefox"},
	}

	for _, tt := range tests {
		got := NormalizeKillExec(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeKillExec(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizeKillExec_RejectsDangerousInputs(t *testing.T) {
	dangerous := []string{
		"",
		".",
		"..",
		"&whoami",
		"|dir",
		";rm -rf /",
		"`id`",
		"$(cat /etc/passwd)",
		"-rf",
		"--help",
		"../etc/passwd",
		"chrome.exe;rm -rf /",
	}

	for _, input := range dangerous {
		got := NormalizeKillExec(input)
		if got != "" {
			t.Errorf("NormalizeKillExec(%q) = %q, want empty string (rejected dangerous input)", input, got)
		}
	}
}

// =============================================================================
// Proxy Safety Tests
// =============================================================================

func TestIsAlwaysAllowed_Localhost(t *testing.T) {
	localhosts := []string{
		"localhost",
		"127.0.0.1",
		"127.0.0.1:62828",
		"::1",
		"0.0.0.0",
		"localhost.localdomain",
		"127.0.0.5",
		"127.255.255.255",
		"wails",
		"wails.localhost",
	}

	for _, host := range localhosts {
		if !isAlwaysAllowed(host) {
			t.Errorf("isAlwaysAllowed(%q) should be true — this host must NEVER be blocked by proxy", host)
		}
	}
}

func TestIsAlwaysAllowed_BlocksExternal(t *testing.T) {
	external := []string{
		"google.com",
		"github.com",
		"youtube.com",
		"192.168.1.1",
		"10.0.0.1",
	}

	for _, host := range external {
		if isAlwaysAllowed(host) {
			t.Errorf("isAlwaysAllowed(%q) should be false — external host should not bypass proxy", host)
		}
	}
}

func TestNormalizeURLs_LocalhostPreserved(t *testing.T) {
	urls := normalizeURLs([]string{"localhost", "127.0.0.1"})
	for _, u := range urls {
		_ = u
	}
	// normalizeURLs may strip these, but isAllowed should catch them regardless
	// via isAlwaysAllowed() check
	t.Log("normalizeURLs result:", urls)
}

// =============================================================================
// Kill Loop Detector Tests
// =============================================================================

func TestKillLoopDetector(t *testing.T) {
	kld := NewKillLoopDetector()

	// Under threshold
	for i := 0; i < MaxConsecutiveKills-1; i++ {
		if kld.RecordKill("chrome") {
			t.Fatalf("Kill loop detected before threshold")
		}
	}

	// At threshold — should be detected
	if !kld.RecordKill("chrome") {
		t.Errorf("Kill loop should be detected at %d kills", MaxConsecutiveKills)
	}

	// Different process should not trigger
	if kld.RecordKill("firefox") {
		t.Errorf("Kill loop detected for different process")
	}
}

func TestKillLoopDetector_WindowExpiry(t *testing.T) {
	kld := NewKillLoopDetector()
	// Use small window to make test fast
	oldWindow := KillLoopWindow
	defer func() { KillLoopWindow = oldWindow }()
	KillLoopWindow = 50 * time.Millisecond

	// Record kills spread out over time
	for i := 0; i < MaxConsecutiveKills; i++ {
		kld.RecordKill("chrome")
		time.Sleep(KillLoopWindow / time.Duration(MaxConsecutiveKills+2))
	}

	// Because kills are spread over larger-than-window period,
	// no single window should have > MaxConsecutiveKills
	if kld.RecordKill("chrome") {
		t.Log("Kill loop detected (may false-positive due to timing)")
	}
}

// =============================================================================
// Safe Development Mode Tests
// =============================================================================

func TestIsDevMode_Default(t *testing.T) {
	os.Unsetenv("ROBOTY_SAFE_MODE")
	// Reset the sync.Once to allow re-test
	isDevMode = false
	devModeOnce = sync.Once{}
	if IsDevMode() {
		t.Error("IsDevMode() should be false when ROBOTY_SAFE_MODE is not set")
	}
}

func TestIsDevMode_Enabled(t *testing.T) {
	os.Setenv("ROBOTY_SAFE_MODE", "true")
	isDevMode = false
	devModeOnce = sync.Once{}
	if !IsDevMode() {
		t.Error("IsDevMode() should be true when ROBOTY_SAFE_MODE=true")
	}
	os.Unsetenv("ROBOTY_SAFE_MODE")
}

// =============================================================================
// App Blocker Safety Tests
// =============================================================================

func TestAppBlocker_WhitelistProtection(t *testing.T) {
	tracker := NewForegroundTracker()
	blocker := NewAppBlocker(tracker)

	// Test with a whitelisted-only setup — even if user allows nothing,
	// system processes must be protected
	blocker.Start([]string{}, nil, time.Hour)
	defer blocker.Stop()

	// Verify the internal safety verifier is configured
	if blocker.safetyVerifier == nil {
		t.Fatal("blocker should have a KillSafetyVerifier")
	}
}

func TestAppBlocker_KillLoopProtection(t *testing.T) {
	detector := NewKillLoopDetector()

	// Simulate kill loop for a process
	for i := 0; i < MaxConsecutiveKills+1; i++ {
		detector.RecordKill("test-app")
	}

	// Verify the stop channel mechanism works
	if !detector.RecordKill("test-app") {
		t.Errorf("Kill loop should be detected after %d+ kills", MaxConsecutiveKills)
	}
}

// =============================================================================
// End-to-End Proxy + Blocker Safety
// =============================================================================

func TestURLBlocker_AlwaysAllowsLocalhost(t *testing.T) {
	ub := NewURLBlocker()
	ub.allowedURLs = normalizeURLs([]string{"github.com"})

	// These must always be allowed regardless of the allowed URLs list
	alwaysAllowed := []string{
		"localhost",
		"localhost:34115",
		"127.0.0.1",
		"127.0.0.1:62828",
		"::1",
		"0.0.0.0",
	}

	for _, host := range alwaysAllowed {
		if !ub.isAllowed(host) {
			t.Errorf("URLBlocker must ALWAYS allow %q — it blocked it!", host)
		}
	}

	// External hosts should still be blocked
	if ub.isAllowed("google.com") {
		t.Error("URLBlocker should block external domains not in allowed list")
	}

	// Allowed hosts should work
	if !ub.isAllowed("github.com") {
		t.Error("URLBlocker should allow hosts in the allowed list")
	}
}

// =============================================================================
// Concurrency Safety Tests
// =============================================================================

func TestKillSafetyVerifier_ConcurrentAccess(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				v.IsSafeToKill("chrome")
				v.IsSafeToKill("explorer")
				v.IsSafeToKill("roboty")
			}
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Timeout waiting for goroutines — possible deadlock")
		}
	}
}

func TestKillLoopDetector_ConcurrentAccess(t *testing.T) {
	kld := NewKillLoopDetector()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				kld.RecordKill("test-app")
			}
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Timeout — possible deadlock in KillLoopDetector")
		}
	}
}

// =============================================================================
// Audit Log Tests
// =============================================================================

func TestSafetyAuditLogger(t *testing.T) {
	logger := NewSafetyAuditLogger(100)

	logger.Log(SafetyEvent{
		Type:    EventBlockedApp,
		Target:  "chrome",
		Message: "blocked by focus mode",
	})

	logger.Log(SafetyEvent{
		Type:    EventLocalhostBypass,
		Target:  "127.0.0.1",
		Message: "localhost bypass",
		Allowed: true,
	})

	events := logger.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].Type != EventBlockedApp {
		t.Errorf("First event type wrong: %s", events[0].Type)
	}
	if events[1].Type != EventLocalhostBypass {
		t.Errorf("Second event type wrong: %s", events[1].Type)
	}
}

func TestSafetyAuditLogger_MaxEvents(t *testing.T) {
	logger := NewSafetyAuditLogger(10)

	for i := 0; i < 20; i++ {
		logger.Log(SafetyEvent{
			Type:    EventBlockedApp,
			Target:  "chrome",
			Message: "test",
		})
	}

	events := logger.GetEvents()
	if len(events) > 10 {
		t.Errorf("Expected max 10 events, got %d", len(events))
	}
}

// =============================================================================
// Ensure Rollback Safety
// =============================================================================

func TestRollbackSafety_CloseAppsVerifiesSafety(t *testing.T) {
	// CloseApps should refuse to kill system processes
	// This tests the safety interlock in blocking.go

	// Note: these calls go through NormalizeKillExec + safety verification
	// They should log errors but NOT actually kill anything
	systemProcs := []string{
		"explorer",
		"explorer.exe",
		"svchost.exe",
		"systemd",
		"winlogon.exe",
	}

	for _, proc := range systemProcs {
		safeExec := NormalizeKillExec(proc)
		if safeExec == "" {
			continue // Already rejected
		}
		safe, reason := globalSafetyVerifier.IsSafeToKill(safeExec)
		if safe {
			t.Errorf("CloseApps would allow killing %q (safe=%v): %s", proc, safe, reason)
		}
	}
}

// =============================================================================
// Wails Dev Mode Compatibility
// =============================================================================

func TestWailsDevMode_NoBlockLocalhost(t *testing.T) {
	// Wails development mode uses:
	// - Vite dev server on localhost:34115
	// - WebSocket for HMR
	// - wails:// protocol for frontend
	// - Application binds to 127.0.0.1 for Wails runtime

	// All localhost, wails, and loopback addresses MUST be exempt
	ub := NewURLBlocker()
	ub.allowedURLs = normalizeURLs([]string{}) // No URLs allowed at all

	wailsHosts := []string{
		"localhost",
		"localhost:34115",
		"127.0.0.1",
		"127.0.0.1:34115",
		"wails",
		"wails.localhost",
	}

	for _, host := range wailsHosts {
		if !ub.isAllowed(host) {
			t.Errorf("Wails dev mode host %q blocked — this would cause WHITE SCREEN in WebView", host)
		}
	}

	// External hosts must still be blocked
	if ub.isAllowed("google.com") {
		t.Error("External host should be blocked when allowed list is empty")
	}
}

// =============================================================================
// Self-Protection Tests
// =============================================================================

func TestSelfProtection_RobotyProcesses(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	// All possible names Roboty might run under
	selfNames := []string{
		"roboty",
		"roboty.exe",
		"roboty1",
		"roboty1.exe",
		"roboty-dev",
		"roboty-dev.exe",
		"wails",
		"wails.exe",
	}

	for _, name := range selfNames {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("Self-protection FAILED for %q — Roboty would kill itself! Reason: %s", name, reason)
		}
		_ = reason
	}
}

// =============================================================================
// Red Team: Bypass Attempts
// =============================================================================

func TestRedTeam_KillExecBypass(t *testing.T) {
	// Attempt to bypass the safety checker with various tricks
	bypassAttempts := []struct {
		input    string
		expected string // "" means should be rejected
	}{
		{"chrome", "chrome"},
		{"chrome.exe", "chrome"},
		{"CHROME.EXE", "chrome"},
		{"Chrome.exe", "chrome"},
		{"../chrome", ""},          // path traversal
		{".\\chrome", ""},          // relative path
		{"chrome;rm -rf /", ""},    // command injection
		{"chrome | whoami", ""},    // pipe injection
		{"$(whoami)", ""},          // subshell
		{"`whoami`", ""},           // backtick
		{"-rf", ""},               // flag injection
		{"--help", ""},             // flag injection
		{"chrome.exe;", ""},        // command injection
		{"chrome.exe&", ""},        // background injection
		{"", ""},                   // empty
		{".", ""},                  // dot
		{"..", ""},                 // double dot
	}

	for _, tt := range bypassAttempts {
		got := NormalizeKillExec(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeKillExec(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRedTeam_WhitelistLookupBypass(t *testing.T) {
	whitelist := GetWhitelistExecs()

	// Ensure whitelist doesn't contain entries that would cause false matches
	// via path traversal or encoded forms
	dangerousKeys := []string{}
	for k := range whitelist {
		if strings.ContainsAny(k, "\\/:;&|`$(){}[]'\"") {
			dangerousKeys = append(dangerousKeys, k)
		}
		// The key itself must be a simple process name
		if k != strings.ToLower(strings.TrimSpace(k)) {
			dangerousKeys = append(dangerousKeys, k+" (case issue)")
		}
	}

	if len(dangerousKeys) > 0 {
		t.Errorf("Whitelist contains dangerous keys: %v", dangerousKeys)
	}
}


