//go:build windows

package modes

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess             = kernel32.NewProc("OpenProcess")
	procGetModuleBaseNameW      = kernel32.NewProc("K32GetModuleBaseNameW")
	procCloseHandle             = kernel32.NewProc("CloseHandle")
	procGetModuleFileNameExW    = kernel32.NewProc("K32GetModuleFileNameExW")
)

const (
	processQueryInformation = 0x0400
	processVMRead          = 0x0010
	maxPath                = 260
)

func getWindowsForegroundInfo() (execName, windowTitle string, pid int, err error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", "", 0, fmt.Errorf("no foreground window")
	}

	// Get window title
	titleBuf := make([]uint16, 512)
	ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), 512)
	if ret > 0 {
		windowTitle = syscall.UTF16ToString(titleBuf[:ret])
	}

	// Get process ID
	var processID uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&processID)))
	if processID == 0 {
		return "", windowTitle, 0, fmt.Errorf("no process ID")
	}

	// Open process
	handle, _, _ := procOpenProcess.Call(processQueryInformation|processVMRead, 0, uintptr(processID))
	if handle == 0 {
		return "", windowTitle, int(processID), nil
	}
	defer procCloseHandle.Call(handle)

	// Try GetModuleBaseNameW first (just the exe name)
	exeBuf := make([]uint16, maxPath)
	ret, _, _ = procGetModuleBaseNameW.Call(handle, 0, uintptr(unsafe.Pointer(&exeBuf[0])), maxPath)
	if ret > 0 {
		execName = syscall.UTF16ToString(exeBuf[:ret])
		log.Printf("[tracker] windows foreground: %s (pid=%d, title=%s)", execName, processID, windowTitle)
		return strings.TrimSuffix(execName, ".exe"), windowTitle, int(processID), nil
	}

	// Fallback to GetModuleFileNameExW (full path)
	fullBuf := make([]uint16, maxPath)
	ret, _, _ = procGetModuleFileNameExW.Call(handle, 0, uintptr(unsafe.Pointer(&fullBuf[0])), maxPath)
	if ret > 0 {
		fullPath := syscall.UTF16ToString(fullBuf[:ret])
		// Extract just the executable name
		parts := strings.Split(fullPath, "\\")
		if len(parts) > 0 {
			execName = strings.TrimSuffix(parts[len(parts)-1], ".exe")
		}
	}

	log.Printf("[tracker] windows foreground: %s (pid=%d, title=%s)", execName, processID, windowTitle)
	return execName, windowTitle, int(processID), nil
}
