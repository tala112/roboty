package modes

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const DefaultProxyPort = 62828

type URLBlocker struct {
	port        int
	allowedURLs []string
	listener    net.Listener
	server      *http.Server
	running     bool
	mu          sync.Mutex
}

func NewURLBlocker() *URLBlocker {
	return &URLBlocker{
		port: DefaultProxyPort,
	}
}

func (ub *URLBlocker) Start(port int, allowedURLs []string) error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	if ub.running {
		return nil
	}

	ub.port = port
	ub.allowedURLs = normalizeURLs(allowedURLs)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("proxy listen on %s: %w", addr, err)
	}
	ub.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/", ub.handleHTTP)

	ub.server = &http.Server{
		Handler: mux,
	}
	ub.running = true

	go func() {
		log.Printf("[proxy] starting on %s", addr)
		if err := ub.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("[proxy] serve error: %v", err)
		}
	}()

	// Enable system proxy settings
	if err := ub.enableSystemProxy(); err != nil {
		log.Printf("[proxy] failed to set system proxy: %v", err)
	}

	return nil
}

func (ub *URLBlocker) Stop() error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	if !ub.running {
		return nil
	}

	ub.running = false

	// Disable system proxy
	if err := ub.disableSystemProxy(); err != nil {
		log.Printf("[proxy] failed to disable system proxy: %v", err)
	}

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ub.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("proxy shutdown: %w", err)
	}

	log.Println("[proxy] stopped")
	return nil
}

func (ub *URLBlocker) SetAllowedURLs(urls []string) {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	ub.allowedURLs = normalizeURLs(urls)
}

func (ub *URLBlocker) IsRunning() bool {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	return ub.running
}

// =============================================================================
// Proxy Logic — block ALL sites except those in the allowed list
// =============================================================================

func (ub *URLBlocker) isAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	// Strip port if present
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		host = host[:idx]
	}

	ub.mu.Lock()
	allowed := ub.allowedURLs
	ub.mu.Unlock()

	// If no allowed URLs configured, block everything
	if len(allowed) == 0 {
		return false
	}

	for _, allowedURL := range allowed {
		if host == allowedURL || strings.HasSuffix(host, "."+allowedURL) {
			return true
		}
	}
	return false
}

func (ub *URLBlocker) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		ub.handleHTTPS(w, r)
		return
	}
	ub.handleHTTPPlain(w, r)
}

func (ub *URLBlocker) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if !ub.isAllowed(host) {
		log.Printf("[proxy] blocked HTTPS: %s", host)
		ub.writeBlockPage(w)
		return
	}

	// Tunnel the connection
	dest, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(dest, clientConn)
	go transfer(clientConn, dest)
}

func (ub *URLBlocker) handleHTTPPlain(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if !ub.isAllowed(host) {
		log.Printf("[proxy] blocked HTTP: %s", host)
		ub.writeBlockPage(w)
		return
	}

	// Forward the request
	transport := &http.Transport{}
	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (ub *URLBlocker) writeBlockPage(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, blockPageHTML)
}

// =============================================================================
// System Proxy Management
// =============================================================================

func (ub *URLBlocker) enableSystemProxy() error {
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", ub.port)

	switch runtime.GOOS {
	case "linux":
		return ub.enableProxyLinux(proxyAddr)
	case "darwin":
		return ub.enableProxyMacOS(proxyAddr)
	case "windows":
		return ub.enableProxyWindows(proxyAddr)
	}
	return nil
}

func (ub *URLBlocker) disableSystemProxy() error {
	switch runtime.GOOS {
	case "linux":
		return ub.disableProxyLinux()
	case "darwin":
		return ub.disableProxyMacOS()
	case "windows":
		return ub.disableProxyWindows()
	}
	return nil
}

func (ub *URLBlocker) enableProxyLinux(proxyAddr string) error {
	cmds := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "manual"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", "127.0.0.1"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", fmt.Sprintf("%d", ub.port)},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", "127.0.0.1"},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", fmt.Sprintf("%d", ub.port)},
	}
	for _, args := range cmds {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			log.Printf("[proxy] linux cmd failed: %v", err)
		}
	}
	log.Println("[proxy] Linux system proxy enabled")
	return nil
}

func (ub *URLBlocker) disableProxyLinux() error {
	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	return cmd.Run()
}

func (ub *URLBlocker) enableProxyMacOS(proxyAddr string) error {
	// Get active network service
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return fmt.Errorf("list network services: %w", err)
	}
	services := strings.Split(string(netService), "\n")
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" || strings.Contains(s, "An asterisk") {
			continue
		}
		exec.Command("networksetup", "-setwebproxy", s, "127.0.0.1", fmt.Sprintf("%d", ub.port)).Run()
		exec.Command("networksetup", "-setsecurewebproxy", s, "127.0.0.1", fmt.Sprintf("%d", ub.port)).Run()
		exec.Command("networksetup", "-setwebproxystate", s, "on").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", s, "on").Run()
	}
	log.Println("[proxy] macOS system proxy enabled")
	return nil
}

func (ub *URLBlocker) disableProxyMacOS() error {
	netService, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return err
	}
	services := strings.Split(string(netService), "\n")
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" || strings.Contains(s, "An asterisk") {
			continue
		}
		exec.Command("networksetup", "-setwebproxystate", s, "off").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", s, "off").Run()
	}
	log.Println("[proxy] macOS system proxy disabled")
	return nil
}

func (ub *URLBlocker) enableProxyWindows(proxyAddr string) error {
	psCmd := fmt.Sprintf(`Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 1 -Type DWord -Force; Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -Value "%s" -Type String -Force`, proxyAddr)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy enable: %w - %s", err, string(out))
	}
	log.Println("[proxy] Windows system proxy enabled")
	return nil
}

func (ub *URLBlocker) disableProxyWindows() error {
	psCmd := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0 -Type DWord -Force`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows proxy disable: %w - %s", err, string(out))
	}
	log.Println("[proxy] Windows system proxy disabled")
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

func normalizeURLs(urls []string) []string {
	normalized := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		u = strings.ToLower(u)
		// Strip protocol
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		// Strip trailing slash
		u = strings.TrimSuffix(u, "/")
		// Strip path
		if idx := strings.Index(u, "/"); idx > 0 {
			u = u[:idx]
		}
		if u != "" {
			normalized = append(normalized, u)
		}
	}
	return normalized
}

func transfer(dest io.WriteCloser, src io.ReadCloser) {
	defer dest.Close()
	defer src.Close()
	io.Copy(dest, src)
}

const blockPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Blocked — Roboty Focus Mode</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #0f0f1a; color: #e0e0e0; }
  .card { background: #1a1a2e; border-radius: 16px; padding: 48px; max-width: 480px; text-align: center; border: 1px solid #2a2a4a; }
  .icon { font-size: 64px; margin-bottom: 16px; }
  h1 { font-size: 24px; margin: 0 0 8px; }
  p { color: #888; margin: 0 0 24px; line-height: 1.5; }
  .badge { display: inline-block; background: #e74c3c22; color: #e74c3c; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">🔒</div>
  <h1>Site Blocked</h1>
  <p>This site is blocked by <strong>Roboty Focus Mode</strong>.<br>Only allowed sites can be accessed during focus sessions.</p>
  <span class="badge">Focus Mode Active</span>
</div>
</body>
</html>`
