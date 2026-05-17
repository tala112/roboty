//go:build darwin

package darwin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"Roboty/internal/platform/types"
)

const killTimeout = 10 * time.Second

func (p *DarwinPlatform) Kill(execName string, timeout time.Duration) error {
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

func (p *DarwinPlatform) IsRunning(execName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", execName)
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (p *DarwinPlatform) ListRunning() ([]types.ProcessInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-eo", "comm")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []types.ProcessInfo

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		base := name
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			base = parts[len(parts)-1]
		}
		if seen[base] {
			continue
		}
		seen[base] = true
		result = append(result, types.ProcessInfo{
			Name: base,
			Exec: strings.ToLower(base),
		})
	}
	return result, nil
}

func (p *DarwinPlatform) GetProcessName(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-o", "comm=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *DarwinPlatform) GetParentPID(pid int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), killTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

func (p *DarwinPlatform) GetAncestorExecs() (map[string]bool, error) {
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
