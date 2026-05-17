//go:build windows

package windows

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

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
	// Notify before making changes — flushes any cached proxy state in running apps
	notifyProxyChange()

	// 1. Clear HKCU proxy values and the primary connection BLOBs that Windows
	//    uses to restore proxy settings on network refresh / sleep-wake / DHCP renew.
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -ErrorAction SilentlyContinue; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyBypass" -ErrorAction SilentlyContinue; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings\Connections" -Name "DefaultConnectionSettings" -ErrorAction SilentlyContinue; ` +
		`Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings\Connections" -Name "SavedLegacySettings" -ErrorAction SilentlyContinue`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy disable: %w - %s", err, string(out))
	}

	// 2. Clear WinHTTP proxy (separate store used by Windows Update, system services).
	//    Requires admin — failure is logged but non-fatal.
	if err := exec.Command("netsh", "winhttp", "reset", "proxy").Run(); err != nil {
		log.Printf("[proxy] netsh winhttp reset non-fatal: %v", err)
	}

	// Brief delay to let registry changes propagate before notification
	time.Sleep(200 * time.Millisecond)
	notifyProxyChange()

	// 3. Verify the proxy is actually disabled
	enabled, err := p.IsEnabled()
	if err != nil {
		log.Printf("[proxy] Disable verification error (non-fatal): %v", err)
	} else if enabled {
		log.Println("[proxy] ⚠️ Proxy still shows enabled after primary disable — retrying")
		retryCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force`
		if retryOut, retryErr := exec.Command("powershell", "-NoProfile", "-Command", retryCmd).CombinedOutput(); retryErr != nil {
			return fmt.Errorf("windows proxy disable retry: %w - %s", retryErr, string(retryOut))
		}
		time.Sleep(200 * time.Millisecond)
		notifyProxyChange()
		if enabled2, _ := p.IsEnabled(); enabled2 {
			return fmt.Errorf("proxy remains enabled after retry")
		}
	}

	log.Println("[proxy] Windows proxy fully disabled (registry, BLOBs, WinHTTP cleared)")
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
