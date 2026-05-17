//go:build windows

package windows

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	syswin "golang.org/x/sys/windows"
)

var (
	wininet                = syswin.NewLazyDLL("wininet.dll")
	procInternetSetOptionW = wininet.NewProc("InternetSetOptionW")
)

func notifyProxyChange() {
	procInternetSetOptionW.Call(0, 39, 0, 0)
	procInternetSetOptionW.Call(0, 37, 0, 0)
}

func (p *WindowsPlatform) Enable(proxyAddr string, port int) error {
	addr := fmt.Sprintf("%s:%d", proxyAddr, port)
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

func (p *WindowsPlatform) Disable() error {
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -ErrorAction SilentlyContinue; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyBypass" -ErrorAction SilentlyContinue`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy disable: %w - %s", err, string(out))
	}
	notifyProxyChange()
	log.Println("[proxy] Windows system proxy disabled and ProxyServer cleared")
	return nil
}

func (p *WindowsPlatform) IsEnabled() (bool, error) {
	psCmd := `(Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable").ProxyEnable`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "1", nil
}
