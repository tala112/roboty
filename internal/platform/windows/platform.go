//go:build windows

package windows

import (
	"os"
	"sync"

	"Roboty/internal/platform/types"
)

type procInfo struct {
	parentPid int
	execName  string
}

type WindowsPlatform struct {
	mu         sync.Mutex
	processMap map[int]procInfo
	currentPID int
}

func NewPlatform() types.Platform {
	return &WindowsPlatform{
		currentPID: os.Getpid(),
	}
}

var _ types.Platform = (*WindowsPlatform)(nil)
