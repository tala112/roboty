package modes

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type AppMappingCategory struct {
	Category string   `json:"category"`
	Apps     []string `json:"apps"`
}

type AppMappingsFile struct {
	Mappings []AppMappingCategory `json:"mappings"`
}

const appMappingsPath = "internal/modes/app-mappings.json"

func resolveMappingsPath() (string, error) {
	paths := []string{
		appMappingsPath,
	}

	exe, err := os.Executable()
	if err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), appMappingsPath))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("app-mappings.json not found in any known path")
}

// GetAppsFromMappings reads app-mappings.json and returns all known apps
func GetAppsFromMappings() ([]InstalledApp, error) {
	absPath, err := resolveMappingsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read app-mappings: %w", err)
	}

	var mappings AppMappingsFile
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("parse app-mappings: %w", err)
	}

	seen := make(map[string]bool)
	var apps []InstalledApp

	for _, cat := range mappings.Mappings {
		for _, entry := range cat.Apps {
			// Entry format: "Display Name|variant1|variant2|..."
			parts := strings.Split(entry, "|")
			if len(parts) == 0 || parts[0] == "" {
				continue
			}

			displayName := strings.TrimSpace(parts[0])
			// Use the last part as exec name (most specific), or lowercased display name
			execName := strings.TrimSpace(parts[len(parts)-1])
			if execName == "" {
				execName = strings.ToLower(displayName)
			}
			execKey := strings.ToLower(execName)

			if seen[execKey] {
				continue
			}
			seen[execKey] = true

			apps = append(apps, InstalledApp{
				Name: displayName,
				Exec: execKey,
			})
		}
	}

	return apps, nil
}

func addAppToMappingsFile(appName, appExec, category string) error {
	absPath, err := resolveMappingsPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read app-mappings: %w", err)
	}

	var mappings AppMappingsFile
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("parse app-mappings: %w", err)
	}

	// Create pipe-delimited entry: "AppName|appexec"
	entry := fmt.Sprintf("%s|%s", appName, appExec)

	// Find or create the category
	found := false
	for i := range mappings.Mappings {
		if strings.EqualFold(mappings.Mappings[i].Category, category) {
			// Check if already exists (by exec)
			for _, existing := range mappings.Mappings[i].Apps {
				existingExec := extractExecFromEntry(existing)
				if strings.EqualFold(existingExec, appExec) {
					found = true
					break
				}
			}
			if !found {
				mappings.Mappings[i].Apps = append(mappings.Mappings[i].Apps, entry)
				found = true
			}
			break
		}
	}

	if !found {
		// Create new category
		mappings.Mappings = append(mappings.Mappings, AppMappingCategory{
			Category: category,
			Apps:     []string{entry},
		})
	}

	updated, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal app-mappings: %w", err)
	}

	if err := os.WriteFile(absPath, updated, 0644); err != nil {
		return fmt.Errorf("write app-mappings: %w", err)
	}

	log.Printf("[mappings] Added %s to category %s", entry, category)
	return nil
}

// checkAppInMappings checks if an app exec is already in app-mappings.json
func checkAppInMappings(appExec string) bool {
	apps, err := GetAppsFromMappings()
	if err != nil {
		return false
	}
	key := strings.ToLower(appExec)
	for _, a := range apps {
		if strings.ToLower(a.Exec) == key {
			return true
		}
	}
	return false
}

// extractExecFromEntry extracts the exec name from a pipe-delimited entry
// e.g., "Visual Studio Code|Code|vscode" -> "vscode"
func extractExecFromEntry(entry string) string {
	parts := strings.Split(entry, "|")
	last := strings.TrimSpace(parts[len(parts)-1])
	if last == "" {
		return strings.ToLower(strings.TrimSpace(parts[0]))
	}
	return strings.ToLower(last)
}
