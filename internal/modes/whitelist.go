package modes

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetWhitelistExecs reads whitelist.json and returns all executable/process names
// that must never be blocked or shown to the user.
func GetWhitelistExecs() map[string]bool {
	paths := []string{
		"internal/modes/whitelist.json",
		"whitelist.json",
	}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
		// Also try relative to executable
		abs, _ := filepath.Abs(p)
		data, err = os.ReadFile(abs)
		if err == nil {
			break
		}
	}
	if data == nil {
		log.Printf("[whitelist] failed to read from any path: %v", err)
		return map[string]bool{}
	}
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("[whitelist] failed to parse: %v", err)
		return map[string]bool{}
	}

	set := make(map[string]bool)
	collectExecs(raw, &set)
	return set
}

func collectExecs(v interface{}, set *map[string]bool) {
	switch val := v.(type) {
	case string:
		(*set)[strings.ToLower(val)] = true
	case []interface{}:
		for _, item := range val {
			collectExecs(item, set)
		}
	case map[string]interface{}:
		for _, child := range val {
			collectExecs(child, set)
		}
	}
}
