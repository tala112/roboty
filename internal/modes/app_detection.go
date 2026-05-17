package modes

import (
	"Roboty/internal/platform"
)

func GetInstalledApps() ([]InstalledApp, error) {
	p := platform.GetGlobal()
	if p == nil {
		return []InstalledApp{}, nil
	}
	apps, err := p.GetInstalledApps()
	if err != nil {
		return nil, err
	}
	result := make([]InstalledApp, len(apps))
	for i, a := range apps {
		result[i] = InstalledApp{Name: a.Name, Exec: a.Exec}
	}
	return result, nil
}