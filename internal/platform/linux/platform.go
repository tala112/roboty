//go:build linux

package linux

import (
	"Roboty/internal/platform/types"
)

type LinuxPlatform struct{}

func NewPlatform() types.Platform {
	return &LinuxPlatform{}
}

var _ types.Platform = (*LinuxPlatform)(nil)
