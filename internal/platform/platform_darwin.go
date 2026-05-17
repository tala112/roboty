//go:build darwin

package platform

import "Roboty/internal/platform/darwin"

func init() {
	SetGlobal(darwin.NewPlatform())
}
