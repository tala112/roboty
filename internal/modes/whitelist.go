package modes

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// GetWhitelistExecs reads whitelist.json and returns all executable/process names
// that must never be blocked or shown to the user.
func GetWhitelistExecs() map[string]bool {
	data, err := os.ReadFile("internal/modes/whitelist.json")
	if err != nil {
		log.Printf("[whitelist] failed to read: %v", err)
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
