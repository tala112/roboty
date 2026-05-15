package modes

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

type AppBlocker struct {
	mu         sync.Mutex
	running    bool
	cancel     context.CancelFunc
	foreground *ForegroundTracker
}

func NewAppBlocker(tracker *ForegroundTracker) *AppBlocker {
	return &AppBlocker{
		foreground: tracker,
	}
}

func (ab *AppBlocker) Start(allowedExecs []string, closeOnActivate []string, interval time.Duration) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.running {
		return
	}

	if len(closeOnActivate) > 0 {
		CloseApps(closeOnActivate)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ab.cancel = cancel
	ab.running = true

	allowedSet := make(map[string]bool)
	for _, e := range allowedExecs {
		allowedSet[strings.ToLower(e)] = true
	}
	// Never block whitelisted system processes
	for e := range GetWhitelistExecs() {
		allowedSet[e] = true
	}

	go func() {
		defer func() {
			ab.mu.Lock()
			ab.running = false
			ab.mu.Unlock()
		}()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[blocker] stopped")
				return
			case <-ticker.C:
				activity, err := ab.foreground.Poll()
				if err != nil {
					log.Printf("[blocker] poll error: %v", err)
					continue
				}
				if activity == nil {
					continue
				}

				execLower := strings.ToLower(activity.ExecName)

				if !allowedSet[execLower] {
					log.Printf("[blocker] blocking %s (exec=%s) — not in allowed list", activity.AppName, activity.ExecName)
					CloseApp(activity.ExecName)
				}
			}
		}
	}()
}

func (ab *AppBlocker) Stop() {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.cancel != nil {
		ab.cancel()
		ab.cancel = nil
	}
	ab.running = false
}

func (ab *AppBlocker) IsRunning() bool {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	return ab.running
}
