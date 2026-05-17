//go:build windows

package modes

import (
	"log"
	"os"
	"strings"
	"unsafe"
)

const (
	th32csSnapProcess = 0x00000002
	maxProcessPath    = 260
)

type PROCESSENTRY32W struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [maxProcessPath]uint16
}

var (
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
)

// GetAncestorExecs returns a set of executable names for all ancestor processes
// of the current process. Uses CreateToolhelp32Snapshot for efficient process tree walking.
func GetAncestorExecs() map[string]bool {
	execs := make(map[string]bool)
	pid := os.Getpid()
	maxDepth := 50

	// Build full process map: pid -> {parentPid, execName}
	type procInfo struct {
		parentPid int
		execName  string
	}
	processMap := make(map[int]procInfo)

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snapshot == uintptr(0xFFFFFFFF) || snapshot == 0 {
		execs["roboty"] = true
		execs["roboty.exe"] = true
		return execs
	}
	defer procCloseHandle.Call(snapshot)

	var pe PROCESSENTRY32W
	pe.dwSize = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		execs["roboty"] = true
		execs["roboty.exe"] = true
		return execs
	}

	for {
		exeName := strings.TrimSuffix(strings.ToLower(syscallUTF16ToString(pe.szExeFile[:])), ".exe")
		processMap[int(pe.th32ProcessID)] = procInfo{
			parentPid: int(pe.th32ParentProcessID),
			execName:  exeName,
		}

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	// Walk ancestor chain using the process map
	currentPid := pid
	for i := 0; i < maxDepth; i++ {
		if info, ok := processMap[currentPid]; ok {
			if info.execName != "" {
				execs[info.execName] = true
				execs[info.execName+".exe"] = true
			}
			ppid := info.parentPid
			if ppid <= 0 || ppid == currentPid {
				break
			}
			if execs[processMap[ppid].execName] {
				break
			}
			currentPid = ppid
		} else {
			break
		}
	}

	// Always protect our own process name
	execs["roboty"] = true
	execs["roboty.exe"] = true

	if len(execs) > 0 {
		names := make([]string, 0, len(execs))
		for e := range execs {
			names = append(names, e)
		}
		log.Printf("[proctree] ancestors (windows): %v", names)
	}
	return execs
}

// syscallUTF16ToString converts a UTF16 buffer to a Go string, stopping at the first null.
func syscallUTF16ToString(buf []uint16) string {
	var sb strings.Builder
	for _, v := range buf {
		if v == 0 {
			break
		}
		sb.WriteRune(rune(v))
	}
	return sb.String()
}
