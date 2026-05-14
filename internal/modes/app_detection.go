package modes

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetInstalledApps() ([]InstalledApp, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxApps()
	case "windows":
		return getWindowsApps()
	default:
		return []InstalledApp{}, nil
	}
}

func getLinuxApps() ([]InstalledApp, error) {
	paths := []string{
		"/usr/share/applications/",
		"/usr/local/share/applications/",
	}

	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths, filepath.Join(home, ".local/share/applications/"))
		snap := filepath.Join(home, "snap", "user")
		if dirExists(snap) {
			paths = append(paths, filepath.Join(snap, "current", "share", "applications"))
		}
	}

	seen := make(map[string]bool)
	var apps []InstalledApp

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}
			app, ok := parseDesktopFile(filepath.Join(dir, entry.Name()))
			if !ok || seen[app.Exec] {
				continue
			}
			seen[app.Exec] = true
			apps = append(apps, app)
		}
	}
	return apps, nil
}

func parseDesktopFile(path string) (InstalledApp, bool) {
	f, err := os.Open(path)
	if err != nil {
		return InstalledApp{}, false
	}
	defer f.Close()

	var name, execCmd string
	noDisplay := false
	inDesktop := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[Desktop Entry]" {
			inDesktop = true
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if inDesktop {
				break
			}
			continue
		}
		if !inDesktop {
			continue
		}

		if strings.HasPrefix(line, "Name=") && name == "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Exec=") && execCmd == "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				fullExec := strings.TrimSpace(parts[1])
				fullExec = strings.Split(fullExec, " ")[0]
				fullExec = strings.Trim(fullExec, "\"'")
				execCmd = filepath.Base(fullExec)
			}
		}
		if strings.Contains(line, "NoDisplay=true") {
			noDisplay = true
		}
		if strings.Contains(line, "Terminal=true") {
			return InstalledApp{}, false
		}
	}

	if name == "" || execCmd == "" || noDisplay {
		return InstalledApp{}, false
	}
	return InstalledApp{Name: name, Exec: execCmd}, true
}

func getWindowsApps() ([]InstalledApp, error) {
	var apps []InstalledApp
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
			apps = append(apps, InstalledApp{
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
