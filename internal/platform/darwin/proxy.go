//go:build darwin

package darwin

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func (p *DarwinPlatform) Enable(proxyAddr string, port int) error {
	addr := fmt.Sprintf("%s:%d", proxyAddr, port)
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		log.Printf("[proxy] macos enable non-fatal (networksetup unavailable): %v", err)
		return nil
	}

	services := strings.Split(string(netService), "\n")
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" || strings.Contains(s, "An asterisk") {
			continue
		}
		hostStr := addr
		portStr := addr
		if idx := strings.LastIndex(addr, ":"); idx >= 0 {
			hostStr = addr[:idx]
			portStr = addr[idx+1:]
		}
		exec.Command("networksetup", "-setwebproxy", s, hostStr, portStr).Run()
		exec.Command("networksetup", "-setsecurewebproxy", s, hostStr, portStr).Run()
		exec.Command("networksetup", "-setwebproxystate", s, "on").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", s, "on").Run()
		exec.Command("networksetup", "-setproxybypassdomains", s,
			"localhost", "127.0.0.1", "::1", "10.0.0.0/8", "192.168.0.0/16").Run()
	}
	log.Println("[proxy] macOS system proxy enabled with localhost bypass")
	return nil
}

func (p *DarwinPlatform) Disable() error {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		log.Printf("[proxy] macos disable non-fatal: %v", err)
		return nil
	}

	services := strings.Split(string(netService), "\n")
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" || strings.Contains(s, "An asterisk") {
			continue
		}
		exec.Command("networksetup", "-setwebproxystate", s, "off").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", s, "off").Run()
		exec.Command("networksetup", "-setproxybypassdomains", s, "Empty").Run()
	}
	log.Println("[proxy] macOS system proxy disabled")
	return nil
}

func (p *DarwinPlatform) IsEnabled() (bool, error) {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		log.Printf("[proxy] macos IsEnabled non-fatal: %v", err)
		return false, nil
	}

	services := strings.Split(string(netService), "\n")
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out, err := exec.Command("networksetup", "-getwebproxy", s).Output()
		if err != nil {
			continue
		}
		if strings.Contains(string(out), "Enabled: Yes") {
			return true, nil
		}
	}
	return false, nil
}
