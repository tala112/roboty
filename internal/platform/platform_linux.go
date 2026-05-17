//go:build linux

package platform

import "Roboty/internal/platform/linux"

func init() {
	SetGlobal(linux.NewPlatform())
}
