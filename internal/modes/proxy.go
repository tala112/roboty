package modes

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// proxyLoopTimeout is the maximum lifetime of a CONNECT tunnel transfer.
const proxyLoopTimeout = 60 * time.Second

const DefaultProxyPort = 62828

type loggingListener struct {
	net.Listener
	addr string
}

type debugConn struct {
	net.Conn
	once sync.Once
}

func (dc *debugConn) Read(b []byte) (int, error) {
	n, err := dc.Conn.Read(b)
	dc.once.Do(func() {
		if n > 0 {
			dump := make([]byte, n)
			copy(dump, b[:n])
			log.Printf("[proxy] 📦 RAW (%d bytes): %x", n, dump)
			log.Printf("[proxy] 📝 TEXT: %q", string(dump))
		}
	})
	return n, err
}

func (ll *loggingListener) Accept() (net.Conn, error) {
	conn, err := ll.Listener.Accept()
	if err == nil {
		log.Printf("[proxy] 🔌 TCP CONNECTION from %s", conn.RemoteAddr())
		conn = &debugConn{Conn: conn}
	}
	return conn, err
}

type URLBlocker struct {
	port        int
	allowedURLs []string
	passThrough bool
	listener    net.Listener
	server      *http.Server
	running     bool
	ready       chan struct{}
	mu          sync.Mutex
	proxyMgr    SystemProxyManager
	stateFile   StateFileManager
	epoch       int64
}

func NewURLBlocker() *URLBlocker {
	return &URLBlocker{
		port:      DefaultProxyPort,
		proxyMgr:  NewRealProxyManager(),
		stateFile: NewFileStateManager(),
	}
}

func NewURLBlockerWithDI(proxyMgr SystemProxyManager, stateFile StateFileManager) *URLBlocker {
	return &URLBlocker{
		port:      DefaultProxyPort,
		proxyMgr:  proxyMgr,
		stateFile: stateFile,
	}
}

func (ub *URLBlocker) Start(port int, allowedURLs []string) error {
	normalized := normalizeURLs(allowedURLs)
	log.Printf("[proxy] Normalized allow-list: %v (from raw: %v)", normalized, allowedURLs)

	ub.mu.Lock()

	// Already running in normal mode — just refresh URL config
	if ub.running && !ub.passThrough {
		ub.allowedURLs = normalized
		ub.mu.Unlock()
		return nil
	}

	// Running in pass-through mode — update config and retry system proxy
	if ub.running && ub.passThrough {
		ub.passThrough = false
		ub.allowedURLs = normalized
		ub.mu.Unlock()
		if err := ub.enableSystemProxy(); err != nil {
			log.Printf("[proxy] ⚠️ pass-through: still cannot set system proxy (%v) — staying in pass-through", err)
			ub.mu.Lock()
			ub.passThrough = true
			ub.allowedURLs = nil
			ub.mu.Unlock()
		} else {
			log.Printf("[proxy] 🔄 Recovered from pass-through — system proxy re-enabled")
		}
		return nil
	}

	// First-time start
	ub.allowedURLs = normalized
	ub.passThrough = false

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("[proxy] ⚠️ Port %d unavailable (%v) — trying random port", port, err)
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			ub.mu.Unlock()
			return fmt.Errorf("proxy listen: %w", err)
		}
		actualPort := listener.Addr().(*net.TCPAddr).Port
		ub.port = actualPort
		addr = fmt.Sprintf("127.0.0.1:%d", actualPort)
		log.Printf("[proxy] 🔄 Bound to random port %d instead", actualPort)
	} else {
		ub.port = port
	}
	ub.listener = listener

	// Wrap in logging listener to trace incoming TCP connections
	logLst := &loggingListener{Listener: listener, addr: addr}

	errLog := log.New(log.Writer(), "[proxy:http] ", log.LstdFlags)
	ub.server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[proxy] 🎯 DIRECT HANDLER: %s %s (Host=%s)", r.Method, r.RequestURI, r.Host)
			ub.handleHTTP(w, r)
		}),
		ErrorLog: errLog,
	}
	ub.ready = make(chan struct{})
	ub.running = true
	startEpoch := atomic.AddInt64(&ub.epoch, 1)
	ub.mu.Unlock()

	go func(epoch int64, srv *http.Server, lst net.Listener) {
		log.Printf("[proxy] 🟢 PROXY STARTED on %s", addr)
		close(ub.ready)

		if err := srv.Serve(lst); err != nil && err != http.ErrServerClosed {
			log.Printf("[proxy] serve error: %v", err)
		}

		ub.mu.Lock()
		if atomic.LoadInt64(&ub.epoch) == epoch {
			ub.running = false
		}
		ub.mu.Unlock()
	}(startEpoch, ub.server, logLst)

	<-ub.ready

	if len(normalized) > 0 {
		log.Printf("[proxy] 🔒 Allow-list: %d URLs — all others will be blocked", len(normalized))
	} else {
		log.Printf("[proxy] ⚠️ No allowed URLs configured — ALL URLs will be blocked")
	}

	if err := ub.enableSystemProxy(); err != nil {
		log.Printf("[proxy] ⚠️ URL blocking disabled: cannot set system proxy (%v). App blocking still active.", err)
		return nil
	}
	log.Printf("[proxy] 🌐 System proxy set to 127.0.0.1:%d", ub.port)

	return nil
}

func (ub *URLBlocker) Stop() error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	if !ub.running {
		return nil
	}

	var returnErr error

	if err := ub.disableSystemProxy(); err != nil {
		log.Printf("[proxy] ⚠️ proxy disable failed: %v — continuing shutdown", err)
		returnErr = fmt.Errorf("disable system proxy: %w", err)
	}

	ub.running = false
	ub.passThrough = false
	ub.allowedURLs = nil

	if ub.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ub.server.Shutdown(ctx); err != nil {
			log.Printf("[proxy] shutdown error: %v", err)
			if returnErr == nil {
				returnErr = fmt.Errorf("proxy shutdown: %w", err)
			}
		}
		ub.server = nil
	}

	log.Println("[proxy] 🛑 PROXY STOPPED")
	return returnErr
}

func (ub *URLBlocker) SetAllowedURLs(urls []string) {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	ub.allowedURLs = normalizeURLs(urls)
}

func (ub *URLBlocker) Port() int {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	return ub.port
}

func (ub *URLBlocker) IsRunning() bool {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	return ub.running
}

func (ub *URLBlocker) IsAllowed(host string) bool {
	return ub.isAllowed(host)
}

func (ub *URLBlocker) isAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))

	if strings.Count(host, ":") == 1 {
		if idx := strings.LastIndex(host, ":"); idx > 0 {
			host = host[:idx]
		}
	}
	host = strings.TrimSuffix(host, ".")

	if isAlwaysAllowed(host) {
		return true
	}

	ub.mu.Lock()
	passThrough := ub.passThrough
	allowed := ub.allowedURLs
	ub.mu.Unlock()

	if passThrough {
		return true
	}

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
	log.Printf("[proxy] 📨 REQUEST: %s %s (Host=%s, URL=%+v)", r.Method, r.RequestURI, r.Host, r.URL)
	if r.Method == http.MethodConnect {
		ub.handleHTTPS(w, r)
		return
	}
	ub.handleHTTPPlain(w, r)
}

func (ub *URLBlocker) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if host == "" {
		host = r.URL.Hostname()
	}
	if !ub.isAllowed(host) {
		log.Printf("[proxy] 🚫 BLOCKED HTTPS: %s", host)
		ub.writeBlockPage(w)
		return
	}

	log.Printf("[proxy] ✅ ALLOWED HTTPS: %s", host)

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

	go transferTimed(dest, clientConn, proxyLoopTimeout)
	go transferTimed(clientConn, dest, proxyLoopTimeout)
}

func (ub *URLBlocker) handleHTTPPlain(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Hostname()
	if host == "" {
		host = r.Host
	}
	if !ub.isAllowed(host) {
		log.Printf("[proxy] 🚫 BLOCKED HTTP: %s", host)
		ub.writeBlockPage(w)
		return
	}

	log.Printf("[proxy] ✅ ALLOWED HTTP: %s", host)

	transport := &http.Transport{
		Proxy: nil, // Never use environment proxy — prevents proxy loops
	}
	// Ensure Host header is set correctly on the outgoing request.
	// Go's http.Server sets r.Host from the incoming Host header, but
	// the transport needs r.Host to match r.URL.Host for proper forwarding.
	if r.Host == "" {
		r.Host = r.URL.Host
	}
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, blockPageHTML)
}

const proxyStateName = "roboty_proxy_state"

func (ub *URLBlocker) saveProxyState() error {
	if ub.stateFile == nil {
		return nil
	}
	return ub.stateFile.SaveState(proxyStateName)
}

func (ub *URLBlocker) clearProxyState() {
	if ub.stateFile == nil {
		return
	}
	ub.stateFile.ClearState(proxyStateName)
}

func (ub *URLBlocker) proxyStateExists() bool {
	if ub.stateFile == nil {
		return false
	}
	return ub.stateFile.StateExists(proxyStateName)
}

func CleanupOrphanedProxy() {
	stateFile := NewFileStateManager()
	if !stateFile.StateExists(proxyStateName) {
		return
	}
	log.Println("[proxy] ⚠️ Detected orphaned proxy state from previous crash — cleaning up")
	stateFile.ClearState(proxyStateName)

	proxyMgr := NewRealProxyManager()
	if err := proxyMgr.Disable(); err != nil {
		log.Printf("[proxy] ⚠️ Orphaned proxy cleanup failed: %v", err)
	} else {
		log.Println("[proxy] ✅ Orphaned proxy cleaned up — system proxy disabled")
	}
}

func (ub *URLBlocker) getProxyManager() SystemProxyManager {
	if ub.proxyMgr != nil {
		return ub.proxyMgr
	}
	return NewRealProxyManager()
}

func (ub *URLBlocker) enableSystemProxy() error {
	mgr := ub.getProxyManager()
	err := mgr.Enable("127.0.0.1", ub.port)
	if err == nil {
		if saveErr := ub.saveProxyState(); saveErr != nil {
			log.Printf("[proxy] proxy enabled but state file not written: %v", saveErr)
		}
		log.Printf("[proxy] 🌐 System proxy enabled -> 127.0.0.1:%d", ub.port)
	} else {
		log.Printf("[proxy] ⚠️ Failed to set system proxy: %v", err)

		if runtime.GOOS == "linux" && strings.Contains(err.Error(), "gsettings") {
			log.Println("[proxy] gsettings not available — proxy setting skipped (headless/CI)")
		}
	}
	return err
}

func (ub *URLBlocker) disableSystemProxy() error {
	mgr := ub.getProxyManager()
	err := mgr.Disable()
	ub.clearProxyState()
	if err == nil {
		log.Println("[proxy] 🌐 System proxy disabled")
	}
	return err
}

func normalizeURLs(urls []string) []string {
	normalized := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		u = strings.ToLower(u)
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		u = strings.TrimSuffix(u, "/")
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

func transferTimed(dest io.WriteCloser, src io.ReadCloser, timeout time.Duration) {
	defer dest.Close()
	defer src.Close()
	if timeout <= 0 {
		io.Copy(dest, src)
		return
	}
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(dest, src)
		done <- err
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		log.Printf("[proxy] ⚠️ CONNECT tunnel timed out after %v", timeout)
	}
}

const blockPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Blocked</title>
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
  <div class="icon">X</div>
  <h1>Site Blocked</h1>
  <p>This site is blocked by Roboty Focus Mode.</p>
  <span class="badge">Focus Mode Active</span>
</div>
</body>
</html>`
