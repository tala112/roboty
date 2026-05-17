package modes

import (
	"strings"

	"Roboty/internal/platform"
)

// GetAncestorExecs returns the set of ancestor process exec names.
// Delegates to the platform if available.
func GetAncestorExecs() map[string]bool {
	if p := platform.GetGlobal(); p != nil {
		execs, _ := p.GetAncestorExecs()
		return execs
	}
	return nil
}

// listWindowsProcesses lists running processes using the platform.
func listWindowsProcesses() ([]InstalledApp, error) {
	killer := NewRealProcessKiller()
	procs, err := killer.ListRunning()
	if err != nil {
		return nil, err
	}
	apps := make([]InstalledApp, 0, len(procs))
	seen := make(map[string]bool)
	for _, p := range procs {
		if seen[p.Exec] {
			continue
		}
		seen[p.Exec] = true
		apps = append(apps, InstalledApp{Name: p.Name, Exec: p.Exec})
	}
	return apps, nil
}

// friendlyAppName translates common process names to user-friendly names.
func (ft *ForegroundTracker) friendlyAppName(execName string) string {
	known := map[string]string{
		"chrome":              "Google Chrome",
		"google chrome":       "Google Chrome",
		"firefox":             "Mozilla Firefox",
		"mozilla firefox":     "Mozilla Firefox",
		"msedge":              "Microsoft Edge",
		"microsoft edge":      "Microsoft Edge",
		"code":                "Visual Studio Code",
		"visual studio code":  "Visual Studio Code",
		"discord":             "Discord",
		"slack":               "Slack",
		"spotify":             "Spotify",
		"teams":               "Microsoft Teams",
		"microsoft teams":     "Microsoft Teams",
		"zoom":                "Zoom",
	}
	name := strings.ToLower(execName)
	if friendly, ok := known[name]; ok {
		return friendly
	}
	return execName
}
