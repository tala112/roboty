//go:build integration

package integration

import (
	"testing"

	"Roboty/internal/modes"
)

// TestIntegration_URLBlocker_WithFakeProxyManager verifies URL blocker lifecycle with DI fakes.
func TestIntegration_URLBlocker_WithFakeProxyManager(t *testing.T) {
	fakeProxy := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)

	port := 62832
	err := ub.Start(port, []string{"example.com"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ub.Stop()

	if !fakeState.StateExists("roboty_proxy_state") {
		t.Error("expected proxy state to exist after start")
	}

	ub.Stop()

	if fakeState.StateExists("roboty_proxy_state") {
		t.Error("expected state file cleaned after stop")
	}
}

type fakeProxyManager struct {
	enabled bool
}

func newFakeProxyManager() *fakeProxyManager {
	return &fakeProxyManager{}
}

func (f *fakeProxyManager) Enable(proxyAddr string, port int) error {
	f.enabled = true
	return nil
}

func (f *fakeProxyManager) Disable() error {
	f.enabled = false
	return nil
}

func (f *fakeProxyManager) IsEnabled() (bool, error) {
	return f.enabled, nil
}

type fakeStateFileManager struct {
	states map[string]bool
}

func newFakeStateFileManager() *fakeStateFileManager {
	return &fakeStateFileManager{states: make(map[string]bool)}
}

func (f *fakeStateFileManager) SaveState(name string) error {
	f.states[name] = true
	return nil
}

func (f *fakeStateFileManager) ClearState(name string) error {
	delete(f.states, name)
	return nil
}

func (f *fakeStateFileManager) StateExists(name string) bool {
	return f.states[name]
}

func (f *fakeStateFileManager) ListOrphans() ([]string, error) {
	var r []string
	for n := range f.states {
		r = append(r, n)
	}
	return r, nil
}

func (f *fakeStateFileManager) CleanupAll() error {
	f.states = make(map[string]bool)
	return nil
}
