package modes

import (
	"context"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const killTimeout = 10 * time.Second

var globalBlockingVerifier = GetGlobalSafetyVerifier()

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

// safeExecName validates and normalizes a process name for killing.
// Returns empty string if the process must NOT be killed.
func safeExecName(app string) string {
	safeExec := NormalizeKillExec(app)
	if safeExec == "" {
		log.Printf("[blocking] SKIP kill of %q: rejected by NormalizeKillExec", app)
		return ""
	}
	safe, reason := globalBlockingVerifier.IsSafeToKill(safeExec)
	if !safe {
		log.Printf("[blocking] SAFETY BLOCKED kill of %q: %s", app, reason)
		return ""
	}
	return safeExec
}

func closeAppsLinux(apps []string) {
	for _, app := range apps {
		safeExec := safeExecName(app)
		if safeExec == "" {
			continue
		}
		if IsDevMode() {
			log.Printf("[dev] WOULD pkill -TERM %s", safeExec)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
		cmd := exec.CommandContext(ctx, "pkill", "-TERM", safeExec)
		err := cmd.Run()
		cancel()
		if err != nil {
			log.Printf("[blocking] close %s: %v", safeExec, err)
		}
	}
}

func closeAppsWindows(apps []string) {
	for _, app := range apps {
		safeExec := safeExecName(app)
		if safeExec == "" {
			continue
		}
		if IsDevMode() {
			log.Printf("[dev] WOULD taskkill /F /IM %s.exe", safeExec)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
		cmd := exec.CommandContext(ctx, "taskkill", "/F", "/IM", safeExec+".exe")
		err := cmd.Run()
		cancel()
		if err != nil {
			log.Printf("[blocking] close %s: %v", safeExec, err)
		}
	}
}

func closeAppsMacOS(apps []string) {
	for _, app := range apps {
		safeExec := safeExecName(app)
		if safeExec == "" {
			continue
		}
		if IsDevMode() {
			log.Printf("[dev] WOULD pkill -TERM %s", safeExec)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
		cmd := exec.CommandContext(ctx, "pkill", "-TERM", safeExec)
		err := cmd.Run()
		cancel()
		if err != nil {
			log.Printf("[blocking] close %s: %v", safeExec, err)
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
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", execName)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

func isProcessRunningWindows(execName string) bool {
	safeExec := NormalizeKillExec(execName)
	if safeExec == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq "+safeExec+".exe")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), safeExec+".exe")
}

func isProcessRunningMacOS(execName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", execName)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}
