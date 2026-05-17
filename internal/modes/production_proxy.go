package modes

import (
	"Roboty/internal/platform"
)

type realProxyManager struct{}

func NewRealProxyManager() SystemProxyManager {
	return &realProxyManager{}
}

func (m *realProxyManager) Enable(proxyAddr string, port int) error {
	if p := platform.GetGlobal(); p != nil {
		return p.Enable(proxyAddr, port)
	}
	return nil
}

func (m *realProxyManager) Disable() error {
	if p := platform.GetGlobal(); p != nil {
		return p.Disable()
	}
	return nil
}

func (m *realProxyManager) IsEnabled() (bool, error) {
	if p := platform.GetGlobal(); p != nil {
		return p.IsEnabled()
	}
	return false, nil
}
