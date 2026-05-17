package modes

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

var (
	wininet                = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOptionW = wininet.NewProc("InternetSetOptionW")
)

func notifyProxyChange() {
	procInternetSetOptionW.Call(0, 39, 0, 0) // INTERNET_OPTION_SETTINGS_CHANGED
	procInternetSetOptionW.Call(0, 37, 0, 0) // INTERNET_OPTION_REFRESH
}

// realProxyManager is the production implementation for OS proxy settings.
type realProxyManager struct{}

func NewRealProxyManager() SystemProxyManager {
	return &realProxyManager{}
}

func (m *realProxyManager) Enable(proxyAddr string, port int) error {
	addr := fmt.Sprintf("%s:%d", proxyAddr, port)

	switch runtime.GOOS {
	case "linux":
		return m.enableLinux(addr)
	case "darwin":
		return m.enableMacOS(addr)
	case "windows":
		return m.enableWindows(addr)
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (m *realProxyManager) Disable() error {
	switch runtime.GOOS {
	case "linux":
		return m.disableLinux()
	case "darwin":
		return m.disableMacOS()
	case "windows":
		return m.disableWindows()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (m *realProxyManager) IsEnabled() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		return m.isEnabledWindows()
	case "linux":
		return m.isEnabledLinux()
	case "darwin":
		return m.isEnabledMacOS()
	}
	return false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func (m *realProxyManager) enableLinux(addr string) error {
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
		}
	}
	log.Println("[proxy] Linux system proxy enabled with localhost bypass")
	return nil
}

func (m *realProxyManager) disableLinux() error {
	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("disable linux proxy: %w", err)
	}
	return nil
}

func (m *realProxyManager) isEnabledLinux() (bool, error) {
	out, err := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "'manual'", nil
}

func (m *realProxyManager) enableMacOS(addr string) error {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return fmt.Errorf("list network services: %w", err)
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

func (m *realProxyManager) disableMacOS() error {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return err
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

func (m *realProxyManager) isEnabledMacOS() (bool, error) {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return false, err
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

func (m *realProxyManager) enableWindows(addr string) error {
	psCmd := fmt.Sprintf(
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 1 -Type DWord -Force; `+
			`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -Value "%s" -Type String -Force; `+
			`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyBypass" -Value @("localhost","127.*","10.*","192.168.*","169.254.*","::1","<local>") -Type MultiString -Force`,
		addr)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy enable: %w - %s", err, string(out))
	}
	notifyProxyChange()
	log.Println("[proxy] Windows system proxy enabled with localhost bypass")
	return nil
}

func (m *realProxyManager) disableWindows() error {
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy disable: %w - %s", err, string(out))
	}
	notifyProxyChange()
	log.Println("[proxy] Windows system proxy disabled")
	return nil
}

func (m *realProxyManager) isEnabledWindows() (bool, error) {
	psCmd := `(Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable").ProxyEnable`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "1", nil
}
