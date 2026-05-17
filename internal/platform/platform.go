package platform

import "Roboty/internal/platform/types"

type (
	ProcessInfo        = types.ProcessInfo
	ForegroundActivity = types.ForegroundActivity
	InstalledApp       = types.InstalledApp
)

type (
	ProcessManager       = types.ProcessManager
	ForegroundDetector   = types.ForegroundDetector
	ProxyManager         = types.ProxyManager
	NotificationManager  = types.NotificationManager
	ProcessTreeWalker    = types.ProcessTreeWalker
	AppDetector          = types.AppDetector
	Platform             = types.Platform
)

var globalPlatform Platform

func SetGlobal(p Platform) {
	globalPlatform = p
}

func GetGlobal() Platform {
	return globalPlatform
}
