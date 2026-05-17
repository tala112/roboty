package modes

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"Roboty/internal/platform"
)

type ForegroundActivity struct {
	AppName     string    `json:"app_name"`
	ExecName    string    `json:"exec_name"`
	WindowTitle string    `json:"window_title"`
	PID         int       `json:"pid"`
	Timestamp   time.Time `json:"timestamp"`
}

type ForegroundTracker struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewForegroundTracker() *ForegroundTracker {
	return &ForegroundTracker{}
}

func (ft *ForegroundTracker) Start(ctx context.Context, interval time.Duration, callback func(ForegroundActivity)) {
	ft.mu.Lock()
	if ft.cancel != nil {
		ft.cancel()
	}
	ft.ctx, ft.cancel = context.WithCancel(ctx)
	ft.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ft.ctx.Done():
				return
			case <-ticker.C:
				activity, err := ft.Poll()
				if err != nil {
					log.Printf("[tracker] poll error: %v", err)
					continue
				}
				if activity != nil {
					callback(*activity)
				}
			}
		}
	}()
}

func (ft *ForegroundTracker) Stop() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if ft.cancel != nil {
		ft.cancel()
		ft.cancel = nil
	}
}

func (ft *ForegroundTracker) Poll() (*ForegroundActivity, error) {
	p := platform.GetGlobal()
	if p == nil {
		return nil, fmt.Errorf("no platform available")
	}
	act, err := p.Poll()
	if err != nil {
		return nil, err
	}
	if act == nil {
		return nil, nil
	}
	return &ForegroundActivity{
		AppName:     act.AppName,
		ExecName:    act.ExecName,
		WindowTitle: act.WindowTitle,
		PID:         act.PID,
		Timestamp:   act.Timestamp,
	}, nil
}

func (ft *ForegroundTracker) ListRunningProcesses() ([]InstalledApp, error) {
	p := platform.GetGlobal()
	if p == nil {
		return nil, nil
	}
	procs, err := p.ListRunning()
	if err != nil {
		return nil, err
	}
	apps := make([]InstalledApp, 0, len(procs))
	seen := make(map[string]bool)
	for _, proc := range procs {
		key := proc.Exec
		if seen[key] {
			continue
		}
		seen[key] = true
		apps = append(apps, InstalledApp{
			Name: proc.Name,
			Exec: proc.Exec,
		})
	}
	return apps, nil
}
