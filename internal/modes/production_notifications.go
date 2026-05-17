package modes

import (
	"Roboty/internal/platform"
)

type realNotificationManager struct{}

func NewRealNotificationManager() NotificationManager {
	return &realNotificationManager{}
}

func (m *realNotificationManager) Mute() error {
	if p := platform.GetGlobal(); p != nil {
		return p.Mute()
	}
	return nil
}

func (m *realNotificationManager) Restore() error {
	if p := platform.GetGlobal(); p != nil {
		return p.Restore()
	}
	return nil
}

func (m *realNotificationManager) IsMuted() (bool, error) {
	if p := platform.GetGlobal(); p != nil {
		return p.IsMuted()
	}
	return false, nil
}
