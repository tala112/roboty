package modes

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func CloseApps(apps []string) {
	if len(apps) == 0 {
		return
	}
	switch runtime.GOOS {
	case "linux":
		closeAppsLinux(apps)
	case "windows":
		closeAppsWindows(apps)
	case "darwin":
		closeAppsMacOS(apps)
	}
}

func CloseApp(appExec string) {
	CloseApps([]string{appExec})
}

func closeAppsLinux(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("pkill", "-TERM", app)
		if err := cmd.Run(); err != nil {
			log.Printf("[blocking] close %s: %v", app, err)
		}
	}
}

func closeAppsWindows(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("taskkill", "/F", "/IM", app+".exe")
		if err := cmd.Run(); err != nil {
			log.Printf("[blocking] close %s: %v", app, err)
		}
	}
}

func closeAppsMacOS(apps []string) {
	for _, app := range apps {
		cmd := exec.Command("pkill", "-TERM", app)
		if err := cmd.Run(); err != nil {
			log.Printf("[blocking] close %s: %v", app, err)
		}
	}
}

func isAppRunning(execName string) bool {
	switch runtime.GOOS {
	case "linux":
		return isProcessRunningLinux(execName)
	case "windows":
		return isProcessRunningWindows(execName)
	case "darwin":
		return isProcessRunningMacOS(execName)
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

func isProcessRunningMacOS(execName string) bool {
	cmd := exec.Command("pgrep", "-x", execName)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}
