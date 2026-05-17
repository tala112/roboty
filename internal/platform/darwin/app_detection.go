//go:build darwin

package darwin

import (
	"os"
	"path/filepath"
	"strings"

	"Roboty/internal/platform/types"
)

func (p *DarwinPlatform) GetInstalledApps() ([]types.InstalledApp, error) {
	// Look in common macOS application directories
	paths := []string{
		"/Applications",
		"/Applications/Utilities",
	}

	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths,
			filepath.Join(home, "Applications"),
		)
	}

	seen := make(map[string]bool)
	var apps []types.InstalledApp

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".app") {
				continue
			}
			appName := strings.TrimSuffix(name, ".app")
			execName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))
			if seen[execName] {
				continue
			}
			seen[execName] = true
			apps = append(apps, types.InstalledApp{
				Name: appName,
				Exec: execName,
			})
		}
	}
	return apps, nil
}
