//go:build darwin

package darwin

import (
	"Roboty/internal/platform/types"
)

type DarwinPlatform struct{}

func NewPlatform() types.Platform {
	return &DarwinPlatform{}
}

var _ types.Platform = (*DarwinPlatform)(nil)
