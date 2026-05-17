//go:build darwin

package darwin

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func (p *DarwinPlatform) Mute() error {
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to true
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("macOS DND enable: %w", err)
	}
	log.Println("[notifications] macOS Do Not Disturb enabled")
	return nil
}

func (p *DarwinPlatform) Restore() error {
	script := `
tell application "System Events"
    tell expose preferences
        set do not disturb to false
    end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("macOS DND disable: %w", err)
	}
	log.Println("[notifications] macOS Do Not Disturb disabled")
	return nil
}

func (p *DarwinPlatform) IsMuted() (bool, error) {
	script := `
tell application "System Events"
    tell expose preferences
        get do not disturb
    end tell
end tell`
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "true", nil
}
