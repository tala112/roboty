package modes

import (
	"fmt"
	"time"

	"Roboty/internal/platform"
)

type realProcessKiller struct{}

func NewRealProcessKiller() ProcessKiller {
	return &realProcessKiller{}
}

func (k *realProcessKiller) Kill(execName string, timeout time.Duration) error {
	if p := platform.GetGlobal(); p != nil {
		return p.Kill(execName, timeout)
	}
	return fmt.Errorf("no platform available")
}

func (k *realProcessKiller) IsRunning(execName string) (bool, error) {
	if p := platform.GetGlobal(); p != nil {
		return p.IsRunning(execName)
	}
	return false, nil
}

func (k *realProcessKiller) ListRunning() ([]ProcessInfo, error) {
	if p := platform.GetGlobal(); p != nil {
		pi, err := p.ListRunning()
		if err != nil {
			return nil, err
		}
		result := make([]ProcessInfo, len(pi))
		for i, v := range pi {
			result[i] = ProcessInfo{Name: v.Name, Exec: v.Exec, PID: v.PID}
		}
		return result, nil
	}
	return nil, nil
}
