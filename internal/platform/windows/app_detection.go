//go:build windows

package windows

import (
	"os"
	"path/filepath"
	"strings"

	"Roboty/internal/platform/types"
)

func (p *WindowsPlatform) GetInstalledApps() ([]types.InstalledApp, error) {
	var apps []types.InstalledApp
	startMenuDirs := []string{
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"),
	}
	seen := make(map[string]bool)
	for _, dir := range startMenuDirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".lnk" {
				return nil
			}
			appName := strings.TrimSuffix(d.Name(), ".lnk")
			appExec := strings.ToLower(appName)
			if seen[appExec] {
				return nil
			}
			seen[appExec] = true
			apps = append(apps, types.InstalledApp{
				Name: appName,
				Exec: appExec,
			})
			return nil
		})
		if err != nil {
			continue
		}
	}
	return apps, nil
}
