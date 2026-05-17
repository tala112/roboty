package modes

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// =============================================================================
// Safety Constants
// =============================================================================

var (
	WatchdogInterval     = 5 * time.Second
	ProxyHealthTimeout   = 3 * time.Second
	MaxConsecutiveKills  = 10
	KillLoopWindow       = 30 * time.Second
	EmergencyFailsafeMax = 3
)

// SafetyEventType categorizes safety-related events for logging
type SafetyEventType string

const (
	EventBlockedApp     SafetyEventType = "blocked_app"
	EventBlockedURL     SafetyEventType = "blocked_url"
	EventLocalhostBypass SafetyEventType = "localhost_bypass"
	EventWhitelistMatch  SafetyEventType = "whitelist_match"
	EventAncestorMatch   SafetyEventType = "ancestor_match"
	EventProxyDown       SafetyEventType = "proxy_down"
	EventProxyRecovery   SafetyEventType = "proxy_recovery"
	EventKillAttempt     SafetyEventType = "kill_attempt"
	EventKillBlocked     SafetyEventType = "kill_blocked_safety"
	EventEmergencyStop   SafetyEventType = "emergency_stop"
	EventWatchdogAction  SafetyEventType = "watchdog_action"
	EventDevModeOverride SafetyEventType = "dev_mode_override"
)

type SafetyEvent struct {
	Type      SafetyEventType `json:"type"`
	Target    string          `json:"target,omitempty"`
	Source    string          `json:"source,omitempty"`
	Allowed   bool            `json:"allowed,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Message   string          `json:"message"`
}

// =============================================================================
// Panic-Safe Goroutine Launcher
// =============================================================================

var globalEmergencyCallback func(reason string)

// SetGlobalEmergencyCallback sets the callback for panic recovery.
// Called by ModeService.EmergencyStop on unrecoverable errors.
func SetGlobalEmergencyCallback(cb func(reason string)) {
	globalEmergencyCallback = cb
}

// safeGo launches a goroutine with panic recovery. If the goroutine panics,
// it logs the error, calls the emergency callback, and recovers to prevent
// the entire process from crashing.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[safety] 🚨 PANIC RECOVERED: %v", r)
				if globalEmergencyCallback != nil {
					globalEmergencyCallback("goroutine-panic")
				}
			}
		}()
		fn()
	}()
}

// SetupSignalHandler installs handlers for SIGINT and SIGTERM that trigger
// the global emergency callback and then exit cleanly.
func SetupSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	safeGo(func() {
		sig := <-sigCh
		log.Printf("[safety] 📡 Received signal %v — initiating emergency stop", sig)
		if globalEmergencyCallback != nil {
			globalEmergencyCallback("signal-" + sig.String())
		}
		os.Exit(1)
	})
	log.Println("[safety] Signal handler installed for SIGINT/SIGTERM")
}

// =============================================================================
// Notification State Persistence (crash recovery)
// =============================================================================

const notifStateName = "roboty_notif_state"

// getNotifStateFile returns the global state file manager for notifications.
func getNotifStateFile() StateFileManager {
	return NewFileStateManager()
}

func saveNotificationState() {
	sf := getNotifStateFile()
	if err := sf.SaveState(notifStateName); err != nil {
		log.Printf("[safety] failed to save notification state: %v", err)
	}
}

func clearNotificationState() {
	sf := getNotifStateFile()
	sf.ClearState(notifStateName)
}

// CleanupOrphanedNotifications checks if notifications were left muted
// by a previous crash and restores them.
func CleanupOrphanedNotifications() {
	sf := getNotifStateFile()
	if !sf.StateExists(notifStateName) {
		return
	}
	log.Println("[safety] ⚠️ Detected orphaned notification state from previous crash — restoring")
	sf.ClearState(notifStateName)
	RestoreNotifications()
}

// =============================================================================
// Safe Development Mode
// =============================================================================

var isDevMode bool
var devModeOnce sync.Once

func IsDevMode() bool {
	devModeOnce.Do(func() {
		val, _ := strconv.ParseBool(os.Getenv("ROBOTY_SAFE_MODE"))
		isDevMode = val
		if isDevMode {
			log.Println("[safety] ⚠️ SAFE DEVELOPMENT MODE ACTIVE — no OS modifications will be made")
		}
	})
	return isDevMode
}

func ResetDevMode() {
	devModeOnce = sync.Once{}
	isDevMode = false
}

// =============================================================================
// Kill Loop Detector
// =============================================================================

type KillLoopDetector struct {
	mu        sync.Mutex
	killLog   map[string][]time.Time
}

func NewKillLoopDetector() *KillLoopDetector {
	return &KillLoopDetector{
		killLog: make(map[string][]time.Time),
	}
}

func (kld *KillLoopDetector) RecordKill(execName string) bool {
	kld.mu.Lock()
	defer kld.mu.Unlock()

	now := time.Now()
	execName = strings.ToLower(execName)

	entries := kld.killLog[execName]
	recent := make([]time.Time, 0, len(entries)+1)
	for _, t := range entries {
		if now.Sub(t) <= KillLoopWindow {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	kld.killLog[execName] = recent

	return len(recent) >= MaxConsecutiveKills
}

// =============================================================================
// Emergency Failsafe
// =============================================================================

type EmergencyFailsafe struct {
	mu               sync.Mutex
	triggered        bool
	protectedTargets int
	callback         func()
}

func NewEmergencyFailsafe(callback func()) *EmergencyFailsafe {
	return &EmergencyFailsafe{
		callback: callback,
	}
}

func (ef *EmergencyFailsafe) Trigger(reason string, eventLog *[]SafetyEvent) {
	ef.mu.Lock()
	defer ef.mu.Unlock()

	if ef.triggered {
		return
	}
	ef.triggered = true

	log.Printf("[safety] 🚨 EMERGENCY FAILSAFE TRIGGERED: %s", reason)

	*eventLog = append(*eventLog, SafetyEvent{
		Type:      EventEmergencyStop,
		Message:   reason,
		Timestamp: time.Now(),
	})

	if ef.callback != nil {
		ef.callback()
	}
}

func (ef *EmergencyFailsafe) IsTriggered() bool {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	return ef.triggered
}

func (ef *EmergencyFailsafe) Reset() {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.triggered = false
}

// =============================================================================
// Proxy Watchdog — auto-recovery daemon
// =============================================================================

type ProxyWatchdog struct {
	urlBlocker    *URLBlocker
	modeService   *ModeService
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
	mu            sync.Mutex
	eventLog      *[]SafetyEvent
	eventLogMu    *sync.Mutex
	failsafe      *EmergencyFailsafe
}

func NewProxyWatchdog(ub *URLBlocker, ms *ModeService, eventLog *[]SafetyEvent, eventLogMu *sync.Mutex) *ProxyWatchdog {
	return &ProxyWatchdog{
		urlBlocker:  ub,
		modeService: ms,
		eventLog:    eventLog,
		eventLogMu:  eventLogMu,
	}
}

func (pw *ProxyWatchdog) Start() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if pw.running {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	pw.ctx = ctx
	pw.cancel = cancel
	pw.running = true

	pw.failsafe = NewEmergencyFailsafe(func() {
		log.Println("[watchdog] 🚨 EMERGENCY: Calling ModeService emergency stop")
		if pw.modeService != nil {
			pw.modeService.EmergencyStop("watchdog-failsafe")
		}
	})

	go pw.watchdogLoop()
	log.Println("[watchdog] ProxyWatchdog started")
}

func (pw *ProxyWatchdog) Stop() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if !pw.running {
		return
	}

	if pw.cancel != nil {
		pw.cancel()
	}
	pw.running = false
	log.Println("[watchdog] ProxyWatchdog stopped")
}

func (pw *ProxyWatchdog) SetFailsafeCallback(cb func()) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	if pw.failsafe != nil {
		pw.failsafe = NewEmergencyFailsafe(cb)
	}
}

func (pw *ProxyWatchdog) watchdogLoop() {
	ticker := time.NewTicker(WatchdogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pw.ctx.Done():
			return
		case <-ticker.C:
			pw.checkHealth()
		}
	}
}

func (pw *ProxyWatchdog) checkHealth() {
	ub := pw.urlBlocker
	if ub == nil || !ub.IsRunning() {
		return
	}

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(ub.port))

	// Test TCP connectivity first
	conn, err := net.DialTimeout("tcp", addr, ProxyHealthTimeout)
	if err != nil {
		log.Printf("[watchdog] ⚠️ Proxy TCP health check FAILED: %v", err)
		pw.handleProxyDown("tcp-check")
		return
	}
	conn.Close()

	// HTTP health check
	client := &http.Client{Timeout: ProxyHealthTimeout}
	resp, err := client.Get("http://" + addr + "/")
	if err != nil {
		log.Printf("[watchdog] ⚠️ Proxy HTTP health check FAILED: %v", err)
		pw.handleProxyDown("http-check")
		return
	}
	resp.Body.Close()

	// Retry health check once if needed — the proxy responds to everything
	// (either forwarding or blocking), so if we get a response, it's alive.
	log.Printf("[watchdog] Proxy health check OK")
}

func (pw *ProxyWatchdog) handleProxyDown(reason string) {
	pw.eventLogMu.Lock()
	*pw.eventLog = append(*pw.eventLog, SafetyEvent{
		Type:      EventProxyDown,
		Message:   "Proxy unreachable: " + reason,
		Timestamp: time.Now(),
	})
	pw.eventLogMu.Unlock()

	log.Printf("[watchdog] Proxy DOWN (%s) — stopping URLBlocker", reason)

	ub := pw.urlBlocker
	if ub != nil {
		if err := ub.Stop(); err != nil {
			log.Printf("[watchdog] URLBlocker.Stop error: %v", err)
		} else {
			log.Printf("[watchdog] URLBlocker stopped after proxy failure")

			pw.eventLogMu.Lock()
			*pw.eventLog = append(*pw.eventLog, SafetyEvent{
				Type:      EventProxyRecovery,
				Message:   "System proxy disabled after proxy failure: " + reason,
				Timestamp: time.Now(),
			})
			pw.eventLogMu.Unlock()
		}
	}

	if pw.failsafe != nil {
		pw.failsafe.Trigger("proxy-down-"+reason, pw.eventLog)
	}
}

// =============================================================================
// Comprehensive Audit Logger
// =============================================================================

type SafetyAuditLogger struct {
	events   []SafetyEvent
	mu       sync.Mutex
	maxEvents int
}

func NewSafetyAuditLogger(maxEvents int) *SafetyAuditLogger {
	if maxEvents <= 0 {
		maxEvents = 1000
	}
	return &SafetyAuditLogger{
		events:    make([]SafetyEvent, 0, maxEvents),
		maxEvents: maxEvents,
	}
}

func (sal *SafetyAuditLogger) Log(evt SafetyEvent) {
	sal.mu.Lock()
	defer sal.mu.Unlock()

	evt.Timestamp = time.Now()
	sal.events = append(sal.events, evt)

	if len(sal.events) > sal.maxEvents {
		sal.events = sal.events[len(sal.events)-sal.maxEvents:]
	}

	// Always log to standard logger
	log.Printf("[audit] %s | target=%s source=%s allowed=%v msg=%s",
		evt.Type, evt.Target, evt.Source, evt.Allowed, evt.Message)
}

func (sal *SafetyAuditLogger) GetEvents() []SafetyEvent {
	sal.mu.Lock()
	defer sal.mu.Unlock()
	result := make([]SafetyEvent, len(sal.events))
	copy(result, sal.events)
	return result
}

// =============================================================================
// Global Safety Verifier (used by blocking.go and tests)
// =============================================================================

var (
	globalSafetyVerifier   *KillSafetyVerifier
	globalSafetyVerifierMu sync.Mutex
)

func GetGlobalSafetyVerifier() *KillSafetyVerifier {
	globalSafetyVerifierMu.Lock()
	defer globalSafetyVerifierMu.Unlock()

	if globalSafetyVerifier == nil {
		globalSafetyVerifier = NewKillSafetyVerifier()
		globalSafetyVerifier.Refresh()
	}
	return globalSafetyVerifier
}

// =============================================================================
// Pre-Kill Safety Verifier
// =============================================================================

type KillSafetyVerifier struct {
	whitelist   map[string]bool
	ancestors   map[string]bool
	selfNames   map[string]bool
	systemNames map[string]bool
}

func NewKillSafetyVerifier() *KillSafetyVerifier {
	v := &KillSafetyVerifier{
		whitelist:   make(map[string]bool),
		ancestors:   make(map[string]bool),
		selfNames:   make(map[string]bool),
		systemNames: make(map[string]bool),
	}

	// Build system-protected names (without .exe extension)
	systemCritical := []string{
		"explorer", "dwm", "csrss", "winlogon", "wininit", "lsass",
		"services", "svchost", "runtimebroker", "taskhostw",
		"shellexperiencehost", "startmenuexperiencehost", "searchhost",
		"searchapp", "searchindexer", "systemsettings", "logonui",
		"lsm", "smartscreen", "applicationframehost", "textinputhost",
		"taskmgr", "ctfmon", "sihost", "conhost", "cmd", "powershell",
		"pwsh", "wt", "windowsterminal", "mstsc",
		// Linux
		"systemd", "systemd-logind", "systemd-journald", "dbus-daemon",
		"networkmanager", "wpa_supplicant", "polkitd", "udevd",
		"gnome-shell", "mutter", "kwin", "plasmashell", "Xorg",
		"Xwayland", "wayland", "pipewire", "pipewire-pulse", "pulseaudio",
		"wireplumber", "lightdm", "gdm", "sddm", "login", "sshd", "init",
		// macOS
		"Finder", "Dock", "SystemUIServer", "ControlCenter",
		"NotificationCenter", "Spotlight", "WindowManager", "WindowServer", "launchd",
	}

	for _, name := range systemCritical {
		v.systemNames[strings.ToLower(name)] = true
	}

	// Self-protection names
	self := []string{"roboty", "roboty1", "roboty-dev", "roboty.exe", "roboty1.exe", "roboty-dev.exe", "wails", "wails.exe"}
	for _, name := range self {
		v.selfNames[strings.ToLower(name)] = true
	}

	return v
}

func (v *KillSafetyVerifier) Refresh() {
	v.whitelist = make(map[string]bool)
	for k := range GetWhitelistExecs() {
		// Normalize: strip .exe to match tracker output
		normalized := strings.ToLower(strings.TrimSuffix(k, ".exe"))
		v.whitelist[normalized] = true
	}

	v.ancestors = make(map[string]bool)
	for k := range GetAncestorExecs() {
		normalized := strings.ToLower(strings.TrimSuffix(k, ".exe"))
		v.ancestors[normalized] = true
	}
}

func (v *KillSafetyVerifier) IsSafeToKill(execName string) (bool, string) {
	key := strings.TrimSuffix(strings.ToLower(execName), ".exe")

	if v.selfNames[key] {
		return false, "self-protection: would kill own process"
	}
	if v.systemNames[key] {
		return false, "system-critical process: " + key
	}
	if v.whitelist[key] {
		return false, "whitelist-protected: " + key
	}
	if v.ancestors[key] {
		return false, "ancestor-protected: " + key
	}

	return true, ""
}

// =============================================================================
// OS-level safe URL check — always allow local/loopback
// =============================================================================

var alwaysAllowedHosts = map[string]bool{
	"localhost":               true,
	"127.0.0.1":               true,
	"::1":                     true,
	"0.0.0.0":                 true,
	"127.0.0.0":               true,
	"localhost.localdomain":   true,
	"local":                   true,
}

var alwaysAllowedPrefixes = []string{
	"127.",
	"10.",
	"169.254.",
	"192.168.",
}

func isAlwaysAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))

	// Strip port only if this is not an IPv6 address (IPv6 has multiple colons)
	if strings.Count(host, ":") == 1 {
		if idx := strings.LastIndex(host, ":"); idx > 0 {
			host = host[:idx]
		}
	}

	// Strip trailing dot
	host = strings.TrimSuffix(host, ".")

	if host == "" {
		return true
	}

	if alwaysAllowedHosts[host] {
		return true
	}

	// IPv4 loopback range
	if strings.HasPrefix(host, "127.") {
		return true
	}

	// IPv6 loopback
	if host == "0:0:0:0:0:0:0:1" || host == "0:0:0:0:0:0:0:0" {
		return true
	}

	// Wails specific
	if strings.HasPrefix(host, "wails") {
		return true
	}

	return false
}

// NormalizeKillExec ensures an exec name is safe for use in taskkill/pkill
func NormalizeKillExec(execName string) string {
	name := strings.TrimSpace(execName)
	name = strings.ToLower(name)
	name = strings.TrimSuffix(name, ".exe")

	// Reject dangerous patterns
	if strings.ContainsAny(name, "&|;`$(){}[]'\"\\\x00") {
		return ""
	}
	if strings.Contains(name, "..") {
		return ""
	}
	if strings.HasPrefix(name, "-") {
		return ""
	}
	if name == "" || name == "." || name == ".." {
		return ""
	}

	// Reject Unicode homoglyph attacks (Cyrillic lookalikes)
	for _, r := range name {
		if r > 0x7F {
			return ""
		}
	}

	return name
}
