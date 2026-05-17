//go:build !windows

package modes

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetAncestorExecs returns a set of executable names for all ancestor processes
// of the current process. This prevents Roboty from accidentally blocking
// its own parent/ancestor processes (terminal, launcher, dev tools, etc.).
func GetAncestorExecs() map[string]bool {
	execs := make(map[string]bool)
	pid := os.Getpid()
	maxDepth := 50

	for i := 0; i < maxDepth; i++ {
		execName, err := getProcessName(pid)
		if err == nil && execName != "" {
			execs[strings.ToLower(execName)] = true
		}

		ppid, err := getParentPID(pid)
		if err != nil || ppid <= 0 || ppid == pid {
			break
		}
		if pid == ppid {
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
	return execs
}

func getProcessName(pid int) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return getProcessNameLinux(pid)
	case "darwin":
		return getProcessNameMacOS(pid)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getProcessNameLinux(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getProcessNameMacOS(pid int) (string, error) {
	cmd := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getParentPID(pid int) (int, error) {
	switch runtime.GOOS {
	case "linux":
		return getParentPIDLinux(pid)
	case "darwin":
		return getParentPIDMacOS(pid)
	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getParentPIDLinux(pid int) (int, error) {
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

func getParentPIDMacOS(pid int) (int, error) {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}
