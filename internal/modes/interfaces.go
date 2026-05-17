package modes

import (
	"context"
	"time"

	"Roboty/internal/db"
)

// ProcessKiller abstracts OS-specific process termination.
type ProcessKiller interface {
	Kill(execName string, timeout time.Duration) error
	IsRunning(execName string) (bool, error)
	ListRunning() ([]ProcessInfo, error)
}

type ProcessInfo struct {
	Name string
	Exec string
	PID  int
}

// SystemProxyManager abstracts OS-level system proxy settings.
type SystemProxyManager interface {
	Enable(proxyAddr string, port int) error
	Disable() error
	IsEnabled() (bool, error)
}

// NotificationManager abstracts OS notification mute/restore.
type NotificationManager interface {
	Mute() error
	Restore() error
	IsMuted() (bool, error)
}

// ForegroundDetector abstracts foreground window tracking.
type ForegroundDetector interface {
	Poll() (*ForegroundActivity, error)
}

// ProcessTreeWalker abstracts ancestor process detection.
type ProcessTreeWalker interface {
	GetAncestorExecs() (map[string]bool, error)
	GetParentPID(pid int) (int, error)
	GetProcessName(pid int) (string, error)
}

// StateFileManager abstracts persistent state file operations.
type StateFileManager interface {
	SaveState(name string) error
	ClearState(name string) error
	StateExists(name string) bool
	ListOrphans() ([]string, error)
	CleanupAll() error
}

// URLBlockerService abstracts the URL blocking proxy lifecycle.
type URLBlockerService interface {
	Start(port int, allowedURLs []string) error
	Stop() error
	SetAllowedURLs(urls []string)
	IsRunning() bool
	IsAllowed(host string) bool
}

// FocusDataStore abstracts all focus mode database operations.
// db.Queries implements this interface implicitly.
type FocusDataStore interface {
	CreateFocusMode(ctx context.Context, arg db.CreateFocusModeParams) (*db.FocusMode, error)
	GetFocusModeByID(ctx context.Context, id string) (*db.FocusMode, error)
	GetAllFocusModes(ctx context.Context) ([]db.FocusMode, error)
	UpdateFocusMode(ctx context.Context, arg db.UpdateFocusModeParams) (*db.FocusMode, error)
	DeleteFocusMode(ctx context.Context, id string) error

	CreateFocusSession(ctx context.Context, arg db.CreateFocusSessionParams) (*db.FocusModeSession, error)
	GetFocusSessionByID(ctx context.Context, id string) (*db.FocusModeSession, error)
	GetActiveFocusSession(ctx context.Context) (*db.FocusModeSession, error)
	GetAllFocusSessions(ctx context.Context) ([]db.FocusModeSession, error)
	UpdateFocusSessionStatus(ctx context.Context, id, status string) (*db.FocusModeSession, error)

	CreateFocusModeApp(ctx context.Context, arg db.CreateFocusModeAppParams) (*db.FocusModeApp, error)
	GetFocusModeAppsByModeID(ctx context.Context, modeID string) ([]db.FocusModeApp, error)
	GetFocusModeAllowedAppsByModeID(ctx context.Context, modeID string) ([]db.FocusModeApp, error)
	DeleteFocusModeAppsByModeID(ctx context.Context, modeID string) error

	CreateFocusModeURL(ctx context.Context, arg db.CreateFocusModeURLParams) (*db.FocusModeURL, error)
	GetFocusModeURLsByModeID(ctx context.Context, modeID string) ([]db.FocusModeURL, error)
	DeleteFocusModeURLsByModeID(ctx context.Context, modeID string) error
}

// FocusModeConfig holds all configurable safety constants.
type FocusModeConfig struct {
	WatchdogInterval     time.Duration
	ProxyHealthTimeout   time.Duration
	MaxConsecutiveKills  int
	KillLoopWindow       time.Duration
	EmergencyFailsafeMax int
	DefaultProxyPort     int
	BlockPollInterval    time.Duration
}

// DefaultFocusModeConfig returns the default configuration.
func DefaultFocusModeConfig() FocusModeConfig {
	return FocusModeConfig{
		WatchdogInterval:     5 * time.Second,
		ProxyHealthTimeout:   3 * time.Second,
		MaxConsecutiveKills:  10,
		KillLoopWindow:       30 * time.Second,
		EmergencyFailsafeMax: 3,
		DefaultProxyPort:     62828,
		BlockPollInterval:    2 * time.Second,
	}
}

// CancellableContext is satisfied by context.Context
type CancellableContext interface {
	Done() <-chan struct{}
	Err() error
}
