package modes

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type whitelistData struct {
	Apps []string `json:"apps"`
}

func GetWhitelistExecs() map[string]bool {
	data, err := os.ReadFile("internal/modes/whitelist.json")
	if err != nil {
		log.Printf("[whitelist] failed to read: %v", err)
		return map[string]bool{}
	}
	var wl whitelistData
	if err := json.Unmarshal(data, &wl); err != nil {
		log.Printf("[whitelist] failed to parse: %v", err)
		return map[string]bool{}
	}
	set := make(map[string]bool, len(wl.Apps))
	for _, a := range wl.Apps {
		set[strings.ToLower(a)] = true
	}
	return set
}
