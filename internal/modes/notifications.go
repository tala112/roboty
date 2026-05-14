package modes

import (
	"log"
	"os/exec"
	"runtime"
)

func MuteNotifications() {
	switch runtime.GOOS {
	case "linux":
		muteNotificationsLinux()
	case "windows":
		muteNotificationsWindows()
	}
}

func RestoreNotifications() {
	switch runtime.GOOS {
	case "linux":
		restoreNotificationsLinux()
	case "windows":
		restoreNotificationsWindows()
	}
}

func muteNotificationsLinux() {
	cmd := exec.Command("notify-send", "--help")
	if err := cmd.Run(); err != nil {
		log.Printf("[modes] notify-send not available, cannot mute via D-Bus")
		return
	}
	cmd = exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibit", "",
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[modes] D-Bus inhibit call failed: %v", err)
	}
	log.Println("[modes] Notifications muted (D-Bus Inhibit)")
}

func restoreNotificationsLinux() {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Uninhibit", "",
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[modes] D-Bus uninhibit call failed: %v", err)
	}
	log.Println("[modes] Notifications restored (D-Bus Uninhibit)")
}

func muteNotificationsWindows() {
	log.Println("[modes] Windows notification muting not yet implemented")
}

func restoreNotificationsWindows() {
	log.Println("[modes] Windows notification restore not yet implemented")
}
