// Package mockprocess provides mock process launchers for testing.
package mockprocess

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Launcher creates mock processes for testing.
type Launcher struct {
	processes []*exec.Cmd
}

func (l *Launcher) Start(name string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Start-Sleep -Seconds 3600; Start-Process '%s' -NoNewWindow", name))
	case "linux", "darwin":
		cmd = exec.Command("sleep", "3600")
	}
	if cmd == nil {
		return fmt.Errorf("unsupported platform")
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	l.processes = append(l.processes, cmd)
	return nil
}

func (l *Launcher) Cleanup() {
	for _, cmd := range l.processes {
		cmd.Process.Kill()
	}
	l.processes = nil
}
