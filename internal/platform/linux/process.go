//go:build linux

package linux

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Roboty/internal/platform/types"
)

const killTimeout = 10 * time.Second

func (p *LinuxPlatform) Kill(execName string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = killTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pkill", "-TERM", execName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pkill %s: %w - %s", execName, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (p *LinuxPlatform) IsRunning(execName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", execName)
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (p *LinuxPlatform) ListRunning() ([]types.ProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []types.ProcessInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid < 100 {
			continue
		}

		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, types.ProcessInfo{
			Name: name,
			Exec: name,
		})
	}
	return result, nil
}

func (p *LinuxPlatform) GetProcessName(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (p *LinuxPlatform) GetParentPID(pid int) (int, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PPid:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return strconv.Atoi(parts[1])
			}
		}
	}
	return 0, fmt.Errorf("no PPid found in /proc/%d/status", pid)
}

func (p *LinuxPlatform) GetAncestorExecs() (map[string]bool, error) {
	execs := make(map[string]bool)
	pid := os.Getpid()
	maxDepth := 50

	for i := 0; i < maxDepth; i++ {
		execName, err := p.GetProcessName(pid)
		if err == nil && execName != "" {
			execs[strings.ToLower(execName)] = true
		}

		ppid, err := p.GetParentPID(pid)
		if err != nil || ppid <= 0 || ppid == pid {
			break
		}
		pid = ppid
	}

	if len(execs) > 0 {
		names := make([]string, 0, len(execs))
		for e := range execs {
			names = append(names, e)
		}
		log.Printf("[proctree] ancestors: %v", names)
	}
	return execs, nil
}
