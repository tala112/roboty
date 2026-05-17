package modes

import (
	"strings"
	"testing"
)

// FuzzNormalizeKillExec tests edge cases and attack vectors for process name normalization.
func FuzzNormalizeKillExec(f *testing.F) {
	corpus := []string{
		"chrome",
		"chrome.exe",
		"explorer",
		"",
		"chrome & notepad",
		"chrome | rm",
		"../dangerous",
		"--help",
		"chrome\x00.exe",
	}
	for _, s := range corpus {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, execName string) {
		result := NormalizeKillExec(execName)

		if result == "" {
			return // rejected — always safe
		}

		// If not rejected, must be a safe, simple ASCII string
		if strings.ContainsAny(result, "&|;`$(){}[]'\"\\\x00") {
			t.Errorf("result contains shell metacharacters: %q from input %q", result, execName)
		}
		if strings.HasPrefix(result, "-") {
			t.Errorf("result starts with flag prefix: %q from input %q", result, execName)
		}
		if strings.Contains(result, "..") {
			t.Errorf("result contains path traversal: %q from input %q", result, execName)
		}
		for _, r := range result {
			if r > 0x7F {
				t.Errorf("result contains non-ASCII rune %q from input %q", r, execName)
			}
		}
	})
}

// FuzzIsAlwaysAllowed tests that localhost bypass logic never has false negatives.
func FuzzIsAlwaysAllowed(f *testing.F) {
	corpus := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"example.com",
		"evil.com",
		"",
		"wails",
		"wails.localhost:34115",
	}
	for _, s := range corpus {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, host string) {
		allowed := isAlwaysAllowed(host)

		if host == "" {
			if !allowed {
				t.Error("empty host should always be allowed")
			}
			return
		}

		lower := strings.ToLower(host)
		// Must never block known-safe hosts
		if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" || lower == "0.0.0.0" {
			if !allowed {
				t.Errorf("loopback host %q should always be allowed", host)
			}
		}
		if strings.HasPrefix(lower, "wails") && !allowed {
			t.Errorf("wails host %q should always be allowed", host)
		}
	})
}

// FuzzIsAllowed tests the URL blocker's allow list logic for edge cases.
func FuzzIsAllowed(f *testing.F) {
	corpus := []string{
		"example.com",
		"sub.example.com",
		"evil.com",
		"localhost",
		"",
	}
	for _, s := range corpus {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, host string) {
		ub := NewURLBlocker()
		ub.allowedURLs = []string{"example.com"}

		allowed := ub.isAllowed(host)

		// localhost must never be blocked regardless of allowed list
		lower := strings.ToLower(strings.TrimSpace(host))
		if lower == "localhost" || lower == "127.0.0.1" {
			if !allowed {
				t.Errorf("localhost %q must never be blocked", host)
			}
		}
	})
}
