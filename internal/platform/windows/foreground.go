//go:build windows

package windows

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"Roboty/internal/platform/types"

	syswin "golang.org/x/sys/windows"
)

var (
	user32                      = syswin.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow     = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW          = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

func (p *WindowsPlatform) Poll() (*types.ForegroundActivity, error) {
	execName, windowTitle, pid, err := getWindowsForegroundInfo()
	if err != nil || execName == "" {
		return nil, err
	}

	return &types.ForegroundActivity{
		AppName:     friendlyAppName(execName),
		ExecName:    strings.ToLower(strings.TrimSuffix(execName, ".exe")),
		WindowTitle: windowTitle,
		PID:         pid,
		Timestamp:   time.Now(),
	}, nil
}

func getWindowsForegroundInfo() (execName, windowTitle string, pid int, err error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", "", 0, fmt.Errorf("no foreground window")
	}

	titleBuf := make([]uint16, 512)
	ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), 512)
	if ret > 0 {
		windowTitle = syscall.UTF16ToString(titleBuf[:ret])
	}

	var processID uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&processID)))
	if processID == 0 {
		return "", windowTitle, 0, fmt.Errorf("no process ID")
	}

	handle, _, _ := procOpenProcess.Call(processQueryInformation|processVMRead, 0, uintptr(processID))
	if handle == 0 {
		return "", windowTitle, int(processID), nil
	}
	defer procCloseHandle.Call(handle)

	exeBuf := make([]uint16, maxPath)
	ret, _, _ = procGetModuleBaseNameW.Call(handle, 0, uintptr(unsafe.Pointer(&exeBuf[0])), maxPath)
	if ret > 0 {
		execName = syscall.UTF16ToString(exeBuf[:ret])
		log.Printf("[tracker] windows foreground: %s (pid=%d, title=%s)", execName, processID, windowTitle)
		return strings.TrimSuffix(execName, ".exe"), windowTitle, int(processID), nil
	}

	fullBuf := make([]uint16, maxPath)
	ret, _, _ = procGetModuleFileNameExW.Call(handle, 0, uintptr(unsafe.Pointer(&fullBuf[0])), maxPath)
	if ret > 0 {
		fullPath := syscall.UTF16ToString(fullBuf[:ret])
		parts := strings.Split(fullPath, "\\")
		if len(parts) > 0 {
			execName = strings.TrimSuffix(parts[len(parts)-1], ".exe")
		}
	}

	log.Printf("[tracker] windows foreground: %s (pid=%d, title=%s)", execName, processID, windowTitle)
	return execName, windowTitle, int(processID), nil
}

func friendlyAppName(execName string) string {
	known := map[string]string{
		"chrome":              "Google Chrome",
		"google chrome":       "Google Chrome",
		"firefox":             "Mozilla Firefox",
		"mozilla firefox":     "Mozilla Firefox",
		"msedge":              "Microsoft Edge",
		"microsoft edge":      "Microsoft Edge",
		"code":                "Visual Studio Code",
		"visual studio code":  "Visual Studio Code",
		"discord":             "Discord",
		"slack":               "Slack",
		"spotify":             "Spotify",
		"teams":               "Microsoft Teams",
		"microsoft teams":     "Microsoft Teams",
		"zoom":                "Zoom",
	}
	name := strings.ToLower(execName)
	if friendly, ok := known[name]; ok {
		return friendly
	}
	return execName
}
