package modes

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AppBlocker struct {
	mu              sync.Mutex
	running         bool
	cancel          context.CancelFunc
	foreground      ForegroundDetector
	safetyVerifier  *KillSafetyVerifier
	killLoopDetector *KillLoopDetector
	auditLogger     *SafetyAuditLogger
	killer          ProcessKiller
}

func NewAppBlocker(tracker ForegroundDetector) *AppBlocker {
	return &AppBlocker{
		foreground:       tracker,
		killer:           NewRealProcessKiller(),
		safetyVerifier:   GetGlobalSafetyVerifier(),
		killLoopDetector: NewKillLoopDetector(),
		auditLogger:      NewSafetyAuditLogger(500),
	}
}

func NewAppBlockerWithDI(tracker ForegroundDetector, killer ProcessKiller) *AppBlocker {
	return &AppBlocker{
		foreground:       tracker,
		killer:           killer,
		safetyVerifier:   GetGlobalSafetyVerifier(),
		killLoopDetector: NewKillLoopDetector(),
		auditLogger:      NewSafetyAuditLogger(500),
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
		key := strings.ToLower(strings.TrimSuffix(e, ".exe"))
		allowedSet[key] = true
	}
	// Never block whitelisted system processes
	// CRITICAL: normalize keys to match tracker output (tracker strips .exe)
	for e := range GetWhitelistExecs() {
		key := strings.ToLower(strings.TrimSuffix(e, ".exe"))
		allowedSet[key] = true
	}
	// Add ancestor processes (never block own parent/launcher/terminal)
	for e := range GetAncestorExecs() {
		key := strings.ToLower(strings.TrimSuffix(e, ".exe"))
		allowedSet[key] = true
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
				execKey := strings.TrimSuffix(execLower, ".exe")

				if !allowedSet[execKey] {
					log.Printf("[blocker] blocking %s (exec=%s) — not in allowed list", activity.AppName, activity.ExecName)

					// SAFETY: Normalize and verify before killing
					safeExec := NormalizeKillExec(activity.ExecName)
					if safeExec == "" {
						log.Printf("[blocker] SKIP kill of %s: rejected by NormalizeKillExec", activity.ExecName)
						ab.auditLogger.Log(SafetyEvent{
							Type:    EventKillBlocked,
							Target:  activity.ExecName,
							Source:  "blocker",
							Message: "rejected by NormalizeKillExec",
						})
						continue
					}

					safe, reason := ab.safetyVerifier.IsSafeToKill(safeExec)
					if !safe {
						log.Printf("[blocker] SAFETY BLOCKED kill of %s (exec=%s): %s", activity.AppName, activity.ExecName, reason)
						ab.auditLogger.Log(SafetyEvent{
							Type:    EventKillBlocked,
							Target:  activity.ExecName,
							Source:  "blocker-safety-verifier",
							Message: reason,
						})
						continue
					}

					// Kill loop detection — prevent rapid re-kill cycles
					if ab.killLoopDetector.RecordKill(safeExec) {
						log.Printf("[blocker] 🚫 KILL LOOP DETECTED for %s (exec=%s) — stopping attempts", activity.AppName, activity.ExecName)
						ab.auditLogger.Log(SafetyEvent{
							Type:    EventKillBlocked,
							Target:  activity.ExecName,
							Source:  "blocker-kill-loop-detector",
							Message: "kill loop detected after " + strconv.Itoa(MaxConsecutiveKills) + " attempts",
						})
						continue
					}

					// Safety check passed — proceed with kill
					ab.auditLogger.Log(SafetyEvent{
						Type:    EventBlockedApp,
						Target:  activity.ExecName,
						Source:  "blocker",
						Message: "blocked by focus mode",
					})
					if ab.killer != nil {
						ab.killer.Kill(activity.ExecName, 10*time.Second)
					} else {
						CloseApp(activity.ExecName)
					}
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
