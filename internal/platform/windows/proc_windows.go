//go:build windows

package windows

import (
	"os"
	"strconv"
	"strings"
	"unsafe"

	syswin "golang.org/x/sys/windows"
)

var (
	kernel32                    = syswin.NewLazySystemDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procGetModuleBaseNameW       = kernel32.NewProc("K32GetModuleBaseNameW")
	procGetModuleFileNameExW     = kernel32.NewProc("K32GetModuleFileNameExW")
)

const (
	th32csSnapProcess       = 0x00000002
	processQueryInformation = 0x0400
	processVMRead           = 0x0010
	maxProcessPath          = 260
	maxPath                 = 260
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

var osGetpid = os.Getpid

func (p *WindowsPlatform) buildProcessMap() (map[int]procInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processMap != nil {
		return p.processMap, nil
	}

	processMap := make(map[int]procInfo)

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snapshot == uintptr(0xFFFFFFFF) || snapshot == 0 {
		return nil, fmtErr("CreateToolhelp32Snapshot failed")
	}
	defer procCloseHandle.Call(snapshot)

	var pe PROCESSENTRY32W
	pe.dwSize = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		procCloseHandle.Call(snapshot)
		return nil, fmtErr("Process32FirstW failed")
	}

	for {
		exeName := strings.TrimSuffix(strings.ToLower(utf16ToString(pe.szExeFile[:])), ".exe")
		processMap[int(pe.th32ProcessID)] = procInfo{
			parentPid: int(pe.th32ParentProcessID),
			execName:  exeName,
		}

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	p.processMap = processMap
	return processMap, nil
}

func (p *WindowsPlatform) getProcessInfo(pid int) (procInfo, error) {
	processMap, err := p.buildProcessMap()
	if err != nil {
		return procInfo{}, err
	}
	if info, ok := processMap[pid]; ok {
		return info, nil
	}
	return procInfo{}, fmtErr("process %d not found", pid)
}

func (p *WindowsPlatform) getAncestorExecsFallback(pid, maxDepth int) map[string]bool {
	execs := make(map[string]bool)
	execs["roboty"] = true
	execs["roboty.exe"] = true

	cmd := execCmd("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process | Select-Object ProcessId, ParentProcessId, Name | ConvertTo-Csv -NoTypeInformation`)
	out, err := cmd.Output()
	if err != nil {
		return execs
	}

	processMap := make(map[int]procInfo)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == `"ProcessId","ParentProcessId","Name"` {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		pidStr := strings.Trim(parts[0], `"`)
		ppidStr := strings.Trim(parts[1], `"`)
		name := strings.Trim(parts[2], `"`)

		pidNum, err1 := strconv.Atoi(pidStr)
		ppidNum, err2 := strconv.Atoi(ppidStr)
		if err1 != nil || err2 != nil {
			continue
		}
		execName := strings.ToLower(strings.TrimSuffix(name, ".exe"))
		processMap[pidNum] = procInfo{parentPid: ppidNum, execName: execName}
	}

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
			currentPid = ppid
		} else {
			break
		}
	}
	return execs
}

func utf16ToString(buf []uint16) string {
	var sb strings.Builder
	for _, v := range buf {
		if v == 0 {
			break
		}
		sb.WriteRune(rune(v))
	}
	return sb.String()
}
