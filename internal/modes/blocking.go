package modes

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func BlockApps(apps []string) {
	if len(apps) == 0 {
		return
	}
	switch runtime.GOOS {
	case "linux":
		blockAppsLinux(apps)
	case "windows":
		blockAppsWindows(apps)
	}
}

func UnblockApps(apps []string) {
	if len(apps) == 0 {
		return
	}
	switch runtime.GOOS {
	case "linux":
		unblockAppsLinux(apps)
	case "windows":
		unblockAppsWindows(apps)
	}
}

func CloseApps(apps []string) {
	if len(apps) == 0 {
		return
	}
	switch runtime.GOOS {
	case "linux":
		closeAppsLinux(apps)
	case "windows":
		closeAppsWindows(apps)
	}
}

func blockAppsLinux(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("pkill", "-STOP", app)
		if err := cmd.Run(); err != nil {
			log.Printf("[modes] block %s (STOP): %v", app, err)
		}
	}
}

func unblockAppsLinux(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("pkill", "-CONT", app)
		if err := cmd.Run(); err != nil {
			log.Printf("[modes] unblock %s (CONT): %v", app, err)
		}
	}
}

func closeAppsLinux(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("pkill", "-TERM", app)
		cmd.Run()
	}
}

func blockAppsWindows(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("taskkill", "/F", "/IM", app+".exe")
		_ = cmd.Run()
	}
}

func unblockAppsWindows(_ []string) {
}

func closeAppsWindows(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("taskkill", "/F", "/IM", app+".exe")
		_ = cmd.Run()
	}
}

func isAppRunning(execName string) bool {
	switch runtime.GOOS {
	case "linux":
		return isProcessRunningLinux(execName)
	case "windows":
		return isProcessRunningWindows(execName)
	}
	return false
}

func isProcessRunningLinux(execName string) bool {
	cmd := exec.Command("pgrep", "-x", execName)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

func isProcessRunningWindows(execName string) bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+execName+".exe")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), execName+".exe")
}
