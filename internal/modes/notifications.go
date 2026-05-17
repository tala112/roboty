package modes

import (
	"log"
)

func MuteNotifications() {
	if IsDevMode() {
		log.Println("[dev] WOULD mute notifications")
		return
	}
	saveNotificationState()
	mgr := NewRealNotificationManager()
	if err := mgr.Mute(); err != nil {
		log.Printf("[notifications] mute failed: %v", err)
	}
}

func RestoreNotifications() {
	if IsDevMode() {
		log.Println("[dev] WOULD restore notifications")
		return
	}
	clearNotificationState()
	mgr := NewRealNotificationManager()
	if err := mgr.Restore(); err != nil {
		log.Printf("[notifications] restore failed: %v", err)
	}
}
