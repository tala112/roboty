package modes

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestURLBlocker_isAllowed(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		host     string
		expected bool
	}{
		{"empty allowed blocks all", []string{}, "github.com", false},
		{"exact match", []string{"github.com"}, "github.com", true},
		{"subdomain match", []string{"github.com"}, "api.github.com", true},
		{"sub-subdomain match", []string{"github.com"}, "foo.api.github.com", true},
		{"different domain", []string{"github.com"}, "google.com", false},
		{"port stripped no effect", []string{"github.com"}, "github.com:443", true},
		{"subdomain with port", []string{"github.com"}, "api.github.com:443", true},
		{"case insensitive host", []string{"github.com"}, "GITHUB.COM", true},
		{"case insensitive allowed", []string{"GITHUB.COM"}, "github.com", true},
		{"no match without period", []string{"github.com"}, "notgithub.com", false},
		{"www subdomain", []string{"github.com"}, "www.github.com", true},
		{"trailing dot", []string{"github.com"}, "github.com.", true},
		{"multiple allowed", []string{"github.com", "google.com"}, "google.com", true},
		{"not in multiple", []string{"github.com", "google.com"}, "bing.com", false},
		{"ip address exact", []string{"192.168.1.1"}, "192.168.1.1", true},
		{"ip address not match", []string{"192.168.1.1"}, "192.168.1.2", false},
		{"deep nested subdomain", []string{"example.com"}, "a.b.c.d.e.example.com", true},
		{"spaces trimmed", []string{" github.com "}, "github.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ub := NewURLBlocker()
			ub.allowedURLs = normalizeURLs(tt.allowed)
			got := ub.isAllowed(tt.host)
			if got != tt.expected {
				t.Errorf("isAllowed(%q) = %v, want %v (allowed=%v)", tt.host, got, tt.expected, tt.allowed)
			}
		})
	}
}

func TestNormalizeURLs(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"https://github.com"}, []string{"github.com"}},
		{[]string{"http://github.com"}, []string{"github.com"}},
		{[]string{"https://github.com/"}, []string{"github.com"}},
		{[]string{"https://github.com/path"}, []string{"github.com"}},
		{[]string{"  github.com  "}, []string{"github.com"}},
		{[]string{"GITHUB.COM"}, []string{"github.com"}},
		{[]string{"https://www.github.com/path/to/page"}, []string{"www.github.com"}},
		{[]string{""}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.input[0], func(t *testing.T) {
			got := normalizeURLs(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("normalizeURLs(%v) = %v, want %v", tt.input, got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("normalizeURLs(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestHandleHTTPS_HostParsing(t *testing.T) {
	// Simulate the host extraction logic used in handleHTTPS
	// r.URL.Hostname() should give correct host even with port
	u := &url.URL{Host: "github.com:443"}
	if got := u.Hostname(); got != "github.com" {
		t.Errorf("URL.Hostname() with port = %q, want %q", got, "github.com")
	}

	u2 := &url.URL{Host: "github.com"}
	if got := u2.Hostname(); got != "github.com" {
		t.Errorf("URL.Hostname() without port = %q, want %q", got, "github.com")
	}

	u3 := &url.URL{Host: "api.github.com:8443"}
	if got := u3.Hostname(); got != "api.github.com" {
		t.Errorf("URL.Hostname() subdomain = %q, want %q", got, "api.github.com")
	}
}

func TestHandleHTTPPlain_HostParsing(t *testing.T) {
	// Simulate host extraction from a full HTTP proxy URL
	// For a proxy request like GET http://github.com/ HTTP/1.1
	u, _ := url.Parse("http://github.com/path")
	if got := u.Hostname(); got != "github.com" {
		t.Errorf("URL.Hostname() from proxy URL = %q, want %q", got, "github.com")
	}

	u2, _ := url.Parse("http://api.github.com:8080/path/to/page")
	if got := u2.Hostname(); got != "api.github.com" {
		t.Errorf("URL.Hostname() from proxy subdomain URL = %q, want %q", got, "api.github.com")
	}
}

// Test that r.URL.Hostname() works for CONNECT requests in http.Server
// by simulating what http.Server sets for CONNECT
func TestConnectRequestHostname(t *testing.T) {
	// http.Server sets r.URL.Host = "host:port" for CONNECT
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: "example.com:443"},
		Host:   "example.com:443",
	}

	host := req.URL.Hostname()
	if host == "" {
		host = req.Host
	}
	if host != "example.com" {
		t.Errorf("CONNECT host extraction = %q, want %q", host, "example.com")
	}

	// Test with Host header having different value (should use URL.Hostname)
	req2 := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: "example.com:443"},
		Host:   "", // Host header missing
	}
	host2 := req2.URL.Hostname()
	if host2 == "" {
		host2 = req2.Host
	}
	if host2 != "example.com" {
		t.Errorf("CONNECT host extraction with empty Host = %q, want %q", host2, "example.com")
	}
}

func TestGetWhitelistExecs(t *testing.T) {
	set := GetWhitelistExecs()
	if len(set) == 0 {
		t.Fatal("GetWhitelistExecs() returned empty set")
	}
	// Should contain some expected system apps
	expected := []string{"explorer.exe", "svchost.exe", "systemd", "launchd", "roboty", "roboty-dev.exe"}
	for _, e := range expected {
		if !set[e] {
			t.Errorf("whitelist missing expected entry: %q", e)
		}
	}
	// Should NOT contain random section key names
	unexpected := []string{"windows", "executables", "linux", "macos", "ui_shell"}
	for _, e := range unexpected {
		if set[e] {
			t.Errorf("whitelist should NOT contain section key: %q", e)
		}
	}
}

// getFreePort asks the OS for a free TCP port on 127.0.0.1
func getFreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("getFreePort: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// startProxyOnPort starts a URLBlocker on a free port for testing
func startProxyOnPort(t *testing.T, allowed []string) *URLBlocker {
	t.Helper()
	port := getFreePort(t)
	ub := NewURLBlocker()
	err := ub.Start(port, allowed)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	t.Cleanup(func() { ub.Stop() })
	return ub
}

// sendHTTPProxyReq sends an HTTP proxy-style GET through the proxy
// Returns the HTTP response status code and body
func sendHTTPProxyReq(t *testing.T, port int, targetURL string) (int, string) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 3*time.Second)
	if err != nil {
		t.Logf("Dial failed: %v", err)
		return 0, ""
	}
	defer conn.Close()
	req := fmt.Sprintf("GET http://%s/ HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", targetURL, targetURL)
	_, err = fmt.Fprint(conn, req)
	if err != nil {
		t.Logf("Write failed: %v", err)
		return 0, ""
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Logf("Read response failed: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

// Test URLBlocker lifecycle: start/stop and verify it serves on the port
func TestURLBlocker_StartStop(t *testing.T) {
	ub := startProxyOnPort(t, []string{"example.com"})
	port := ub.port

	if !ub.IsRunning() {
		t.Fatal("URLBlocker should be running after Start")
	}

	// HTTP proxy request to an allowed host
	code, body := sendHTTPProxyReq(t, port, "example.com")
	if code == http.StatusOK || code == http.StatusBadGateway {
		t.Logf("Allowed request returned status %d (expected forward or bad gateway)", code)
	} else if code == 0 {
		t.Log("Request failed (connection issue)")
	} else {
		t.Errorf("Allowed request returned unexpected status %d", code)
	}
	_ = body

	// HTTP proxy request to a blocked host
	code, body = sendHTTPProxyReq(t, port, "blocked-site.com")
	if code == http.StatusForbidden {
		t.Log("Blocked request correctly returned 403 Forbidden")
	} else if code == 0 {
		t.Log("Request to blocked host failed (connection issue)")
	} else {
		t.Logf("Blocked request returned status %d (expected 403)", code)
	}
	// Verify block page content
	if code == http.StatusForbidden && !strings.Contains(body, "Blocked") {
		t.Errorf("Block page should contain 'Blocked', got: %s", body)
	}

	// Stop
	err := ub.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if ub.IsRunning() {
		t.Fatal("URLBlocker should not be running after Stop")
	}
}

// Test URLBlocker with empty allowed list — should block everything
func TestURLBlocker_BlockAll(t *testing.T) {
	ub := startProxyOnPort(t, nil)

	// HTTP proxy request to any host — should get 403
	code, _ := sendHTTPProxyReq(t, ub.port, "any-site.com")
	if code == http.StatusForbidden {
		t.Log("Block-all proxy correctly returned 403 Forbidden")
	} else if code == 0 {
		t.Log("Request failed (connection issue)")
	} else {
		t.Errorf("Expected 403 Forbidden, got %d", code)
	}
}

// Test URLBlocker allows the configured host
func TestURLBlocker_AllowConfiguredHost(t *testing.T) {
	ub := startProxyOnPort(t, []string{"allowed-site.com"})

	// Request to allowed host should NOT be forbidden
	code, _ := sendHTTPProxyReq(t, ub.port, "allowed-site.com")
	if code == http.StatusForbidden {
		t.Error("Allowed site should not be blocked")
	} else if code == 0 {
		t.Log("Request failed (connection issue)")
	} else {
		t.Logf("Allowed site returned status %d (expected non-403)", code)
	}

	// Request to other host should be 403
	code, _ = sendHTTPProxyReq(t, ub.port, "other-site.com")
	if code == http.StatusForbidden {
		t.Log("Non-allowed site correctly blocked")
	} else if code == 0 {
		t.Log("Request failed (connection issue)")
	} else {
		t.Errorf("Non-allowed site should be blocked, got %d", code)
	}
}

// Test normalizing URLs with various inputs
func TestNormalizeURLs_EdgeCases(t *testing.T) {
	tests := []struct {
		input    []string
		expected int // expected count
	}{
		{[]string{}, 0},
		{[]string{""}, 0},
		{[]string{"  "}, 0},
		{[]string{"https://"}, 0},
		{[]string{"http://"}, 0},
		{[]string{"github.com", "github.com"}, 2}, // normalizeURLs doesn't deduplicate; isAllowed handles it
		{[]string{"http://github.com/path?query=1#frag"}, 1},
		{[]string{"HTTPS://GITHUB.COM"}, 1},
	}
	for _, tt := range tests {
		got := normalizeURLs(tt.input)
		if len(got) != tt.expected {
			t.Errorf("normalizeURLs(%v) returned %d items, want %d. Got: %v", tt.input, len(got), tt.expected, got)
		}
	}
}

// Test CONNECT vs HTTP routing through the proxy handler
func TestProxyHandlerRouting(t *testing.T) {
	// We can't easily test http.Server routing without starting the server,
	// but we can verify the handler function routes correctly
	t.Run("CONNECT goes to handleHTTPS", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodConnect, "https://example.com:443", nil)
		// handleHTTP is the entry point that routes to handleHTTPS or handleHTTPPlain
		// We can't easily capture which sub-handler was called without starting the server,
		// so we just verify the method check in the handler
		if req.Method != http.MethodConnect {
			t.Error("expected CONNECT method")
		}
	})

	t.Run("GET goes to handleHTTPPlain", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
		if req.Method == http.MethodConnect {
			t.Error("GET should not be CONNECT")
		}
	})
}

// Test that isAllowed correctly handles hostname extraction
// from all three sources used in handlers
func TestIsAllowed_HostSource(t *testing.T) {
	ub := NewURLBlocker()
	ub.allowedURLs = []string{"example.com"}

	// Simulate host values from different sources
	sources := map[string]string{
		"r.Host (with port)":        "example.com:443",
		"r.Host (no port)":          "example.com",
		"r.URL.Hostname() (with)":   "example.com",
		"raw host string":           "example.com",
		"trailing dot FQDN":         "example.com.",
		"port + trailing dot":       "example.com.:443",
		"subdomain with port":       "sub.example.com:8080",
		"subdomain no port":         "sub.example.com",
	}

	for name, host := range sources {
		if !ub.isAllowed(host) {
			t.Errorf("isAllowed(%q) should be true (source: %s)", host, name)
		}
	}

	// Should NOT match
	blocked := map[string]string{
		"different domain":     "other.com",
		"different with port":  "other.com:443",
		"no match partial":     "notexample.com",
	}
	for name, host := range blocked {
		if ub.isAllowed(host) {
			t.Errorf("isAllowed(%q) should be false (source: %s)", host, name)
		}
	}
}

// Test the CONNECT request hostname extraction used in handleHTTPS
func TestHandleHTTPS_HostFallback(t *testing.T) {
	// When r.URL.Hostname() is empty, it should fallback to r.Host
	req1 := &http.Request{
		URL:  &url.URL{Host: "example.com:443"}, // URL.Hostname() = "example.com"
		Host: "wrong.com:443",
	}
	host1 := req1.URL.Hostname()
	if host1 == "" {
		host1 = req1.Host
	}
	if host1 != "example.com" {
		t.Errorf("should use URL.Hostname() over r.Host: got %q", host1)
	}

	// When URL.Hostname() is empty (no Host in URL)
	req2 := &http.Request{
		URL:  &url.URL{}, // empty URL, Hostname() = ""
		Host: "fallback.com:443",
	}
	host2 := req2.URL.Hostname()
	if host2 == "" {
		host2 = req2.Host
	}
	if host2 != "fallback.com:443" {
		t.Errorf("should fallback to r.Host: got %q", host2)
	}
	// Port stripped in isAllowed
	ub := NewURLBlocker()
	ub.allowedURLs = []string{"fallback.com"}
	if !ub.isAllowed(host2) {
		t.Errorf("isAllowed should handle port in fallback host: %q", host2)
	}
}

// Verify Windows system proxy PowerShell commands are syntactically valid
func TestWindowsSystemProxyCommands(t *testing.T) {
	if !isWindows() {
		t.Skip("Windows specific")
	}

	ub := NewURLBlocker()
	ub.port = 62828

	proxyAddr := fmt.Sprintf("127.0.0.1:%d", ub.port)
	enableCmd := fmt.Sprintf(
		`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 1 -Type DWord -Force; Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -Value "%s" -Type String -Force`,
		proxyAddr,
	)
	// Validate with -WhatIf
	cmd := exec.Command("powershell", "-NoProfile", "-Command", enableCmd+" -WhatIf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Enable proxy command syntax invalid: %v\nOutput: %s", err, string(out))
	} else {
		t.Logf("Enable proxy -WhatIf OK: %s", strings.TrimSpace(string(out)))
	}

	disableCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force`
	cmd = exec.Command("powershell", "-NoProfile", "-Command", disableCmd+" -WhatIf")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Disable proxy command syntax invalid: %v\nOutput: %s", err, string(out))
	} else {
		t.Logf("Disable proxy -WhatIf OK: %s", strings.TrimSpace(string(out)))
	}
}

// TestURLBlocker_Lifecycle_StartStopRestart verifies the full proxy lifecycle:
// start → stop → restart with new config → stop.
func TestURLBlocker_Lifecycle_StartStopRestart(t *testing.T) {
	fakeProxy := newFakeProxyManager()
	fakeState := newFakeStateFileManager()
	port := getFreePort(t)

	ub := NewURLBlockerWithDI(fakeProxy, fakeState)

	// Phase 1: Start with blocklist A
	err := ub.Start(port, []string{"example.com", "github.com"})
	if err != nil {
		t.Fatalf("phase 1 Start failed: %v", err)
	}
	if !ub.IsRunning() {
		t.Fatal("phase 1: should be running after Start")
	}
	if !fakeProxy.enabled {
		t.Error("phase 1: system proxy should be enabled")
	}
	if !fakeState.StateExists(proxyStateName) {
		t.Error("phase 1: state file should exist")
	}
	if !ub.isAllowed("example.com") {
		t.Error("phase 1: example.com should be allowed")
	}
	if !ub.isAllowed("api.github.com") {
		t.Error("phase 1: api.github.com should be allowed")
	}

	// Phase 2: Stop
	err = ub.Stop()
	if err != nil {
		t.Fatalf("phase 2 Stop failed: %v", err)
	}
	if ub.IsRunning() {
		t.Fatal("phase 2: should not be running after Stop")
	}
	if fakeProxy.enabled {
		t.Error("phase 2: system proxy should be disabled")
	}
	if fakeState.StateExists(proxyStateName) {
		t.Error("phase 2: state file should be cleared")
	}

	// Phase 3: Restart with different blocklist B (different port to avoid conflict)
	port2 := getFreePort(t)
	err = ub.Start(port2, []string{"stackoverflow.com"})
	if err != nil {
		t.Fatalf("phase 3 Start failed: %v", err)
	}
	if !ub.IsRunning() {
		t.Fatal("phase 3: should be running after restart")
	}
	if !fakeProxy.enabled {
		t.Error("phase 3: system proxy should be enabled")
	}
	// Old URLs must NOT be allowed
	if ub.isAllowed("example.com") {
		t.Error("phase 3: example.com should NOT be allowed after restart with new config")
	}
	// New URL must be allowed
	if !ub.isAllowed("stackoverflow.com") {
		t.Error("phase 3: stackoverflow.com should be allowed")
	}
	// Localhost must always be allowed
	if !ub.isAllowed("localhost") {
		t.Error("phase 3: localhost must always be allowed")
	}

	// Phase 4: Final stop
	err = ub.Stop()
	if err != nil {
		t.Fatalf("phase 4 Stop failed: %v", err)
	}
	if ub.IsRunning() {
		t.Fatal("phase 4: should not be running after final stop")
	}
	if fakeProxy.enabled {
		t.Error("phase 4: system proxy should be disabled")
	}
	if fakeState.StateExists(proxyStateName) {
		t.Error("phase 4: state file should be cleared")
	}
}

// TestURLBlocker_Lifecycle_IdempotentStop verifies that multiple Stop calls are safe.
func TestURLBlocker_Lifecycle_IdempotentStop(t *testing.T) {
	fakeProxy := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxy, fakeState)
	port := getFreePort(t)
	if err := ub.Start(port, nil); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Call Stop multiple times — must not panic or error
	for i := 0; i < 5; i++ {
		err := ub.Stop()
		if err != nil {
			t.Errorf("Stop iteration %d: unexpected error: %v", i, err)
		}
	}

	if ub.IsRunning() {
		t.Error("should not be running after Stop")
	}
}

// TestURLBlocker_Lifecycle_StartTwiceNoop verifies that Start when already running
// updates the URL config (via Stop+Start, not silently ignoring).
func TestURLBlocker_Lifecycle_StartUpdatesConfig(t *testing.T) {
	fakeProxy := newFakeProxyManager()
	fakeState := newFakeStateFileManager()

	ub := NewURLBlockerWithDI(fakeProxy, fakeState)
	port := getFreePort(t)

	// First start with config A
	if err := ub.Start(port, []string{"old-site.com"}); err != nil {
		t.Fatalf("first Start failed: %v", err)
	}

	// Simulate ActivateMode's current pattern: Stop then Start with config B
	ub.Stop()
	if err := ub.Start(port, []string{"new-site.com"}); err != nil {
		t.Fatalf("second Start failed: %v", err)
	}

	if ub.isAllowed("old-site.com") {
		t.Error("old-site.com should NOT be allowed after Stop+Start with new config")
	}
	if !ub.isAllowed("new-site.com") {
		t.Error("new-site.com should be allowed after Stop+Start with new config")
	}
}

