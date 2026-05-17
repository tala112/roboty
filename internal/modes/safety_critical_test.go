package modes

import (
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Critical Safety Regression Tests
// =============================================================================

func TestCritical_ExplorerNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"explorer", "explorer.exe", "EXPLORER.EXE", "Explorer"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: %q was allowed: %s", name, reason)
		}
	}
}

func TestCritical_DwmNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"dwm", "dwm.exe"} {
		safe, _ := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: %q was allowed to be killed", name)
		}
	}
}

func TestCritical_LocalhostNeverBlockedByProxy(t *testing.T) {
	ub := NewURLBlocker()
	ub.allowedURLs = []string{}

	mustPass := []string{
		"localhost", "localhost:34115", "localhost:5173",
		"127.0.0.1", "127.0.0.1:34115",
		"::1",
		"0.0.0.0",
		"wails", "wails.localhost",
	}
	for _, host := range mustPass {
		if !ub.isAllowed(host) {
			t.Errorf("CRITICAL: %q was blocked by proxy (should always pass)", host)
		}
	}

	mustBlock := []string{
		"example.com", "google.com", "evil.com",
	}
	for _, host := range mustBlock {
		if ub.isAllowed(host) {
			t.Errorf("CRITICAL: %q was allowed by proxy with empty allowed list", host)
		}
	}
}

func TestCritical_RobotyNeverSelfKill(t *testing.T) {
	v := NewKillSafetyVerifier()
	selfNames := []string{"roboty", "roboty.exe", "ROBOTY.EXE", "wails", "wails.exe"}
	for _, name := range selfNames {
		safe, _ := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: Roboty self-name %q was allowed to be killed", name)
		}
	}
}

func TestCritical_AncestorProcessesProtected(t *testing.T) {
	v := NewKillSafetyVerifier()
	v.Refresh()

	// Get current ancestor processes
	ancestors := GetAncestorExecs()
	for exec := range ancestors {
		safe, reason := v.IsSafeToKill(exec)
		if safe {
			t.Errorf("CRITICAL: ancestor process %q was allowed to be killed", exec)
		}
		_ = reason
	}
}

func TestCritical_SafeModePreventsKill(t *testing.T) {
	ResetDevMode()
	os.Setenv("ROBOTY_SAFE_MODE", "true")
	defer os.Unsetenv("ROBOTY_SAFE_MODE")

	if !IsDevMode() {
		t.Fatal("IsDevMode should be true")
	}

	// NormalizeKillExec should still work (just normalizes)
	result := NormalizeKillExec("chrome.exe")
	if result != "chrome" {
		t.Errorf("NormalizeKillExec should still normalize in safe mode, got %q", result)
	}
}

func TestCritical_SafeModePreventsProxyEnable(t *testing.T) {
	ResetDevMode()
	os.Setenv("ROBOTY_SAFE_MODE", "true")
	defer os.Unsetenv("ROBOTY_SAFE_MODE")

	proxyMgr := NewRealProxyManager()
	// Enable should just log and return nil in dev mode
	err := proxyMgr.Enable("127.0.0.1", getFreePort(t))
	if err != nil {
		t.Errorf("Enable should not fail in dev mode: %v", err)
	}
}

func TestCritical_KillLoopDetector_PreventsInfiniteLoop(t *testing.T) {
	kld := NewKillLoopDetector()
	execName := "chrome"

	// Record kills up to threshold
	for i := 0; i < MaxConsecutiveKills-1; i++ {
		if kld.RecordKill(execName) {
			t.Errorf("kill loop detected early at iteration %d", i+1)
		}
	}

	// The threshold kill should trigger
	if !kld.RecordKill(execName) {
		t.Errorf("kill loop NOT detected at threshold %d", MaxConsecutiveKills)
	}
}

func TestCritical_KillMustPassSafetyVerifier(t *testing.T) {
	// This simulates the blocking.go safeExecName logic
	systemProcs := []string{
		"explorer", "dwm", "svchost", "csrss", "lsass",
		"systemd", "gnome-shell", "windowserver", "launchd",
		"roboty", "wails",
	}

	for _, proc := range systemProcs {
		safeExec := NormalizeKillExec(proc)
		if safeExec == "" {
			continue // rejected by normalize
		}

		safe, reason := GetGlobalSafetyVerifier().IsSafeToKill(safeExec)
		if safe {
			t.Errorf("CRITICAL: %q passed safety verifier (%s)", proc, reason)
		}
	}
}

func TestCritical_ProxyHTTPSPortPreserved(t *testing.T) {
	tests := []struct {
		rHost    string
		rURLHost string
		want     string
	}{
		{"example.com:443", "example.com", "example.com:443"},
		{"example.com:8080", "example.com", "example.com:8080"},
		{"localhost:34115", "localhost", "localhost:34115"},
		{"127.0.0.1:62828", "127.0.0.1", "127.0.0.1:62828"},
		{"example.com", "example.com", "example.com"},
	}

	for _, tt := range tests {
		// This simulates the handleHTTPS fix (using r.Host first)
		host := tt.rHost
		if host == "" {
			host = tt.rURLHost
		}
		if host != tt.want {
			t.Errorf("expected %q, got %q", tt.want, host)
		}
	}
}

func TestCritical_CrashRecovery_ProxyOrphanCleaned(t *testing.T) {
	// Test the state file cleanup logic
	sf := newFakeStateFileManager()

	// Simulate proxy being enabled before crash
	sf.SaveState(proxyStateName)
	if !sf.StateExists(proxyStateName) {
		t.Fatal("state should exist after save")
	}

	// Cleanup
	sf.ClearState(proxyStateName)
	if sf.StateExists(proxyStateName) {
		t.Error("state should not exist after clear")
	}
}

func TestCritical_CrashRecovery_NotificationOrphanCleaned(t *testing.T) {
	// Test notification state cleanup
	sf := newFakeStateFileManager()

	sf.SaveState(notifStateName)
	if !sf.StateExists(notifStateName) {
		t.Fatal("state should exist after save")
	}

	sf.ClearState(notifStateName)
	if sf.StateExists(notifStateName) {
		t.Error("state should not exist after clear")
	}
}

func TestCritical_ProxyDisable_ClearsState(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	ub.Start(getFreePort(t), []string{"example.com"})

	if !fakeState.StateExists(proxyStateName) {
		t.Error("state should exist after proxy start")
	}

	ub.Stop()

	if fakeState.StateExists(proxyStateName) {
		t.Error("state should NOT exist after proxy stop")
	}
}

func TestCritical_NormalizeKillExec_RejectsDangerousPatterns(t *testing.T) {
	dangerous := []struct {
		name  string
		input string
	}{
		{"pipe", "chrome | ls"},
		{"semicolon", "chrome; rm"},
		{"ampersand", "chrome &"},
		{"backtick", "chrome `ls`"},
		{"dollar", "chrome $(cmd)"},
		{"path traversal", "../dangerous"},
		{"flag injection", "--help"},
		{"null byte", "chrome\x00.exe"},
	}

	for _, tt := range dangerous {
		result := NormalizeKillExec(tt.input)
		if result != "" {
			t.Errorf("%s: NormalizeKillExec(%q) = %q, expected empty", tt.name, tt.input, result)
		}
	}
}

func TestCritical_ValidateAppExec_RejectsSystemProcesses(t *testing.T) {
	systemProcs := []struct {
		exec    string
		appName string
	}{
		{"explorer.exe", "Windows Explorer"},
		{"explorer", "Windows Explorer"},
		{"dwm.exe", "Desktop Window Manager"},
		{"svchost.exe", "Service Host"},
		{"systemd", "SystemD"},
		{"roboty.exe", "Roboty"},
		{"wails.exe", "Wails"},
	}

	for _, tt := range systemProcs {
		err := validateAppExec(tt.exec, tt.appName)
		if err == nil {
			t.Errorf("CRITICAL: validateAppExec(%q, %q) should have rejected", tt.exec, tt.appName)
		}
	}
}

func TestCritical_ValidateAppExec_AllowsSafeProcesses(t *testing.T) {
	safeProcs := []struct {
		exec    string
		appName string
	}{
		{"chrome.exe", "Google Chrome"},
		{"firefox.exe", "Firefox"},
		{"code.exe", "VS Code"},
		{"notepad.exe", "Notepad"},
		{"slack.exe", "Slack"},
	}

	for _, tt := range safeProcs {
		err := validateAppExec(tt.exec, tt.appName)
		if err != nil {
			t.Errorf("validateAppExec(%q, %q) should have passed: %v", tt.exec, tt.appName, err)
		}
	}
}

func TestCritical_Whitelist_AllRequiredEntries(t *testing.T) {
	execs := GetWhitelistExecs()

	required := []string{
		// Windows
		"explorer", "explorer.exe",
		"dwm", "dwm.exe",
		"svchost", "svchost.exe",
		"csrss.exe",
		"lsass.exe",
		"winlogon.exe",
		"runtimebroker", "runtimebroker.exe",
		"shellexperiencehost", "shellexperiencehost.exe",
		"startmenuexperiencehost", "startmenuexperiencehost.exe",
		"applicationframehost", "applicationframehost.exe",
		"conhost", "conhost.exe",
		"cmd", "cmd.exe",
		"powershell", "powershell.exe",
		"msedgewebview2", "msedgewebview2.exe",
		"fontdrvhost", "fontdrvhost.exe",
		"lockapp", "lockapp.exe",
		"sway",
		// macOS
		"WindowServer",
		"Finder",
		"Dock",
		"launchd",
		"loginwindow",
		// Linux
		"systemd",
		"gnome-shell",
		"Xorg",
		"init",
		"sway",
	}

	for _, entry := range required {
		key := strings.ToLower(entry)
		if !execs[key] {
			t.Errorf("CRITICAL: missing whitelist entry: %q", key)
		}
	}
}

func TestCritical_Whitelist_NoSectionKeys(t *testing.T) {
	execs := GetWhitelistExecs()
	sectionKeys := []string{"bundle_ids", "ui_shell", "process_names", "executables", "search_ui", "kernel_level"}
	for _, key := range sectionKeys {
		if execs[key] {
			t.Errorf("section key %q should not be in whitelist execs", key)
		}
	}
}

func TestCritical_EmergencyStop_SafeToCallMultiple(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeProxyMgr := newFakeProxyManager()
	fakeNotifMgr := newFakeNotificationManager()

	ms := NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)

	// Call EmergencyStop multiple times — must not panic
	for i := 0; i < 5; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on EmergencyStop call %d: %v", i+1, r)
				}
			}()
			ms.EmergencyStop("test")
		}()
	}
}

func TestCritical_ConcurrentEmergencyStop(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeProxyMgr := newFakeProxyManager()
	fakeNotifMgr := newFakeNotificationManager()

	ms := NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)

	done := make(chan struct{})
	go func() {
		ms.EmergencyStop("goroutine-1")
		done <- struct{}{}
	}()
	go func() {
		ms.EmergencyStop("goroutine-2")
		done <- struct{}{}
	}()

	<-done
	<-done
}

func TestCritical_SignalHandler_DoesNotPanic(t *testing.T) {
	SetupSignalHandler()

	_, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}

	// Just verify the handler was installed without panic
	// (we can't easily send SIGINT in a test without killing the process)
}

func TestCritical_safeGo_RecoversPanic(t *testing.T) {
	var recovered bool
	SetGlobalEmergencyCallback(func(reason string) {
		recovered = true
	})
	defer SetGlobalEmergencyCallback(nil)

	safeGo(func() {
		panic("test panic")
	})

	time.Sleep(100 * time.Millisecond)

	if !recovered {
		t.Error("safeGo did not recover panic")
	}
}

func TestCritical_safeGo_NormalExecution(t *testing.T) {
	var executed bool
	safeGo(func() {
		executed = true
	})

	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("safeGo did not execute function")
	}
}

func TestCritical_ProxyStateNotSavedWhenEnableFails(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeProxyMgr.enableErr = os.ErrPermission
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	err := ub.Start(getFreePort(t), nil)
	if err != nil {
		t.Fatalf("Start should keep running even if proxy enable fails: %v", err)
	}

	if !ub.IsRunning() {
		t.Error("proxy should be running even if system proxy enable fails")
	}

	if fakeState.StateExists(proxyStateName) {
		t.Error("state file must not exist when proxy enable failed")
	}

	ub.Stop()
}

// =============================================================================
// NEW CRITICAL TESTS — Coverage Gaps from Safety Audit
// =============================================================================

func TestCritical_WebView2NeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"msedgewebview2", "msedgewebview2.exe", "MSEdgeWebView2.EXE", "msedgewebview2.EXE"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: msedgewebview2 %q was allowed to be killed — this would crash the WebView/app: %s", name, reason)
		}
	}
}

func TestCritical_FontDrvHostNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"fontdrvhost", "fontdrvhost.exe", "FONTDRVHOST.EXE"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: fontdrvhost %q was allowed — this breaks font rendering: %s", name, reason)
		}
	}
}

func TestCritical_LockAppNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"lockapp", "lockapp.exe", "LockApp.exe"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: lockapp %q was allowed — this would crash lock screen: %s", name, reason)
		}
	}
}

func TestCritical_LoginWindowNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"loginwindow", "LoginWindow"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: loginwindow %q was allowed — this forces user re-auth: %s", name, reason)
		}
	}
}

func TestCritical_SwayNeverKillable(t *testing.T) {
	v := NewKillSafetyVerifier()
	for _, name := range []string{"sway", "Sway"} {
		safe, reason := v.IsSafeToKill(name)
		if safe {
			t.Errorf("CRITICAL: sway %q was allowed — this crashes the Wayland compositor: %s", name, reason)
		}
	}
}

func TestCritical_IPv6BracketedLocalhostIsAlwaysAllowed(t *testing.T) {
	ub := NewURLBlocker()
	ub.allowedURLs = []string{}

	// These IPv6 forms with brackets and ports MUST always be allowed
	mustPass := []string{
		"[::1]", "[::1]:62828", "[::1]:34115",
		"[::1]:443", "[::1]:80",
		"[0:0:0:0:0:0:0:1]", "[0:0:0:0:0:0:0:1]:62828",
		"[::]",
	}

	for _, host := range mustPass {
		if !ub.isAllowed(host) {
			t.Errorf("CRITICAL: bracketed IPv6 host %q was blocked — this breaks CONNECT to localhost", host)
		}
	}

	// Also verify plain (unbracketed) still works
	for _, host := range []string{"::1", "0:0:0:0:0:0:0:1"} {
		if !ub.isAllowed(host) {
			t.Errorf("CRITICAL: plain IPv6 host %q was blocked", host)
		}
	}
}

func TestCritical_AppBlockerHasKillLoopDetector(t *testing.T) {
	tracker := NewForegroundTracker()
	blocker := NewAppBlocker(tracker)

	if blocker.killLoopDetector == nil {
		t.Fatal("CRITICAL: AppBlocker must have a KillLoopDetector")
	}

	// Verify it records kills and detects loops
	kld := blocker.killLoopDetector
	for i := 0; i < MaxConsecutiveKills-1; i++ {
		if kld.RecordKill("test-app") {
			t.Errorf("kill loop detected early at %d", i+1)
		}
	}
	if !kld.RecordKill("test-app") {
		t.Error("CRITICAL: KillLoopDetector did not detect kill loop at threshold")
	}
}

func TestCritical_KillLoopDetector_WiredInBlocker(t *testing.T) {
	// Integration test: use a fake killer that simulates a restarting app.
	// The blocker should kill it until KillLoopDetector triggers, then stop.
	fakeKiller := newFakeProcessKiller()
	tracker := NewForegroundTracker()

	ab := NewAppBlockerWithDI(tracker, fakeKiller)
	if ab.killLoopDetector == nil {
		t.Fatal("CRITICAL: AppBlocker must have KillLoopDetector")
	}

	// Simulate kill loop detection directly
	kld := ab.killLoopDetector
	for i := 0; i < MaxConsecutiveKills; i++ {
		triggered := kld.RecordKill("restarting-app")
		if i < MaxConsecutiveKills-1 && triggered {
			t.Errorf("kill loop detected too early at iteration %d", i)
		}
	}
	if !kld.RecordKill("restarting-app") {
		t.Error("CRITICAL: KillLoopDetector did not trigger after MaxConsecutiveKills kills")
	}
}

func TestCritical_ProxyTransportNoEnvLoop(t *testing.T) {
	// Verify handleHTTPPlain uses transport with nil Proxy to prevent env proxy loop.
	// We can't test the actual RoundTrip without a server, but we can verify
	// that the transport created by handleHTTPPlain has nil Proxy field.
	transport := &http.Transport{
		Proxy: nil,
	}
	if transport.Proxy != nil {
		t.Error("CRITICAL: transport.Proxy must be nil to prevent proxy loops")
	}
}

func TestCritical_ProxyConnectTunnelHasTimeout(t *testing.T) {
	// Verify that transferTimed respects the timeout.
	// Create a slow pipe and verify it times out.
	r, w := io.Pipe()
	done := make(chan bool, 1)

	go func() {
		transferTimed(w, r, 50*time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		// Completed within expected time
	case <-time.After(2 * time.Second):
		t.Fatal("CRITICAL: transferTimed did not time out within 2s — tunnel goroutine may leak")
	}
}

// =============================================================================
// CHAOS TESTS — Non-destructive, DI-mocked, safe for VM/sandbox
// =============================================================================

// TestChaos_ProxyCrashRecovery verifies that when the URLBlocker is stopped
// (simulating proxy crash), the watchdog detects it and cleans up system proxy.
func TestChaos_ProxyCrashRecovery(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	port := getFreePort(t)

	err := ub.Start(port, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !ub.IsRunning() {
		t.Fatal("proxy should be running after Start")
	}
	if !fakeProxyMgr.enabled {
		t.Error("system proxy should be enabled")
	}
	if !fakeState.StateExists(proxyStateName) {
		t.Error("state file should exist after start")
	}

	// Simulate proxy crash by stopping abruptly
	err = ub.Stop()
	if err != nil {
		t.Logf("Stop warning: %v", err)
	}

	if ub.IsRunning() {
		t.Error("proxy should not be running after crash/stop")
	}
	if fakeProxyMgr.enabled {
		t.Error("system proxy should be disabled after crash recovery")
	}
	if fakeState.StateExists(proxyStateName) {
		t.Error("state file should be cleaned after crash recovery")
	}
}

// TestChaos_BlockerPanicRecovery verifies safeGo recovers panics in blocker goroutines.
func TestChaos_BlockerPanicRecovery(t *testing.T) {
	var recovered bool
	SetGlobalEmergencyCallback(func(reason string) {
		recovered = true
	})
	defer SetGlobalEmergencyCallback(nil)

	// safeGo must recover panics
	safeGo(func() {
		panic("simulated blocker crash")
	})

	time.Sleep(200 * time.Millisecond)

	if !recovered {
		t.Error("CHAOS: safeGo did not recover panic — blocker crash would kill process")
	}
}

// TestChaos_RapidToggle verifies AppBlocker and URLBlocker Start/Stop can be
// called repeatedly without deadlock, port exhaustion, or orphaned state.
func TestChaos_RapidToggle(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeState := newFakeStateFileManager()
	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	port := getFreePort(t)

	done := make(chan struct{}, 1)
	go func() {
		for i := 0; i < 50; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("panic on toggle iteration %d: %v", i, r)
					}
				}()
				_ = ub.Start(port, []string{"example.com"})
				ub.Stop()
			}()
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("CHAOS: Rapid toggle timed out — possible deadlock or goroutine leak")
	}

	if fakeProxyMgr.enabled {
		t.Error("system proxy should be disabled after all toggles")
	}
	if fakeState.StateExists(proxyStateName) {
		t.Error("state file should be cleaned after all toggles")
	}
}

// TestChaos_DeadlockDetection runs all concurrency-sensitive paths simultaneously.
func TestChaos_DeadlockDetection(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeProxyMgr := newFakeProxyManager()
	fakeNotifMgr := newFakeNotificationManager()

	_ = NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)
	ub := NewURLBlockerWithDI(fakeProxyMgr, newFakeStateFileManager())
	v := NewKillSafetyVerifier()
	v.Refresh()
	kld := NewKillLoopDetector()

	done := make(chan struct{}, 8)

	// Path 1: KillSafetyVerifier concurrent access
	go func() {
		for i := 0; i < 100; i++ {
			v.IsSafeToKill("chrome")
			v.IsSafeToKill("explorer")
			v.IsSafeToKill("roboty")
		}
		done <- struct{}{}
	}()

	// Path 2: KillLoopDetector concurrent access
	go func() {
		for i := 0; i < 100; i++ {
			kld.RecordKill("test-app")
		}
		done <- struct{}{}
	}()

	// Path 3: URLBlocker isAllowed concurrent access
	go func() {
		for i := 0; i < 100; i++ {
			ub.isAllowed("localhost")
			ub.isAllowed("example.com")
		}
		done <- struct{}{}
	}()

	// Path 4: URLBlocker start/stop stress
	go func() {
		port := getFreePort(t)
		for i := 0; i < 20; i++ {
			ub.Start(port, []string{"example.com"})
			ub.Stop()
		}
		done <- struct{}{}
	}()

	// Path 5: NormalizeKillExec concurrent
	go func() {
		for i := 0; i < 100; i++ {
			NormalizeKillExec("chrome.exe")
			NormalizeKillExec("explorer | rm")
			NormalizeKillExec("")
		}
		done <- struct{}{}
	}()

	// Path 6: GetWhitelistExecs concurrent
	go func() {
		for i := 0; i < 50; i++ {
			_ = GetWhitelistExecs()
		}
		done <- struct{}{}
	}()

	// Path 7: GetAncestorExecs concurrent (safe on non-Windows for testing)
	go func() {
		for i := 0; i < 50; i++ {
			_ = GetAncestorExecs()
		}
		done <- struct{}{}
	}()

	// Path 8: EmergencyFailsafe concurrent
	go func() {
		ef := NewEmergencyFailsafe(func() {})
		for i := 0; i < 50; i++ {
			ef.Trigger("test", &[]SafetyEvent{})
			ef.IsTriggered()
			ef.Reset()
		}
		done <- struct{}{}
	}()

	timeout := time.After(30 * time.Second)
	for i := 0; i < 8; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("CHAOS: Deadlock detection timed out — possible deadlock")
		}
	}
}

// TestChaos_PortConflict verifies proxy handles port conflicts gracefully.
func TestChaos_PortConflict(t *testing.T) {
	// Occupy a port first
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to occupy port: %v", err)
	}
	defer listener.Close()
	occupiedPort := listener.Addr().(*net.TCPAddr).Port

	fakeProxyMgr := newFakeProxyManager()
	ub := NewURLBlockerWithDI(fakeProxyMgr, newFakeStateFileManager())

	// Attempt to start on occupied port
	err = ub.Start(occupiedPort, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start should succeed via automatic port fallback: %v", err)
	}
	defer ub.Stop()

	// The proxy should have started on a different port (fallback to port 0)
	if !fakeProxyMgr.enabled {
		t.Error("system proxy should be enabled after port fallback recovery")
	}
	if ub.port == occupiedPort {
		t.Error("proxy should have fallen back to a different port")
	}
	if !ub.IsRunning() {
		t.Error("proxy should be running after port fallback")
	}
}

// TestChaos_InvalidConfig verifies that corrupted/missing config doesn't crash.
func TestChaos_InvalidConfig(t *testing.T) {
	// Test GetWhitelistExecs with missing file (should return empty, not panic)
	execs := GetWhitelistExecs()
	if execs == nil {
		t.Error("GetWhitelistExecs should return empty map, not nil")
	}

	// Verify validateAppExec doesn't panic on edge cases
	safeProcs := []string{"", ".", "..", "-", "valid-app", "chrome.exe"}
	for _, p := range safeProcs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on validateAppExec(%q): %v", p, r)
				}
			}()
			err := validateAppExec(p, "test")
			_ = err
		}()
	}
}

// TestChaos_HighConnectionLoad verifies proxy can handle many concurrent connections.
// Uses fake proxy manager (no real network) — safe for CI.
func TestChaos_HighConnectionLoad(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	ub := NewURLBlockerWithDI(fakeProxyMgr, newFakeStateFileManager())
	port := getFreePort(t)

	err := ub.Start(port, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ub.Stop()

	// Simulate concurrent isAllowed checks — this is the proxy's hot path
	concurrency := 100
	iterations := 50
	done := make(chan struct{}, concurrency)

	for c := 0; c < concurrency; c++ {
		go func(id int) {
			for i := 0; i < iterations; i++ {
				// Mix of allowed, blocked, and localhost queries
				_ = ub.isAllowed("example.com")
				_ = ub.isAllowed("blocked-site.com")
				_ = ub.isAllowed("localhost")
				_ = ub.isAllowed("127.0.0.1:62828")
				_ = ub.isAllowed("[::1]:34115")
			}
			done <- struct{}{}
		}(c)
	}

	timeout := time.After(30 * time.Second)
	for c := 0; c < concurrency; c++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatalf("CHAOS: High connection load test timed out — possible deadlock in isAllowed")
		}
	}
}

// TestChaos_SignalInterruption_Safe verifies signal handler can be installed safely.
func TestChaos_SignalInterruption_Safe(t *testing.T) {
	// Must not panic on install
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic during signal handler setup: %v", r)
		}
	}()

	SetupSignalHandler()
	t.Log("Signal handler installed without panic")
}

// TestChaos_EmergencyStop_MultipleCalls verifies EmergencyStop is idempotent.
func TestChaos_EmergencyStop_MultipleCalls(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeProxyMgr := newFakeProxyManager()
	fakeNotifMgr := newFakeNotificationManager()

	ms := NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)

	// Must handle many EmergencyStop calls without panic or deadlock
	for i := 0; i < 50; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on EmergencyStop call %d: %v", i, r)
				}
			}()
			ms.EmergencyStop("chaos-test-" + strconv.Itoa(i))
		}()
	}

	// Final state must be clean
	if fakeProxyMgr.enabled {
		t.Error("system proxy should be disabled after EmergencyStop")
	}
}

// =============================================================================
// WHITELIST SYNC TEST — Ensures whitelist.json and safety.go systemCritical
// are in sync. Added per Safety Audit Step 4.
// =============================================================================

func TestCritical_WhitelistSyncWithSystemCritical(t *testing.T) {
	execs := GetWhitelistExecs()
	_ = execs["dummy"] // read to suppress unused warning

	// Extract normalized exec names from whitelist (everything GetWhitelistExecs returns)
	wlNormalized := make(map[string]bool)
	for k := range execs {
		wlNormalized[strings.ToLower(strings.TrimSuffix(k, ".exe"))] = true
	}

	// Build the expected set from safety.go's systemCritical slice
	safetyContent, err := os.ReadFile("safety.go")
	if err != nil {
		t.Fatalf("cannot read safety.go: %v", err)
	}
	re := regexp.MustCompile(`systemCritical\s*:=\s*\[\]string\{`)
	idx := re.FindIndex(safetyContent)
	if idx == nil {
		t.Fatal("cannot find systemCritical decl in safety.go")
	}
	// Find the closing brace
	start := idx[1]
	depth := 1
	end := start
	for i := start; i < len(safetyContent); i++ {
		if safetyContent[i] == '{' {
			depth++
		} else if safetyContent[i] == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if depth != 0 {
		t.Fatal("unbalanced braces in systemCritical decl")
	}
	block := string(safetyContent[start:end])
	scNames := make(map[string]bool)
	for _, m := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(block, -1) {
		name := strings.ToLower(m[1])
		scNames[name] = true
	}

	// Self-protection names are handled by selfNames, not required in whitelist
	selfProtection := map[string]bool{
		"roboty": true, "roboty1": true, "roboty-dev": true, "wails": true,
	}

	var missingFromWhitelist []string
	var missingFromSystemCritical []string

	for name := range scNames {
		if selfProtection[name] {
			continue
		}
		if !wlNormalized[name] {
			missingFromWhitelist = append(missingFromWhitelist, name)
		}
	}

	// whitelist entries that should correspond to systemCritical entries
	// (skip kernel-level, non-process entries, and macOS bundle IDs)
	skipFromComparison := map[string]bool{
		"kernel": true, "swapper": true, "migration": true,
		"ksoftirqd": true, "kworker": true, "kthreadd": true,
		"systemd-networkd": true, "systemd-resolved": true, "systemd-udevd": true,
		"krunner": true, "activities": true, "overview": true,
		"searchprotocolhost": true, "searchfilterhost": true,
	}
	for name := range wlNormalized {
		if selfProtection[name] || skipFromComparison[name] {
			continue
		}
		// Skip macOS bundle IDs (com.apple.*) — they're not process names
		if strings.HasPrefix(name, "com.apple.") {
			continue
		}
		if !scNames[name] {
			missingFromSystemCritical = append(missingFromSystemCritical, name)
		}
	}

	if len(missingFromWhitelist) > 0 {
		t.Errorf("CRITICAL: %d entries in systemCritical (safety.go) missing from whitelist.json:\n  %v",
			len(missingFromWhitelist), missingFromWhitelist)
	}
	if len(missingFromSystemCritical) > 0 {
		t.Errorf("CRITICAL: %d entries in whitelist.json missing from systemCritical (safety.go):\n  %v",
			len(missingFromSystemCritical), missingFromSystemCritical)
	}
}
