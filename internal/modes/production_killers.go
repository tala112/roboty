package modes

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const prodKillTimeout = 10 * time.Second

// realProcessKiller is the production implementation that uses taskkill/pkill.
type realProcessKiller struct{}

func NewRealProcessKiller() ProcessKiller {
	return &realProcessKiller{}
}

func (k *realProcessKiller) Kill(execName string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = prodKillTimeout
	}

	safeExec := NormalizeKillExec(execName)
	if safeExec == "" {
		return fmt.Errorf("rejected by NormalizeKillExec: %q", execName)
	}

	safe, reason := GetGlobalSafetyVerifier().IsSafeToKill(safeExec)
	if !safe {
		return fmt.Errorf("safety blocked: %s", reason)
	}

	if IsDevMode() {
		log.Printf("[dev] WOULD kill %s", safeExec)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		return k.killWindows(ctx, safeExec)
	case "linux", "darwin":
		return k.killUnix(ctx, safeExec)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (k *realProcessKiller) killWindows(ctx context.Context, safeExec string) error {
	cmd := exec.CommandContext(ctx, "taskkill", "/F", "/IM", safeExec+".exe")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill %s: %w - %s", safeExec, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (k *realProcessKiller) killUnix(ctx context.Context, safeExec string) error {
	cmd := exec.CommandContext(ctx, "pkill", "-TERM", safeExec)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pkill %s: %w - %s", safeExec, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (k *realProcessKiller) IsRunning(execName string) (bool, error) {
	safeExec := NormalizeKillExec(execName)
	if safeExec == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), prodKillTimeout)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		return k.isRunningWindows(ctx, safeExec)
	case "linux", "darwin":
		return k.isRunningUnix(ctx, safeExec)
	}
	return false, nil
}

func (k *realProcessKiller) isRunningWindows(ctx context.Context, safeExec string) (bool, error) {
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq "+safeExec+".exe")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), safeExec+".exe"), nil
}

func (k *realProcessKiller) isRunningUnix(ctx context.Context, safeExec string) (bool, error) {
	cmd := exec.CommandContext(ctx, "pgrep", "-x", safeExec)
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (k *realProcessKiller) ListRunning() ([]ProcessInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return listWindowsProcessesWithKiller()
	case "linux":
		return listLinuxProcessesWithKiller()
	case "darwin":
		return listMacOSProcessesWithKiller()
	}
	return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func listWindowsProcessesWithKiller() ([]ProcessInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), prodKillTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result []ProcessInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			name := strings.Trim(parts[0], "\"")
			execName := strings.TrimSuffix(name, ".exe")
			result = append(result, ProcessInfo{
				Name: name,
				Exec: strings.ToLower(execName),
			})
		}
	}
	return result, nil
}

func listLinuxProcessesWithKiller() ([]ProcessInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), prodKillTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-eo", "comm=")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result []ProcessInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, ProcessInfo{
			Name: line,
			Exec: strings.ToLower(line),
		})
	}
	return result, nil
}

func listMacOSProcessesWithKiller() ([]ProcessInfo, error) {
	return listLinuxProcessesWithKiller()
}
