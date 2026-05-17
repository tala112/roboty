package faken

import (
	"fmt"
	"sync"
	"time"

	"Roboty/internal/platform/types"
)

type killRecord struct {
	ExecName  string
	Timestamp time.Time
}

type FakePlatform struct {
	mu sync.Mutex

	killed    map[string]int
	running   map[string]bool
	blockList map[string]string
	killLog   []killRecord

	proxyEnabled   bool
	proxyEnableErr error

	notifMuted   bool
	notifMuteErr error

	foregroundResult *types.ForegroundActivity
	foregroundErr    error

	ancestors   map[string]bool
	ancestorErr error

	installedApps []types.InstalledApp
	installErr    error
}

func NewFakePlatform() *FakePlatform {
	return &FakePlatform{
		killed:    make(map[string]int),
		running:   make(map[string]bool),
		blockList: make(map[string]string),
		ancestors: make(map[string]bool),
	}
}

func (f *FakePlatform) Kill(execName string, timeout time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if reason, blocked := f.blockList[execName]; blocked {
		return fmt.Errorf("safety blocked: %s", reason)
	}
	f.killed[execName]++
	f.running[execName] = false
	f.killLog = append(f.killLog, killRecord{ExecName: execName, Timestamp: time.Now()})
	return nil
}

func (f *FakePlatform) IsRunning(execName string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.running[execName], nil
}

func (f *FakePlatform) ListRunning() ([]types.ProcessInfo, error) {
	return nil, nil
}

func (f *FakePlatform) SetRunning(execName string, running bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running[execName] = running
}

func (f *FakePlatform) BlockKill(execName, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blockList[execName] = reason
}

func (f *FakePlatform) KillCount(execName string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.killed[execName]
}

func (f *FakePlatform) Poll() (*types.ForegroundActivity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.foregroundErr != nil {
		return nil, f.foregroundErr
	}
	return f.foregroundResult, nil
}

func (f *FakePlatform) SetForegroundResult(act *types.ForegroundActivity) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.foregroundResult = act
}

func (f *FakePlatform) SetForegroundErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.foregroundErr = err
}

func (f *FakePlatform) Enable(proxyAddr string, port int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.proxyEnableErr != nil {
		return f.proxyEnableErr
	}
	f.proxyEnabled = true
	return nil
}

func (f *FakePlatform) Disable() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.proxyEnabled = false
	return nil
}

func (f *FakePlatform) IsEnabled() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.proxyEnabled, nil
}

func (f *FakePlatform) SetProxyEnableErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.proxyEnableErr = err
}

func (f *FakePlatform) Mute() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.notifMuteErr != nil {
		return f.notifMuteErr
	}
	f.notifMuted = true
	return nil
}

func (f *FakePlatform) Restore() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifMuted = false
	return nil
}

func (f *FakePlatform) IsMuted() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.notifMuted, nil
}

func (f *FakePlatform) GetAncestorExecs() (map[string]bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ancestorErr != nil {
		return nil, f.ancestorErr
	}
	result := make(map[string]bool)
	for k := range f.ancestors {
		result[k] = true
	}
	return result, nil
}

func (f *FakePlatform) GetParentPID(pid int) (int, error) {
	return 0, nil
}

func (f *FakePlatform) GetProcessName(pid int) (string, error) {
	return "", nil
}

func (f *FakePlatform) SetAncestors(ancestors map[string]bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ancestors = ancestors
}

func (f *FakePlatform) GetInstalledApps() ([]types.InstalledApp, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.installErr != nil {
		return nil, f.installErr
	}
	return f.installedApps, nil
}

func (f *FakePlatform) SetInstalledApps(apps []types.InstalledApp) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.installedApps = apps
}
