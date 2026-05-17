package modes

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Test Fakes (white-box testing — package modes)
// =============================================================================

type killRecord struct {
	execName  string
	timestamp time.Time
}

type fakeProcessKiller struct {
	mu         sync.Mutex
	killed     map[string]int
	running    map[string]bool
	blockList  map[string]string
	failOnKill map[string]error
	killLog    []killRecord
}

func newFakeProcessKiller() *fakeProcessKiller {
	return &fakeProcessKiller{
		killed:     make(map[string]int),
		running:    make(map[string]bool),
		blockList:  make(map[string]string),
		failOnKill: make(map[string]error),
	}
}

func (f *fakeProcessKiller) Kill(execName string, timeout time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if reason, blocked := f.blockList[execName]; blocked {
		return fmt.Errorf("safety blocked: %s", reason)
	}
	if err, shouldFail := f.failOnKill[execName]; shouldFail {
		return err
	}

	f.killed[execName]++
	f.running[execName] = false
	f.killLog = append(f.killLog, killRecord{execName: execName, timestamp: time.Now()})
	return nil
}

func (f *fakeProcessKiller) IsRunning(execName string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.running[execName], nil
}

func (f *fakeProcessKiller) ListRunning() ([]ProcessInfo, error) {
	return nil, nil
}

func (f *fakeProcessKiller) setRunning(execName string, running bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running[execName] = running
}

func (f *fakeProcessKiller) blockKill(execName, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blockList[execName] = reason
}

func (f *fakeProcessKiller) killCount(execName string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.killed[execName]
}

type fakeProxyManager struct {
	mu         sync.Mutex
	enabled    bool
	enableErr  error
	disableErr error
}

func newFakeProxyManager() *fakeProxyManager {
	return &fakeProxyManager{}
}

func (f *fakeProxyManager) Enable(proxyAddr string, port int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.enableErr != nil {
		return f.enableErr
	}
	f.enabled = true
	return nil
}

func (f *fakeProxyManager) Disable() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.disableErr != nil {
		return f.disableErr
	}
	f.enabled = false
	return nil
}

func (f *fakeProxyManager) IsEnabled() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.enabled, nil
}

func (f *fakeProxyManager) assertEnabled(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.enabled {
		t.Error("expected proxy to be enabled")
	}
}

func (f *fakeProxyManager) assertDisabled(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.enabled {
		t.Error("expected proxy to be disabled")
	}
}

type fakeNotificationManager struct {
	mu      sync.Mutex
	muted   bool
	muteErr error
}

func newFakeNotificationManager() *fakeNotificationManager {
	return &fakeNotificationManager{}
}

func (f *fakeNotificationManager) Mute() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.muteErr != nil {
		return f.muteErr
	}
	f.muted = true
	return nil
}

func (f *fakeNotificationManager) Restore() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.muted = false
	return nil
}

func (f *fakeNotificationManager) IsMuted() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.muted, nil
}

type fakeStateFileManager struct {
	mu     sync.Mutex
	states map[string]bool
}

func newFakeStateFileManager() *fakeStateFileManager {
	return &fakeStateFileManager{states: make(map[string]bool)}
}

func (f *fakeStateFileManager) SaveState(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states[name] = true
	return nil
}

func (f *fakeStateFileManager) ClearState(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.states, name)
	return nil
}

func (f *fakeStateFileManager) StateExists(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.states[name]
}

func (f *fakeStateFileManager) ListOrphans() ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var r []string
	for n := range f.states {
		r = append(r, n)
	}
	return r, nil
}

func (f *fakeStateFileManager) CleanupAll() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states = make(map[string]bool)
	return nil
}

// =============================================================================
// Tests
// =============================================================================

func TestNewModeServiceWithDI_AllFakes(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeProxyMgr := newFakeProxyManager()
	fakeNotifMgr := newFakeNotificationManager()

	ms := NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)
	if ms == nil {
		t.Fatal("NewModeServiceWithDI returned nil")
	}
	if ms.killer == nil {
		t.Error("killer should be set")
	}
	if ms.proxyMgr == nil {
		t.Error("proxyMgr should be set")
	}
	if ms.notifMgr == nil {
		t.Error("notifMgr should be set")
	}
}

func TestAppBlockerWithDI_UsesFakeKiller(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeKiller.setRunning("chrome", true)

	ab := NewAppBlockerWithDI(nil, fakeKiller)
	if ab == nil {
		t.Fatal("NewAppBlockerWithDI returned nil")
	}
}

func TestURLBlockerWithDI_UsesFakeProxy(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	if ub == nil {
		t.Fatal("NewURLBlockerWithDI returned nil")
	}

	port := getFreePort(t)
	err := ub.Start(port, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ub.Stop()

	fakeProxyMgr.assertEnabled(t)
	ub.Stop()
	fakeProxyMgr.assertDisabled(t)
}

func TestURLBlockerWithDI_ProxyEnableFailure_NoOrphan(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeProxyMgr.enableErr = fmt.Errorf("simulated failure")
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	port := getFreePort(t)
	err := ub.Start(port, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start should keep running even if proxy enable fails: %v", err)
	}

	if fakeState.StateExists(proxyStateName) {
		t.Error("state file should not exist when proxy enable fails")
	}

	if !ub.IsRunning() {
		t.Error("proxy should be running even if system proxy enable fails")
	}

	ub.Stop()
	if ub.IsRunning() {
		t.Error("proxy should not be running after Stop")
	}
}

func TestURLBlockerWithDI_ProxyDisableOnStop(t *testing.T) {
	fakeProxyMgr := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxyMgr, fakeState)
	port := getFreePort(t)
	ub.Start(port, []string{"example.com"})
	ub.Stop()

	fakeProxyMgr.assertDisabled(t)
}

func TestFakeProcessKiller_KillRecords(t *testing.T) {
	fakeKiller := newFakeProcessKiller()

	err := fakeKiller.Kill("chrome", 0)
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}
	if fakeKiller.killCount("chrome") != 1 {
		t.Errorf("expected 1 kill, got %d", fakeKiller.killCount("chrome"))
	}
}

func TestFakeProcessKiller_BlockKill(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeKiller.blockKill("explorer", "system process")

	err := fakeKiller.Kill("explorer", 0)
	if err == nil {
		t.Error("expected error for blocked process")
	}
	if fakeKiller.killCount("explorer") != 0 {
		t.Error("blocked process should not increment kill count")
	}
}

func TestFakeProcessKiller_KillStopsRunning(t *testing.T) {
	fakeKiller := newFakeProcessKiller()
	fakeKiller.setRunning("chrome", true)

	running, _ := fakeKiller.IsRunning("chrome")
	if !running {
		t.Error("expected chrome to be running")
	}

	fakeKiller.Kill("chrome", 0)
	running, _ = fakeKiller.IsRunning("chrome")
	if running {
		t.Error("expected chrome to not be running after kill")
	}
}

func TestFakeNotificationManager_MuteRestore(t *testing.T) {
	fakeNotif := newFakeNotificationManager()

	muted, _ := fakeNotif.IsMuted()
	if muted {
		t.Error("should not be muted initially")
	}

	fakeNotif.Mute()
	muted, _ = fakeNotif.IsMuted()
	if !muted {
		t.Error("should be muted after Mute()")
	}

	fakeNotif.Restore()
	muted, _ = fakeNotif.IsMuted()
	if muted {
		t.Error("should not be muted after Restore()")
	}
}

func TestFakeStateFileManager_SaveClearExists(t *testing.T) {
	sf := newFakeStateFileManager()

	if sf.StateExists("test") {
		t.Error("should not exist before save")
	}

	sf.SaveState("test")
	if !sf.StateExists("test") {
		t.Error("should exist after save")
	}

	orphans, _ := sf.ListOrphans()
	if len(orphans) != 1 || orphans[0] != "test" {
		t.Errorf("expected 1 orphan 'test', got %v", orphans)
	}

	sf.ClearState("test")
	if sf.StateExists("test") {
		t.Error("should not exist after clear")
	}

	sf.SaveState("a")
	sf.SaveState("b")
	sf.CleanupAll()
	orphans, _ = sf.ListOrphans()
	if len(orphans) != 0 {
		t.Errorf("expected 0 orphans after cleanup, got %v", orphans)
	}
}

func TestFakeProxyManager_EnableDisable(t *testing.T) {
	fakeProxy := newFakeProxyManager()

	enabled, _ := fakeProxy.IsEnabled()
	if enabled {
		t.Error("should not be enabled initially")
	}

	fakeProxy.Enable("127.0.0.1", 62828)
	enabled, _ = fakeProxy.IsEnabled()
	if !enabled {
		t.Error("should be enabled after Enable()")
	}

	fakeProxy.Disable()
	enabled, _ = fakeProxy.IsEnabled()
	if enabled {
		t.Error("should not be enabled after Disable()")
	}
}

func TestFakeProxyManager_EnableError(t *testing.T) {
	fakeProxy := newFakeProxyManager()
	fakeProxy.enableErr = fmt.Errorf("access denied")

	err := fakeProxy.Enable("127.0.0.1", 62828)
	if err == nil {
		t.Error("expected error when enableErr set")
	}

	enabled, _ := fakeProxy.IsEnabled()
	if enabled {
		t.Error("should not be enabled after failed Enable()")
	}
}
