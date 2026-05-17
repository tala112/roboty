package modes

import (
	"log"
	"os"
	"path/filepath"
)

// fileStateManager persists state markers next to the executable for crash recovery.
type fileStateManager struct {
	baseDir string
}

func NewFileStateManager() StateFileManager {
	dir := ""
	exe, err := os.Executable()
	if err == nil {
		dir = filepath.Dir(exe)
	}
	return &fileStateManager{baseDir: dir}
}

func (m *fileStateManager) statePath(name string) string {
	if m.baseDir != "" {
		return filepath.Join(m.baseDir, "."+name)
	}
	return "." + name
}

func (m *fileStateManager) SaveState(name string) error {
	path := m.statePath(name)
	return os.WriteFile(path, []byte("1"), 0644)
}

func (m *fileStateManager) ClearState(name string) error {
	path := m.statePath(name)
	return os.Remove(path)
}

func (m *fileStateManager) StateExists(name string) bool {
	path := m.statePath(name)
	_, err := os.Stat(path)
	return err == nil
}

func (m *fileStateManager) ListOrphans() ([]string, error) {
	if m.baseDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, err
	}
	var orphans []string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 1 && e.Name()[0] == '.' {
			orphans = append(orphans, e.Name())
		}
	}
	return orphans, nil
}

func (m *fileStateManager) CleanupAll() error {
	orphans, err := m.ListOrphans()
	if err != nil {
		return err
	}
	for _, name := range orphans {
		if err := os.Remove(m.statePath(name)); err != nil {
			log.Printf("[state] failed to remove orphan %s: %v", name, err)
		}
	}
	return nil
}
