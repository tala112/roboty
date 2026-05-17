//go:build windows

package windows

import (
	"log"
	"os/exec"
	"strings"
)

func (p *WindowsPlatform) Mute() error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 0 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast disable failed: %v", err)
	}

	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		`New-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications" -Name "QuietHours" -Force | Out-Null; Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 1 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist enable failed: %v", err)
	}
	log.Println("[notifications] Windows notifications muted")
	return nil
}

func (p *WindowsPlatform) Restore() error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows toast restore failed: %v", err)
	}

	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\QuietHours" -Name "Enabled" -Value 0 -Type DWord -Force`)
	if err := cmd.Run(); err != nil {
		log.Printf("[notifications] Windows Focus Assist disable failed: %v", err)
	}
	log.Println("[notifications] Windows notifications restored")
	return nil
}

func (p *WindowsPlatform) IsMuted() (bool, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED").NOC_GLOBAL_SETTING_TOASTS_ENABLED`)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "0", nil
}
