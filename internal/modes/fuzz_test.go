package modes

import (
	"strings"
	"testing"
)

// TestSeed_NormalizeKillExec verifies the fuzz seed corpus for NormalizeKillExec.
// Kept as a regular test for reliable CI runs (fuzz tests use Go 1.26's
// context-based timer which can report "context deadline exceeded" as FAIL).
func TestSeed_NormalizeKillExec(t *testing.T) {
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
	for _, execName := range corpus {
		result := NormalizeKillExec(execName)

		if result == "" {
			continue // rejected — always safe
		}

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
	}
}

// TestSeed_IsAlwaysAllowed verifies the fuzz seed corpus for isAlwaysAllowed.
func TestSeed_IsAlwaysAllowed(t *testing.T) {
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
	for _, host := range corpus {
		allowed := isAlwaysAllowed(host)

		if host == "" {
			if !allowed {
				t.Error("empty host should always be allowed")
			}
			continue
		}

		lower := strings.ToLower(host)
		if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" || lower == "0.0.0.0" {
			if !allowed {
				t.Errorf("loopback host %q should always be allowed", host)
			}
		}
		if strings.HasPrefix(lower, "wails") && !allowed {
			t.Errorf("wails host %q should always be allowed", host)
		}
	}
}

// TestSeed_IsAllowed verifies the fuzz seed corpus for isAllowed.
func TestSeed_IsAllowed(t *testing.T) {
	corpus := []string{
		"example.com",
		"sub.example.com",
		"evil.com",
		"localhost",
		"",
	}
	for _, host := range corpus {
		ub := NewURLBlocker()
		ub.allowedURLs = []string{"example.com"}

		allowed := ub.isAllowed(host)

		lower := strings.ToLower(strings.TrimSpace(host))
		if lower == "localhost" || lower == "127.0.0.1" {
			if !allowed {
				t.Errorf("localhost %q must never be blocked", host)
			}
		}
	}
}

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
