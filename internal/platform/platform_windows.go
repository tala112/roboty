//go:build windows

package platform

import "Roboty/internal/platform/windows"

func init() {
	SetGlobal(windows.NewPlatform())
}
