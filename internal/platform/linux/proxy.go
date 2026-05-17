//go:build linux

package linux

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func (p *LinuxPlatform) Enable(proxyAddr string, port int) error {
	addr := fmt.Sprintf("%s:%d", proxyAddr, port)
	portStr := addr
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		portStr = addr[idx+1:]
	}
	hostStr := addr
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		hostStr = addr[:idx]
	}

	cmds := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "manual"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", hostStr},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", portStr},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", hostStr},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", portStr},
		{"gsettings", "set", "org.gnome.system.proxy", "ignore-hosts",
			"['localhost', '127.0.0.0/8', '::1', '10.0.0.0/8', '192.168.0.0/16', '169.254.0.0/16']"},
	}
	for _, args := range cmds {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			log.Printf("[proxy] linux cmd failed: %v", err)
			if strings.Contains(err.Error(), "gsettings") {
				return fmt.Errorf("gsettings not available: %w", err)
			}
		}
	}
	log.Println("[proxy] Linux system proxy enabled with localhost bypass")
	return nil
}

func (p *LinuxPlatform) Disable() error {
	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("disable linux proxy: %w", err)
	}
	return nil
}

func (p *LinuxPlatform) IsEnabled() (bool, error) {
	out, err := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "'manual'", nil
}
