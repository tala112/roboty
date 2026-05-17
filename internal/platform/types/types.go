package types

import "time"

type ProcessManager interface {
	Kill(execName string, timeout time.Duration) error
	IsRunning(execName string) (bool, error)
	ListRunning() ([]ProcessInfo, error)
}

type ProcessInfo struct {
	Name string
	Exec string
	PID  int
}

type ForegroundDetector interface {
	Poll() (*ForegroundActivity, error)
}

type ForegroundActivity struct {
	AppName     string    `json:"app_name"`
	ExecName    string    `json:"exec_name"`
	WindowTitle string    `json:"window_title"`
	PID         int       `json:"pid"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProxyManager interface {
	Enable(proxyAddr string, port int) error
	Disable() error
	IsEnabled() (bool, error)
}

type NotificationManager interface {
	Mute() error
	Restore() error
	IsMuted() (bool, error)
}

type ProcessTreeWalker interface {
	GetAncestorExecs() (map[string]bool, error)
	GetParentPID(pid int) (int, error)
	GetProcessName(pid int) (string, error)
}

type InstalledApp struct {
	Name string `json:"name"`
	Exec string `json:"exec"`
	Icon string `json:"icon,omitempty"`
}

type AppDetector interface {
	GetInstalledApps() ([]InstalledApp, error)
}

type Platform interface {
	ProcessManager
	ForegroundDetector
	ProxyManager
	NotificationManager
	ProcessTreeWalker
	AppDetector
}
