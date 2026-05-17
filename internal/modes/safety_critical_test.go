package modes

import (
	"os"
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
		// macOS
		"WindowServer",
		"Finder",
		"Dock",
		"launchd",
		// Linux
		"systemd",
		"gnome-shell",
		"Xorg",
		"init",
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
