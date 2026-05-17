//go:build linux

package linux

import (
	"fmt"
	"log"
	"os/exec"
)

func (p *LinuxPlatform) Mute() error {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibit", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("D-Bus inhibit: %w", err)
	}
	log.Println("[notifications] muted (D-Bus Inhibit)")
	return nil
}

func (p *LinuxPlatform) Restore() error {
	cmd := exec.Command("busctl", "--user", "call",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Uninhibit", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("D-Bus uninhibit: %w", err)
	}
	log.Println("[notifications] restored (D-Bus Uninhibit)")
	return nil
}

func (p *LinuxPlatform) IsMuted() (bool, error) {
	out, err := exec.Command("busctl", "--user", "get-property",
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications",
		"Inhibited").Output()
	if err != nil {
		return false, err
	}
	return string(out) == "true\n", nil
}
