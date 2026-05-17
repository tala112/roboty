//go:build windows

package windows

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
	"os"

	"Roboty/internal/platform/types"
)

const killTimeout = 10 * time.Second

func (p *WindowsPlatform) Kill(execName string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = killTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "taskkill", "/F", "/IM", execName+".exe")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill %s: %w - %s", execName, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (p *WindowsPlatform) IsRunning(execName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq "+execName+".exe")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), execName+".exe"), nil
}

func (p *WindowsPlatform) ListRunning() ([]types.ProcessInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result []types.ProcessInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			name := strings.Trim(parts[0], "\"")
			execName := strings.TrimSuffix(name, ".exe")
			result = append(result, types.ProcessInfo{
				Name: name,
				Exec: strings.ToLower(execName),
			})
		}
	}
	return result, nil
}

func (p *WindowsPlatform) GetProcessName(pid int) (string, error) {
	info, err := p.getProcessInfo(pid)
	if err != nil {
		return "", err
	}
	return info.execName, nil
}

func (p *WindowsPlatform) GetParentPID(pid int) (int, error) {
	processMap, err := p.buildProcessMap()
	if err != nil {
		return 0, err
	}
	if info, ok := processMap[pid]; ok {
		return info.parentPid, nil
	}
	return 0, fmt.Errorf("process %d not found", pid)
}

func (p *WindowsPlatform) GetAncestorExecs() (map[string]bool, error) {
	execs := make(map[string]bool)
	pid := os.Getpid()

	processMap, err := p.buildProcessMap()
	if err != nil {
		return p.getAncestorExecsFallback(pid, 50), nil
	}

	currentPid := pid
	for i := 0; i < 50; i++ {
		if info, ok := processMap[currentPid]; ok {
			if info.execName != "" {
				execs[info.execName] = true
				execs[info.execName+".exe"] = true
			}
			ppid := info.parentPid
			if ppid <= 0 || ppid == currentPid {
				break
			}
			currentPid = ppid
		} else {
			break
		}
	}

	execs["roboty"] = true
	execs["roboty.exe"] = true

	if len(execs) > 0 {
		names := make([]string, 0, len(execs))
		for e := range execs {
			names = append(names, e)
		}
		log.Printf("[proctree] ancestors (windows): %v", names)
	}
	return execs, nil
}
