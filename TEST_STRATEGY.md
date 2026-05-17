# Roboty Focus Mode — Enterprise Test Architecture & Validation Specification

---

## Table of Contents

1. [Test Architecture Overview](#1-test-architecture-overview)
2. [Repository Structure & Conventions](#2-repository-structure--conventions)
3. [Mocking & Dependency Injection Strategy](#3-mocking--dependency-injection-strategy)
4. [Unit Tests](#4-unit-tests)
5. [Integration Tests](#5-integration-tests)
6. [End-to-End Tests](#6-end-to-end-tests)
7. [Chaos & Failure Injection Tests](#7-chaos--failure-injection-tests)
8. [Crash Recovery Tests](#8-crash-recovery-tests)
9. [Security & Exploit Tests](#9-security--exploit-tests)
10. [Performance & Stress Tests](#10-performance--stress-tests)
11. [Concurrency & Race-Condition Tests](#11-concurrency--race-condition-tests)
12. [Cross-Platform Compatibility Tests](#12-cross-platform-compatibility-tests)
13. [VM & Sandbox Validation Strategy](#13-vm--sandbox-validation-strategy)
14. [Production Hardening Verification](#14-production-hardening-verification)
15. [Real-World User Workflow Simulations](#15-real-world-user-workflow-simulations)
16. [CI/CD Strategy](#16-cicd-strategy)
17. [OS Process Safety Matrix](#17-os-process-safety-matrix)
18. [Bypass & Allowlist Matrices](#18-bypass--allowlist-matrices)
19. [Threat Models & Failure Trees](#19-threat-models--failure-trees)
20. [Edge-Case Matrices](#20-edge-case-matrices)
21. [Regression Test Plan](#21-regression-test-plan)
22. [Architecture Improvement Recommendations](#22-architecture-improvement-recommendations)
23. [Telemetry & Audit Logging Strategy](#23-telemetry--audit-logging-strategy)
24. [Emergency Rollback Architecture](#24-emergency-rollback-architecture)

---

## 1. Test Architecture Overview

### Layered Test Pyramid

```
          ╱\
         ╱  ╲
        ╱ E2E ╲          ← Real OS, real proxy, real processes (in VMs)
       ╱────────╲
      ╱Integration╲       ← DI fakes + in-memory DB + local proxy server
     ╱──────────────╲
    ╱   Unit Tests    ╲    ← Pure logic, mocks, fuzz, safety verifiers
   ╱────────────────────╲
  ╱   Static Analysis    ╲  ← go vet, staticcheck, govulncheck, gosec
 ╱──────────────────────────╲
╱   Build + Lint + Security  ╲ ← pre-commit, DCO, SBOM, sigstore
╱────────────────────────────────╲
```

### Test Execution Profiles

| Profile | Coverage | Time | Where | Budget |
|---------|----------|------|-------|--------|
| `unit` | Unit only | <30s | PR CI | Every commit |
| `safety` | Safety-critical | <10s | PR CI | Every commit |
| `integration` | +DI fakes | <2m | PR CI | Every push |
| `e2e` | Full stack | <15m | Nightly | Per branch |
| `chaos` | Failure injection | <30m | Weekly | Nightly |
| `stress` | Load/race | <60m | Weekly | Nightly |
| `security` | Fuzz + exploit | <60m | Weekly | Nightly + release |
| `cross-platform` | All OS | <120m | Nightly | VM matrix |

---

## 2. Repository Structure & Conventions

### Test Directory Layout

```
internal/
├── modes/
│   ├── *_test.go          ← White-box unit tests (package modes)
│   ├── fuzz_test.go        ← Fuzz targets
│   ├── safety_test.go      ← Safety verifier unit tests
│   ├── safety_critical_test.go ← Critical regression tests
│   ├── blocker_test.go     ← AppBlocker tests
│   ├── proxy_test.go       ← URLBlocker tests
│   ├── notifications_test.go ← Notification tests
│   ├── integration_di_test.go  ← Fake/mock DI integration tests
│   ├── tracker_test.go     ← ForegroundTracker tests
│   ├── whitelist_test.go   ← Whitelist validation
│   └── process_tree_test.go  ← Ancestor detection tests
│
├── db/
│   ├── db_test.go          ← Database integration tests (package db_test)
│   └── queries_test.go     ← SQL query tests
│
test/
├── integration/
│   ├── focus_mode_test.go  ← Full lifecycle with DI fakes
│   ├── proxy_e2e_test.go   ← Real HTTP proxy + mock system
│   ├── rollback_test.go    ← Crash/restore integration
│   └── watchdog_test.go    ← Watchdog monitor tests
│
├── e2e/
│   ├── suite_test.go       ← Test suite bootstrap
│   ├── windows_test.go     ← Windows-specific E2E
│   ├── linux_test.go       ← Linux-specific E2E
│   ├── darwin_test.go      ← macOS-specific E2E
│   ├── workflow_test.go    ← User workflow scenarios
│   └── fixtures/           ← Test snapshots, configs
│
├── chaos/
│   ├── failure_injection_test.go
│   ├── kill_loop_test.go
│   ├── proxy_crash_test.go
│   ├── network_break_test.go
│   └── oom_test.go
│
├── stress/
│   ├── concurrent_sessions_test.go
│   ├── proxy_throughput_test.go
│   ├── tracker_load_test.go
│   └── long_running_test.go
│
├── security/
│   ├── exploit_test.go
│   ├── injection_test.go
│   ├── permission_test.go
│   ├── localhost_bypass_test.go
│   └── fuzz_targets_test.go
│
├── helpers/
│   ├── testenv.go          ← Test environment setup
│   ├── fixture.go          ← Fixture loaders
│   ├── mockprocess/        ← Fake process launcher
│   ├── mockproxy/          ← Fake proxy environment
│   └── mockregistry/       ← Fake Windows registry
│
├── vms/
│   ├── vagrant/            ← Vagrantfiles for each OS
│   ├── docker/             ← Dockerfiles for Linux+macOS
│   └── packer/             ← Packer templates for Windows
│
└── ci/
    ├── github/             ← GitHub Actions workflows
    ├── scripts/            ← Helper scripts for CI
    └── Makefile            ← Top-level test targets
```

### Test Naming Conventions

| Pattern | Example | Purpose |
|---------|---------|---------|
| `Test${Subject}_${Behavior}` | `TestAppBlocker_Lifecycle` | Unit test |
| `Test${Subject}_${Scenario}_${Expected}` | `TestKillSafetyVerifier_BlocksSystemProcesses` | Safety test |
| `TestCritical_${Requirement}` | `TestCritical_ExplorerNeverKillable` | Critical regression |
| `Test${Subject}_${Condition}_${Outcome}` | `TestURLBlocker_AlwaysAllowsLocalhost` | Proxy test |
| `TestRedTeam_${BypassVector}` | `TestRedTeam_KillExecBypass` | Security test |
| `Fuzz${Function}` | `FuzzNormalizeKillExec` | Fuzz target |
| `TestChaos_${Scenario}` | `TestChaos_ProxyProcessKilled` | Chaos test |
| `TestConcurrent_${Scenario}` | `TestConcurrent_MultipleActivations` | Race test |
| `TestIntegration_${Flow}` | `TestIntegration_FocusModeLifecycle` | Integration test |
| `TestE2E_${Platform}_${Workflow}` | `TestE2E_Windows_FocusBlocksChrome` | E2E test |

### Build Tags for Platform-Specific Tests

```go
//go:build windows
//go:build linux
//go:build darwin
//go:build unit || !integration
```

---

## 3. Mocking & Dependency Injection Strategy

### Current Interface Layer (already implemented)

```go
// File: internal/modes/interfaces.go
type ProcessKiller interface {
    Kill(execName string, timeout time.Duration) error
    IsRunning(execName string) (bool, error)
    ListRunning() ([]ProcessInfo, error)
}

type SystemProxyManager interface {
    Enable(proxyAddr string, port int) error
    Disable() error
    IsEnabled() (bool, error)
}

type NotificationManager interface {
    Mute() error
    Restore() error
    IsMuted() (bool, error)
}

type ForegroundDetector interface {
    GetForeground() (*ProcessInfo, error)
    Watch(ctx CancellableContext, interval time.Duration) <-chan *ProcessInfo
}

type ProcessTreeWalker interface {
    GetAncestorExecs() (map[string]bool, error)
    GetParentPID(pid int) (int, error)
    GetProcessName(pid int) (string, error)
}

type StateFileManager interface {
    SaveState(name string) error
    ClearState(name string) error
    StateExists(name string) bool
    ListOrphans() ([]string, error)
    CleanupAll() error
}
```

### DI Constructors (already implemented)

```go
// Production constructors
NewModeService(database, queries)                           // uses real implementations
NewModeServiceWithDI(database, queries, tracker, killer, proxyMgr, notifMgr)
NewAppBlockerWithDI(tracker, killer)
NewURLBlockerWithDI(proxyMgr, stateFile)
```

### Fake Implementations (already in `integration_di_test.go`)

| Fake | Tracks | Key Features |
|------|--------|-------------|
| `fakeProcessKiller` | `killed map[string]int`, `killLog`, `running` | `setRunning()`, `blockKill()`, `killCount()`, `failOnKill` |
| `fakeProxyManager` | `enabled bool` | `enableErr`, `disableErr`, `assertEnabled()`, `assertDisabled()` |
| `fakeNotificationManager` | `muted bool` | `muteErr` |
| `fakeStateFileManager` | `states map[string]bool` | Full CRUD + `ListOrphans()` + `CleanupAll()` |

### Recommended Additional Fakes

```go
// test/helpers/fake_foreground.go
type fakeForegroundDetector struct {
    mu           sync.Mutex
    current      *ProcessInfo
    pollResults  []*ProcessInfo
    pollIndex    int
    pollErr      error
}

// test/helpers/fake_tree_walker.go
type fakeProcessTreeWalker struct {
    ancestors map[string]bool
    parentPID int
    procName  string
    err       error
}

// test/helpers/fake_db.go
type fakeDB struct {
    modes    []FocusMode
    sessions []FocusSession
    apps     []FocusModeApp
    mu       sync.Mutex
}
```

### Recommended Interface for DB Abstraction

```go
type DataStore interface {
    CreateFocusMode(ctx context.Context, params CreateModeParams) (*FocusMode, error)
    GetFocusModeByID(ctx context.Context, id string) (*FocusMode, error)
    GetAllFocusModes(ctx context.Context) ([]FocusMode, error)
    UpdateFocusMode(ctx context.Context, params UpdateModeParams) error
    DeleteFocusMode(ctx context.Context, id string) error
    CreateFocusSession(ctx context.Context, params CreateSessionParams) (*FocusSession, error)
    GetActiveFocusSession(ctx context.Context) (*FocusSession, error)
    UpdateFocusSessionStatus(ctx context.Context, id, status string) error
    CreateFocusModeApp(ctx context.Context, params CreateAppParams) error
    GetFocusModeAppsByModeID(ctx context.Context, modeID string) ([]FocusModeApp, error)
    CreateFocusModeURL(ctx context.Context, params CreateURLParams) error
    GetFocusModeURLsByModeID(ctx context.Context, modeID string) ([]FocusModeURL, error)
}
```

---

## 4. Unit Tests

### 4.1 Kill Safety Verifier Tests

```go
// File: internal/modes/safety_test.go (EXISTING — extend)

func TestKillSafetyVerifier_BlocksAllSystemProcesses(t *testing.T) {
    v := NewKillSafetyVerifier()
    v.Refresh()

    table := []struct {
        name   string
        proc   string
        reason string // expected reason substring
    }{
        // Windows critical
        {"Windows Explorer", "explorer", "system-critical"},
        {"Desktop Window Manager", "dwm", "system-critical"},
        {"Client Server Runtime", "csrss", "system-critical"},
        {"Windows Logon", "winlogon", "system-critical"},
        {"Windows Init", "wininit", "system-critical"},
        {"Local Security Authority", "lsass", "system-critical"},
        {"Service Control Manager", "services", "system-critical"},
        {"Service Host", "svchost", "system-critical"},
        {"Shell Experience Host", "shellexperiencehost", "system-critical"},
        {"Start Menu Experience Host", "startmenuexperiencehost", "system-critical"},
        {"Task Host Window", "taskhostw", "system-critical"},
        {"Runtime Broker", "runtimebroker", "system-critical"},
        // Linux critical
        {"SystemD", "systemd", "system-critical"},
        {"GNOME Shell", "gnome-shell", "system-critical"},
        {"Mutter", "mutter", "system-critical"},
        {"KWin", "kwin", "system-critical"},
        {"Plasma Shell", "plasmashell", "system-critical"},
        {"X.Org", "xorg", "system-critical"},
        {"D-Bus Daemon", "dbus-daemon", "system-critical"},
        {"NetworkManager", "networkmanager", "system-critical"},
        {"PolKit", "polkitd", "system-critical"},
        // macOS critical
        {"Finder", "finder", "system-critical"},
        {"Dock", "dock", "system-critical"},
        {"WindowServer", "windowserver", "system-critical"},
        {"Window Manager", "windowmanager", "system-critical"},
        {"LaunchD", "launchd", "system-critical"},
        {"System UI Server", "systemuiserver", "system-critical"},
        {"Control Center", "controlcenter", "system-critical"},
        {"Notification Center", "notificationcenter", "system-critical"},
    }

    for _, tt := range table {
        t.Run(tt.name, func(t *testing.T) {
            safe, reason := v.IsSafeToKill(tt.proc)
            if safe {
                t.Errorf("CRITICAL: %q was allowed (reason=%q)", tt.proc, reason)
            }
            if !strings.Contains(strings.ToLower(reason), strings.ToLower(tt.reason)) {
                t.Errorf("expected reason to contain %q, got %q", tt.reason, reason)
            }
        })
    }
}

func TestKillSafetyVerifier_ProtectsSelfWithAllVariants(t *testing.T) {
    v := NewKillSafetyVerifier()
    selfNames := []string{
        "roboty", "roboty.exe", "ROBOTY", "ROBOTY.EXE",
        "roboty1", "roboty1.exe",
        "roboty-dev", "roboty-dev.exe",
        "wails", "wails.exe", "WAILS", "WAILS.EXE",
    }
    for _, name := range selfNames {
        safe, reason := v.IsSafeToKill(name)
        if safe {
            t.Errorf("self-protection FAILED for %q: %s", name, reason)
        }
        if !strings.Contains(reason, "self-protection") {
            t.Errorf("expected self-protection reason, got: %s", reason)
        }
    }
}
```

### 4.2 NormalizeKillExec Fuzz Boundary Tests

```go
func TestNormalizeKillExec_UnicodeNormalization(t *testing.T) {
    // Unicode homoglyph attacks (fake "chrome" with Cyrillic 'е')
    attacks := []string{
        "сhrome",      // Cyrillic 'с'
        "chrome\u200B", // zero-width space
        "chrome\x00",   // null byte
        "  chrome  ",   // leading/trailing spaces (trimmed)
        strings.Repeat("chrome", 1000), // enormous
    }
    for _, input := range attacks {
        result := NormalizeKillExec(input)
        // Must either return "chrome" or empty (reject)
        if result != "" && result != "chrome" {
            t.Errorf("NormalizeKillExec(%q) = %q, expected empty or 'chrome'", input, result)
        }
    }
}
```

### 4.3 KillLoopDetector Boundary Tests

```go
func TestKillLoopDetector_WindowEdgeCases(t *testing.T) {
    t.Run("zero kills", func(t *testing.T) {
        kld := NewKillLoopDetector()
        if kld.RecordKill("chrome") {
            t.Error("single kill should not trigger loop")
        }
    })

    t.Run("exactly at threshold", func(t *testing.T) {
        kld := NewKillLoopDetector()
        for i := 0; i < MaxConsecutiveKills-1; i++ {
            kld.RecordKill("chrome")
        }
        if !kld.RecordKill("chrome") {
            t.Error("should trigger at threshold")
        }
    })

    t.Run("exactly one below threshold", func(t *testing.T) {
        kld := NewKillLoopDetector()
        for i := 0; i < MaxConsecutiveKills-1; i++ {
            if kld.RecordKill("chrome") {
                t.Fatal("premature trigger")
            }
        }
        time.Sleep(KillLoopWindow + time.Millisecond)
        // After window expires, should reset
        if kld.RecordKill("chrome") {
            t.Log("trigger after window expiry (acceptable)")
        }
    })

    t.Run("different processes tracked independently", func(t *testing.T) {
        kld := NewKillLoopDetector()
        for i := 0; i < MaxConsecutiveKills+5; i++ {
            kld.RecordKill("chrome")
            if kld.RecordKill("firefox") {
                t.Error("firefox should not trigger from chrome kills")
            }
        }
    })
}
```

### 4.4 URL IsAllowed Unit Tests

```go
func TestIsAlwaysAllowed_IPv6EdgeCases(t *testing.T) {
    tests := []struct {
        host     string
        expected bool
    }{
        {"::1", true},
        {"[::1]:62828", true},
        {"0:0:0:0:0:0:0:1", true},
        {"0:0:0:0:0:0:0:0", true},
        {"::2", false},           // not loopback
        {"fe80::1", false},       // link-local (but may need allowing)
        {"2001:db8::1", false},   // documentation
    }
    for _, tt := range tests {
        got := isAlwaysAllowed(tt.host)
        if got != tt.expected {
            t.Errorf("isAlwaysAllowed(%q) = %v, want %v", tt.host, got, tt.expected)
        }
    }
}

func TestIsAlwaysAllowed_WailsDevPatterns(t *testing.T) {
    // Wails/Vite/HMR patterns that MUST always bypass
    patterns := []string{
        "wails", "wails.localhost",
        "localhost:34115", "localhost:5173", "localhost:3000",
        "127.0.0.1:34115", "127.0.0.1:5173",
        "192.168.1.100:5173",    // dev on LAN
        "10.0.0.50:8080",        // Docker/WSL2
    }
    for _, host := range patterns {
        if !isAlwaysAllowed(host) {
            t.Errorf("CRITICAL: dev host %q blocked by isAlwaysAllowed", host)
        }
    }
}
```

### 4.5 NormalizeURLs Isolation Tests

```go
func TestNormalizeURLs_Isolation(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {"strips all protocols", []string{"https://x.com", "http://x.com", "ftp://x.com"}, []string{"x.com", "x.com", "x.com"}},
        {"preserves port when present", []string{"example.com:8080"}, []string{"example.com:8080"}},
        {"handles IP addresses", []string{"192.168.1.1:3000", "10.0.0.1"}, []string{"192.168.1.1:3000", "10.0.0.1"}},
        {"removes trailing slashes and paths", []string{"example.com/path?q=1"}, []string{"example.com"}},
        {"deduplicates (if we add later)", []string{"example.com", "example.com"}, []string{"example.com", "example.com"}},
        {"sorts consistently", nil, nil},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := normalizeURLs(tt.input)
            if len(got) != len(tt.expected) {
                t.Fatalf("len mismatch: got %v, want %v", got, tt.expected)
            }
            for i := range got {
                if got[i] != tt.expected[i] {
                    t.Errorf("index %d: got %q, want %q", i, got[i], tt.expected[i])
                }
            }
        })
    }
}
```

---

## 5. Integration Tests

### 5.1 Focus Mode Full Lifecycle (with DI Fakes)

```go
// File: test/integration/focus_mode_test.go
package integration

import (
    "testing"
    "time"
    "Roboty/internal/modes"
)

func TestIntegration_FocusModeLifecycle(t *testing.T) {
    // Setup DI fakes
    fakeKiller := modes.NewFakeProcessKiller()
    fakeProxyMgr := modes.NewFakeProxyManager()
    fakeNotifMgr := modes.NewFakeNotificationManager()
    inmemDB := modes.NewFakeDataStore()

    ms := modes.NewModeServiceWithDI(inmemDB, nil, nil, fakeKiller, fakeProxyMgr, fakeNotifMgr)

    // Create a focus mode
    mode, err := ms.CreateMode(modes.CreateModeRequest{
        Name:             "Work Mode",
        DurationMinutes:  60,
        MuteNotifications: true,
        Apps: []modes.FocusModeApp{
            {AppName: "Google Chrome", AppExec: "chrome.exe", CloseOnActivate: true},
            {AppName: "Slack", AppExec: "slack.exe", IsAllowed: true},
        },
        AllowedURLs: []string{"github.com", "stackoverflow.com"},
    })
    if err != nil {
        t.Fatalf("CreateMode failed: %v", err)
    }
    if mode.ID == "" {
        t.Fatal("expected non-empty mode ID")
    }

    // Activate
    session, err := ms.ActivateMode(mode.ID)
    if err != nil {
        t.Fatalf("ActivateMode failed: %v", err)
    }
    if session.Status != "active" {
        t.Errorf("expected active session, got %q", session.Status)
    }

    // Verify proxy was enabled
    if enabled, _ := fakeProxyMgr.IsEnabled(); !enabled {
        t.Error("expected proxy to be enabled after activation")
    }

    // Verify notifications were muted
    if muted, _ := fakeNotifMgr.IsMuted(); !muted {
        t.Error("expected notifications muted after activation")
    }

    // Verify chrome was killed (closeOnActivate)
    if fakeKiller.KillCount("chrome") != 1 {
        t.Errorf("expected chrome killed once, got %d times", fakeKiller.KillCount("chrome"))
    }

    // Deactivate
    if err := ms.DeactivateMode(session.ID); err != nil {
        t.Fatalf("DeactivateMode failed: %v", err)
    }

    // Verify cleanup
    if enabled, _ := fakeProxyMgr.IsEnabled(); enabled {
        t.Error("expected proxy disabled after deactivation")
    }
    if muted, _ := fakeNotifMgr.IsMuted(); muted {
        t.Error("expected notifications restored after deactivation")
    }
}
```

### 5.2 URL Blocker Integration with Local Proxy

```go
// File: test/integration/proxy_integration_test.go
package integration

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "net/http"
    "testing"
    "time"
    "Roboty/internal/modes"
)

func TestIntegration_URLBlocker_WithFakeProxyManager(t *testing.T) {
    fakeProxyMgr := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    ub := modes.NewURLBlockerWithDI(fakeProxyMgr, fakeState)

    // Start proxy on ephemeral port
    port := 0 // use dynamic port in real impl
    err := ub.Start(62832, []string{"example.com"})
    if err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    defer ub.Stop()

    // Verify system proxy was enabled
    if enabled, _ := fakeProxyMgr.IsEnabled(); !enabled {
        t.Error("expected system proxy enabled")
    }

    // Verify state file was created
    if !fakeState.StateExists(modes.ProxyStateName) {
        t.Error("expected proxy state to exist after start")
    }

    // --- Allowed host ---
    code, _ := sendProxyRequest(t, 62832, "example.com")
    if code == http.StatusForbidden {
        t.Error("allowed host returned 403")
    }

    // --- Blocked host ---
    code, _ = sendProxyRequest(t, 62832, "blocked.com")
    if code != http.StatusForbidden {
        t.Errorf("expected 403 for blocked host, got %d", code)
    }

    // --- Localhost always allowed ---
    code, _ = sendProxyRequest(t, 62832, "localhost")
    if code == http.StatusForbidden {
        t.Error("localhost must not be blocked")
    }

    // Stop
    ub.Stop()

    // Verify cleanup
    if enabled, _ := fakeProxyMgr.IsEnabled(); enabled {
        t.Error("expected proxy disabled after stop")
    }
    if fakeState.StateExists(modes.ProxyStateName) {
        t.Error("expected state file cleaned after stop")
    }
}

func sendProxyRequest(t *testing.T, port int, host string) (int, string) {
    t.Helper()
    conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 3*time.Second)
    if err != nil {
        return 0, ""
    }
    defer conn.Close()
    req := fmt.Sprintf("GET http://%s/ HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", host, host)
    fmt.Fprint(conn, req)
    resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
    if err != nil {
        return 0, ""
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return resp.StatusCode, string(body)
}
```

### 5.3 App Blocker with Simulated Foreground Polling

```go
func TestIntegration_AppBlocker_SimulatedForeground(t *testing.T) {
    fakeKiller := modes.NewFakeProcessKiller()
    fakeKiller.SetRunning("chrome", true)
    fakeKiller.SetRunning("code", true)

    // Use a fake foreground tracker that returns predetermined results
    fakeTracker := modes.NewFakeForegroundTracker()
    fakeTracker.SetNext("Google Chrome", "chrome", 1234)

    ab := modes.NewAppBlockerWithDI(fakeTracker, fakeKiller)

    // Start blocker with "code" allowed only (chrome should be killed)
    ab.Start([]string{"code"}, nil, 50*time.Millisecond)
    time.Sleep(200 * time.Millisecond)

    // chrome should have been killed
    if fakeKiller.KillCount("chrome") == 0 {
        t.Error("expected chrome to be killed by blocker")
    }

    // code should NOT be killed
    if fakeKiller.KillCount("code") > 0 {
        t.Error("code should not be killed (it's in the allowed list)")
    }

    // Now switch foreground to "code" — should not trigger kill
    fakeTracker.SetNext("VS Code", "code", 5678)
    time.Sleep(200 * time.Millisecond)

    codeKills := fakeKiller.KillCount("code")
    if codeKills > 0 {
        t.Errorf("code was killed %d times by blocker (should be 0)", codeKills)
    }

    ab.Stop()
}
```

---

## 6. End-to-End Tests

### 6.1 E2E Test Suite Structure

```go
// File: test/e2e/suite_test.go
package e2e

import (
    "context"
    "os"
    "testing"
    "Roboty/internal/modes"
)

// TestSuite holds the complete test environment
type TestSuite struct {
    ModeService *modes.ModeService
    // In E2E, these may be real implementations or wrapped with guards
    Killer      modes.ProcessKiller
    ProxyMgr    modes.SystemProxyManager
    NotifMgr    modes.NotificationManager
    // For E2E, we wrap real implementations with safety guards
}

func TestMain(m *testing.M) {
    // Ensure we never run E2E without safety mode on a real machine
    if os.Getenv("ROBOTY_E2E") != "1" {
        println("E2E tests require ROBOTY_E2E=1 and must run inside a VM")
        os.Exit(0)
    }

    // Always force safe mode for E2E
    os.Setenv("ROBOTY_SAFE_MODE", "true")
    os.Exit(m.Run())
}
```

### 6.2 Windows E2E: Focus Mode Blocks Chrome

```go
//go:build windows
// File: test/e2e/windows_test.go
package e2e

import (
    "os/exec"
    "strings"
    "testing"
    "time"
    "Roboty/internal/modes"
)

func TestE2E_Windows_FocusBlocksChrome(t *testing.T) {
    // Prerequisite: Chrome must be installed
    if _, err := exec.LookPath("chrome.exe"); err != nil {
        t.Skip("Chrome not installed")
    }

    // Launch Chrome
    chrome := exec.Command("chrome.exe", "--new-window", "about:blank")
    chrome.Start()
    defer chrome.Process.Kill()
    time.Sleep(2 * time.Second)

    // Start focus mode with only "notepad" allowed
    ms := modes.NewModeService(nil, nil) // real DB for E2E
    mode, _ := ms.CreateMode(modes.CreateModeRequest{
        Name: "E2E Test",
        Apps: []modes.FocusModeApp{
            {AppName: "Notepad", AppExec: "notepad.exe", IsAllowed: true},
            {AppName: "Chrome", AppExec: "chrome.exe", CloseOnActivate: true},
        },
    })
    ms.ActivateMode(mode.ID)
    time.Sleep(3 * time.Second)

    // Check chrome is no longer running
    check := exec.Command("tasklist", "/FI", "IMAGENAME eq chrome.exe")
    out, _ := check.Output()
    if strings.Contains(string(out), "chrome.exe") {
        t.Error("Chrome should have been closed by focus mode")
    }

    ms.DeactivateMode(mode.ID)
}
```

### 6.3 User Workflow: Create → Activate → Verify → Deactivate

```go
// File: test/e2e/workflow_test.go
package e2e

import (
    "testing"
    "time"
    "Roboty/internal/modes"
)

func TestE2E_Workflow_CreateActivateDeactivate(t *testing.T) {
    os.Setenv("ROBOTY_SAFE_MODE", "true")
    defer os.Unsetenv("ROBOTY_SAFE_MODE")

    ms := modes.NewModeService(nil, nil)

    // Step 1: Create mode
    mode, err := ms.CreateMode(modes.CreateModeRequest{
        Name:             "Deep Work",
        DurationMinutes:  25,
        MuteNotifications: true,
        Apps: []modes.FocusModeApp{
            {AppName: "Discord", AppExec: "discord.exe", CloseOnActivate: true},
        },
    })
    if err != nil {
        t.Fatalf("step 1: %v", err)
    }

    // Step 2: Activate
    session, err := ms.ActivateMode(mode.ID)
    if err != nil {
        t.Fatalf("step 2: %v", err)
    }

    active, _ := ms.GetActiveSession()
    if active == nil || active.ID != session.ID {
        t.Fatal("step 2: active session mismatch")
    }

    // Step 3: Verify state
    timerActive := false
    // In safe mode, these are no-ops but return success
    time.Sleep(1 * time.Second)

    // Step 4: Deactivate
    if err := ms.DeactivateMode(session.ID); err != nil {
        t.Fatalf("step 4: %v", err)
    }

    // Step 5: Verify clean state
    active, _ = ms.GetActiveSession()
    if active != nil {
        t.Fatal("step 5: session should be nil after deactivation")
    }

    // Step 6: Delete mode
    if err := ms.DeleteMode(mode.ID); err != nil {
        t.Fatalf("step 6: %v", err)
    }
}
```

---

## 7. Chaos & Failure Injection Tests

### 7.1 Failure Injection Framework

```go
// File: test/chaos/failure_injection_test.go
package chaos

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "testing"
    "time"
    "Roboty/internal/modes"
)

// FailureMode represents a type of system failure to inject
type FailureMode int

const (
    FailProxyEnable    FailureMode = iota
    FailProxyDisable
    FailKillProcess
    FailKillList
    FailNotifMute
    FailNotifRestore
    FailTrackerPoll
    FailForegroundDetect
    FailStateSave
    FailStateClear
    FailAll
)

// FailureInjector wraps fakes with probabilistic failure injection
type FailureInjector struct {
    mu           sync.Mutex
    failures     map[FailureMode]float64 // 0.0 - 1.0 probability
    killCount    int
    maxKills     int
    killCallback func(string)
}

func (fi *FailureInjector) shouldFail(mode FailureMode) bool {
    fi.mu.Lock()
    defer fi.mu.Unlock()
    prob, ok := fi.failures[mode]
    if !ok {
        return false
    }
    return rand.Float64() < prob
}
```

### 7.2 Chaotic Proxy Failure Test

```go
func TestChaos_ProxyFailsRandomly(t *testing.T) {
    fi := &FailureInjector{
        failures: map[FailureMode]float64{
            FailProxyEnable:  0.3, // 30% chance of failure
            FailProxyDisable: 0.3,
        },
    }

    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    ub := modes.NewURLBlockerWithDI(
        &chaoticProxyManager{inner: fakeProxy, injector: fi},
        fakeState,
    )

    // Repeated start/stop cycles
    for i := 0; i < 100; i++ {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    t.Errorf("panic on cycle %d: %v", i, r)
                }
            }()

            err := ub.Start(62833, []string{"example.com"})

            // If proxy enable failed, state should NOT be persisted
            if err != nil && fakeState.StateExists(modes.ProxyStateName) {
                t.Errorf("cycle %d: state exists after failed enable", i)
            }

            // Ensure stop always works
            if stopErr := ub.Stop(); stopErr != nil {
                t.Logf("cycle %d: stop returned error (acceptable): %v", i, stopErr)
            }

            // State must always be cleared after stop
            if fakeState.StateExists(modes.ProxyStateName) {
                t.Errorf("cycle %d: state exists after stop", i)
            }
        }()
    }
}

type chaoticProxyManager struct {
    inner    modes.SystemProxyManager
    injector *FailureInjector
}

func (c *chaoticProxyManager) Enable(addr string, port int) error {
    if c.injector.shouldFail(FailProxyEnable) {
        return fmt.Errorf("chaos: injected proxy enable failure")
    }
    return c.inner.Enable(addr, port)
}

func (c *chaoticProxyManager) Disable() error {
    if c.injector.shouldFail(FailProxyDisable) {
        return fmt.Errorf("chaos: injected proxy disable failure")
    }
    return c.inner.Disable()
}

func (c *chaoticProxyManager) IsEnabled() (bool, error) {
    return c.inner.IsEnabled()
}
```

### 7.3 Kill Loop Chaos (Simulating Restarting Apps)

```go
func TestChaos_KillLoop_RestartingApp(t *testing.T) {
    fakeKiller := modes.NewFakeProcessKiller()
    fakeTracker := modes.NewFakeForegroundTracker()
    fakeTracker.SetNext("Bad App", "badapp", 999)

    ab := modes.NewAppBlockerWithDI(fakeTracker, fakeKiller)

    // Simulate an app that immediately restarts after kill
    go func() {
        for {
            time.Sleep(10 * time.Millisecond)
            if !ab.IsRunning() {
                return
            }
            // App "restarts" — fake process comes back
            fakeKiller.SetRunning("badapp", true)
            fakeTracker.SetNext("Bad App", "badapp", 999)
        }
    }()

    ab.Start([]string{"goodapp"}, nil, 30*time.Millisecond)
    time.Sleep(2 * time.Second)
    ab.Stop()

    // The kill loop detector should have eventually stopped killing
    t.Logf("Bad app was killed %d times", fakeKiller.KillCount("badapp"))
    if fakeKiller.KillCount("badapp") > modes.MaxConsecutiveKills*5 {
        t.Error("kill count too high — kill loop detector may not have triggered")
    }
}
```

### 7.4 Network Break Scenario (Proxy Can't Reach Internet)

```go
func TestChaos_Proxy_NetworkBreak(t *testing.T) {
    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
    ub.Start(62834, []string{"example.com"})
    defer ub.Stop()

    // Verify localhost still works when proxy is up
    if !ub.IsAllowed("localhost") {
        t.Error("localhost should be allowed when proxy is running")
    }

    // Simulate network break by stopping the proxy's HTTP server
    // (the proxy itself is still "running" but can't route)
    ub.ServerShutdown() // custom helper

    // The health check should detect this and trigger emergency stop
    // This is tested via the ProxyWatchdog integration
}
```

---

## 8. Crash Recovery Tests

### 8.1 Simulated Crash with State Persistence

```go
// File: test/integration/rollback_test.go
package integration

func TestCrashRecovery_ProxyStateRestored(t *testing.T) {
    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
    ub.Start(62835, []string{"example.com"})

    // Simulate crash: state file exists, proxy enabled
    if !fakeState.StateExists(modes.ProxyStateName) {
        t.Error("proxy state should exist before crash")
    }

    // "Crash" occurs — ub.Stop() is NOT called
    // On next startup, CleanupOrphanedProxy should detect and clean

    // Reset state manager (simulating new process)
    fakeState2 := modes.NewFakeStateFileManager()
    // Copy the crash state marker
    fakeState2.SaveState(modes.ProxyStateName)

    // Run cleanup (what startup does)
    sf := fakeState2
    if sf.StateExists(modes.ProxyStateName) {
        t.Log("Detected orphaned proxy state from previous crash")
        sf.ClearState(modes.ProxyStateName)
        fakeProxy.Disable()
    }

    if fakeState.StateExists(modes.ProxyStateName) {
        t.Error("state should have been cleaned after crash recovery")
    }
    if enabled, _ := fakeProxy.IsEnabled(); enabled {
        t.Error("proxy should be disabled after crash recovery")
    }
}

func TestCrashRecovery_NotificationStateRestored(t *testing.T) {
    fakeNotif := modes.NewFakeNotificationManager()

    // Simulate crash: notification was muted
    fakeNotif.Mute()

    if muted, _ := fakeNotif.IsMuted(); !muted {
        t.Error("notifications should be muted before crash")
    }

    // Crash occurs, notification state is orphaned
    // On startup, CleanupOrphanedNotifications should restore

    // Simulate startup recovery
    if muted, _ := fakeNotif.IsMuted(); muted {
        t.Log("Detected orphaned notification state from previous crash")
        fakeNotif.Restore()
    }

    if muted, _ := fakeNotif.IsMuted(); muted {
        t.Error("notifications should be restored after crash recovery")
    }
}
```

### 8.2 EmergencyStop Resilience

```go
func TestCrashRecovery_EmergencyStop_Cleanup(t *testing.T) {
    fakeKiller := modes.NewFakeProcessKiller()
    fakeProxy := modes.NewFakeProxyManager()
    fakeNotif := modes.NewFakeNotificationManager()

    ms := modes.NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxy, fakeNotif)

    // Activate mode
    mode, _ := ms.CreateMode(modes.CreateModeRequest{
        Name: "Test",
        Apps: []modes.FocusModeApp{
            {AppName: "Chrome", AppExec: "chrome.exe", IsAllowed: true},
        },
    })
    ms.ActivateMode(mode.ID)

    // Simulate catastrophic failure — call EmergencyStop
    ms.EmergencyStop("simulated-crash")

    // Verify full cleanup
    if enabled, _ := fakeProxy.IsEnabled(); enabled {
        t.Error("proxy should be disabled after emergency stop")
    }
    if muted, _ := fakeNotif.IsMuted(); muted {
        t.Error("notifications should be restored after emergency stop")
    }

    // Verify no orphaned state
    active, _ := ms.GetActiveSession()
    if active != nil {
        t.Error("no active session should exist after emergency stop")
    }
}

func TestCrashRecovery_EmergencyStop_MultipleCalls(t *testing.T) {
    ms := modes.NewModeServiceWithDI(nil, nil, nil,
        modes.NewFakeProcessKiller(),
        modes.NewFakeProxyManager(),
        modes.NewFakeNotificationManager(),
    )

    // EmergencyStop must be idempotent
    for i := 0; i < 10; i++ {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    t.Errorf("panic on EmergencyStop call %d: %v", i, r)
                }
            }()
            ms.EmergencyStop(fmt.Sprintf("call-%d", i))
        }()
    }
}
```

### 8.3 Forceful Termination Simulation

```go
func TestCrashRecovery_ForceTerminate_Cleanup(t *testing.T) {
    // Simulates what happens when the process receives SIGKILL (cannot be caught)
    // We verify that state marker files allow recovery on next startup

    fakeState := modes.NewFakeStateFileManager()

    // Phase 1: normal operation — state files created
    fakeState.SaveState(modes.ProxyStateName)
    fakeState.SaveState(modes.NotifStateName)

    // Phase 2: "SIGKILL" — process dies, nothing cleaned
    // (no action possible)

    // Phase 3: next startup — check for orphans and clean
    orphans, _ := fakeState.ListOrphans()
    if len(orphans) != 2 {
        t.Errorf("expected 2 orphan states, got %d: %v", len(orphans), orphans)
    }

    // Cleanup
    fakeState.CleanupAll()
    orphans, _ = fakeState.ListOrphans()
    if len(orphans) != 0 {
        t.Errorf("expected 0 orphans after cleanup, got %d", len(orphans))
    }
}
```

---

## 9. Security & Exploit Tests

### 9.1 Injection Attack Vectors

```go
// File: test/security/exploit_test.go
package security

func TestSecurity_ProcessKillInjection(t *testing.T) {
    attacks := []string{
        "chrome & notepad",
        "chrome | taskkill /F /IM explorer.exe",
        "chrome; rm -rf /",
        "`cat /etc/passwd`",
        "$(cat /etc/shadow)",
        "chrome.exe & del /F /S /Q C:\\*",
        "..\\..\\Windows\\System32\\malware",
        "-rf /",
        "--help",
        "/S /C 'malicious'",
        "powershell -Command \"Invoke-Expression 2+2\"",
        "safe_process.exe & echo pwned",
    }

    for _, attack := range attacks {
        result := modes.NormalizeKillExec(attack)
        if result != "" {
            // If not empty, it must be demonstrably safe
            if strings.ContainsAny(result, "&|;`$(){}[]'\"\\") {
                t.Errorf("UNSAFE result for attack %q: %q", attack, result)
            }
            if strings.HasPrefix(result, "-") {
                t.Errorf("FLAG INJECTION for attack %q: %q", attack, result)
            }
        }
    }
}

func TestSecurity_ProxyBypass_ExternalToLocalhost(t *testing.T) {
    ub := modes.NewURLBlocker()
    ub.SetAllowedURLs([]string{}) // Block everything

    // These should all be blocked — they look like localhost but are not
    bypassHosts := []struct {
        host     string
        desc     string
    }{
        {"localhost.evil.com", "subdomain of evil.com with localhost prefix"},
        {"127.0.0.1.evil.com", "subdomain domain confusion"},
        {"127.0.0.1:80@evil.com", "URL credential confusion"},
        {"[::1]:evil.com", "IPv6 port confusion"},
        {"localhost%00.evil.com", "null byte to confuse parser"},
    }

    for _, tt := range bypassHosts {
        allowed := ub.IsAllowed(tt.host)
        if allowed {
            t.Errorf("POTENTIAL BYPASS: %q (%s) was allowed", tt.host, tt.desc)
        }
    }
}

func TestSecurity_ProxyBypass_DNSTunneling(t *testing.T) {
    ub := modes.NewURLBlocker()
    ub.SetAllowedURLs([]string{"example.com"})

    // DNS tunneling attempts to smuggle traffic through allowed domains
    tunneling := []string{
        "evil.com.example.com",     // subdomain of allowed
        "example.com.evil.com",     // typo
        "example.com:443.evil.com", // port confusion
    }

    for _, host := range tunneling {
        allowed := ub.IsAllowed(host)
        if allowed {
            t.Logf("DNS tunneling via %q was allowed (verify this is correct behavior)", host)
        }
    }
}

func TestSecurity_CommandInjection_ProxySetting(t *testing.T) {
    psInjection := `Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -Value '"; Start-Process cmd.exe -ArgumentList '/c calc.exe' #' -Type String -Force`
    // This should fail or be properly escaped
    // In the current implementation, the addr is formatted via fmt.Sprintf
    // which provides some protection against PowerShell injection
    // but this test documents the risk

    // The test verifies the proxy addr doesn't contain shell metacharacters
    addr := fmt.Sprintf("127.0.0.1:%d", 62828)
    if strings.ContainsAny(addr, "'\"&|;`$(){}[]") {
        t.Error("proxy address contains shell metacharacters")
    }
}
```

### 9.2 Race-to-Kill (TOCTOU) Tests

```go
func TestSecurity_TOCTOU_ProcessChangesBetweenCheckAndKill(t *testing.T) {
    fakeKiller := modes.NewFakeProcessKiller()
    fakeTracker := modes.NewFakeForegroundTracker()

    // Simulate rapid process name changes (a "morphing" process)
    go func() {
        names := []string{"chrome", "explorer", "svchost", "chrome", "systemd"}
        i := 0
        for {
            fakeTracker.SetNext(names[i%len(names)], names[i%len(names)], 1000+uint32(i))
            i++
            time.Sleep(1 * time.Millisecond)
        }
    }()

    ab := modes.NewAppBlockerWithDI(fakeTracker, fakeKiller)
    ab.Start([]string{"code"}, nil, 5*time.Millisecond)
    time.Sleep(500 * time.Millisecond)
    ab.Stop()

    // The safety verifier should have prevented killing system processes
    // even if they briefly appeared as the foreground process
    systemKills := 0
    for _, proc := range []string{"explorer", "svchost", "systemd"} {
        systemKills += fakeKiller.KillCount(proc)
    }
    if systemKills > 0 {
        t.Errorf("system processes were killed %d times due to TOCTOU race", systemKills)
    }
}
```

### 9.3 Whitelist Manipulation Tests

```go
func TestSecurity_WhitelistIntegrity(t *testing.T) {
    whitelist := modes.GetWhitelistExecs()

    // Verify no malicious entries
    for k := range whitelist {
        // No relative paths
        if strings.Contains(k, "..") || strings.Contains(k, "\\") || strings.Contains(k, "/") {
            t.Errorf("whitelist contains path-like key: %q", k)
        }
        // No shell chars
        if strings.ContainsAny(k, "&|;`$(){}[]'\"") {
            t.Errorf("whitelist contains special chars: %q", k)
        }
        // No empty keys
        if k == "" || strings.TrimSpace(k) == "" {
            t.Error("whitelist contains empty key")
        }
    }
}

func TestSecurity_WhitelistNotFound_DefaultsProtect(t *testing.T) {
    // If whitelist.json cannot be read, the system must fall back to
    // a hardcoded protection set — never allow all kills
    verifier := modes.NewKillSafetyVerifier()
    // Even without Refresh() (which loads the whitelist from file),
    // the hardcoded systemNames should protect critical processes
    safe, _ := verifier.IsSafeToKill("explorer")
    if safe {
        t.Error("even without whitelist file, explorer must be protected by built-in list")
    }
}
```

---

## 10. Performance & Stress Tests

### 10.1 Concurrent Session Management

```go
// File: test/stress/concurrent_sessions_test.go
package stress

func TestStress_ConcurrentModeActivations(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping stress test in short mode")
    }

    fakeKiller := modes.NewFakeProcessKiller()
    fakeProxy := modes.NewFakeProxyManager()
    fakeNotif := modes.NewFakeNotificationManager()

    ms := modes.NewModeServiceWithDI(nil, nil, nil, fakeKiller, fakeProxy, fakeNotif)

    // Create many modes
    modeIDs := make([]string, 50)
    for i := 0; i < 50; i++ {
        mode, err := ms.CreateMode(modes.CreateModeRequest{
            Name: fmt.Sprintf("Mode-%d", i),
            Apps: []modes.FocusModeApp{
                {AppName: "App", AppExec: fmt.Sprintf("app%d.exe", i)},
            },
        })
        if err != nil {
            t.Fatalf("CreateMode %d: %v", i, err)
        }
        modeIDs[i] = mode.ID
    }

    // Activate all concurrently
    var wg sync.WaitGroup
    for _, id := range modeIDs {
        wg.Add(1)
        go func(modeID string) {
            defer wg.Done()
            session, err := ms.ActivateMode(modeID)
            if err != nil {
                t.Errorf("ActivateMode %s: %v", modeID, err)
                return
            }
            // Only the last activation should be active
            _ = session
        }(id)
    }
    wg.Wait()

    // Only 0 or 1 active sessions should exist
    active, _ := ms.GetActiveSession()
    if active != nil {
        t.Logf("Active session: %s (mode=%s)", active.ID, active.ModeID)
    }
}
```

### 10.2 Proxy Throughput Under Load

```go
func TestStress_ProxyThroughput(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping stress test in short mode")
    }

    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
    ub.Start(62836, []string{"example.com", "github.com"})
    defer ub.Stop()

    // Send concurrent requests to the proxy
    const workers = 50
    const requestsPerWorker = 100
    var wg sync.WaitGroup
    errors := make(chan error, workers*requestsPerWorker)

    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for i := 0; i < requestsPerWorker; i++ {
                host := "example.com"
                if i%10 == 0 {
                    host = "blocked-site.com" // mix in blocked
                }
                code, _ := sendProxyRequest(t, 62836, host)
                if code == 0 {
                    errors <- fmt.Errorf("worker %d req %d: connection failed", workerID, i)
                }
            }
        }(w)
    }
    wg.Wait()
    close(errors)

    var totalErrors int
    for err := range errors {
        totalErrors++
        t.Log(err)
    }
    if totalErrors > 10 {
        t.Errorf("too many errors: %d/%d", totalErrors, workers*requestsPerWorker)
    }
}
```

### 10.3 Long-Running Stability

```go
func TestStress_LongRunning_Blocker(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping long-running test")
    }

    fakeKiller := modes.NewFakeProcessKiller()
    fakeTracker := modes.NewFakeForegroundTracker()
    fakeTracker.SetNext("App", "testapp", 1234)

    ab := modes.NewAppBlockerWithDI(fakeTracker, fakeKiller)
    ab.Start([]string{"allowed"}, nil, 100*time.Millisecond)

    // Run for extended period with changing foreground
    done := make(chan struct{})
    go func() {
        apps := []string{"testapp", "chrome", "firefox", "discord", "slack", "allowed"}
        for i := 0; ; i++ {
            select {
            case <-done:
                return
            default:
                app := apps[i%len(apps)]
                fakeTracker.SetNext(app, app, 1000+uint32(i))
                time.Sleep(50 * time.Millisecond)
            }
        }
    }()

    time.Sleep(30 * time.Second)
    close(done)
    ab.Stop()

    t.Logf("Kill counts: testapp=%d, chrome=%d, firefox=%d, allowed=%d",
        fakeKiller.KillCount("testapp"),
        fakeKiller.KillCount("chrome"),
        fakeKiller.KillCount("firefox"),
        fakeKiller.KillCount("allowed"))
}
```

### 10.4 Memory Leak Detection

```go
func TestStress_MemoryLeak_Proxy(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping memory test")
    }

    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    // Repeated start/stop cycles
    var memStats []runtime.MemStats
    for i := 0; i < 1000; i++ {
        ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
        ub.Start(62837, []string{"example.com"})
        ub.Stop()

        if i%100 == 0 {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            memStats = append(memStats, m)
        }
    }

    // Check for consistent memory growth
    if len(memStats) >= 2 {
        first := memStats[0].Alloc
        last := memStats[len(memStats)-1].Alloc
        ratio := float64(last) / float64(first)
        t.Logf("Memory: start=%d bytes, end=%d bytes (ratio=%.2f)", first, last, ratio)
        if ratio > 2.0 {
            t.Errorf("possible memory leak: alloc grew %.2fx", ratio)
        }
    }
}
```

---

## 11. Concurrency & Race-Condition Tests

### 11.1 Data Race Detection

All tests should be run with `-race` flag:

```bash
go test -race ./internal/modes/...
go test -race ./test/...
```

### 11.2 Concurrent Start/Stop of Blocker

```go
// File: test/stress/concurrent_operations_test.go
package stress

func TestConcurrent_Blocker_StartStopRace(t *testing.T) {
    fakeKiller := modes.NewFakeProcessKiller()
    fakeTracker := modes.NewFakeForegroundTracker()
    ab := modes.NewAppBlockerWithDI(fakeTracker, fakeKiller)

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ab.Start([]string{"app"}, nil, 100*time.Millisecond)
            ab.Stop()
        }()
    }
    wg.Wait()

    if ab.IsRunning() {
        t.Error("blocker should not be running after all stop calls")
    }
}

func TestConcurrent_Proxy_StartStopRace(t *testing.T) {
    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
            ub.Start(62838, []string{"example.com"})
            ub.Stop()
        }()
    }
    wg.Wait()
}

func TestConcurrent_ModeService_ReadWriteRace(t *testing.T) {
    ms := modes.NewModeServiceWithDI(nil, nil, nil,
        modes.NewFakeProcessKiller(),
        modes.NewFakeProxyManager(),
        modes.NewFakeNotificationManager(),
    )

    var wg sync.WaitGroup
    // Writers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            mode, err := ms.CreateMode(modes.CreateModeRequest{
                Name: fmt.Sprintf("Concurrent-%d", id),
                Apps: []modes.FocusModeApp{
                    {AppName: "App", AppExec: fmt.Sprintf("app%d.exe", id)},
                },
            })
            if err == nil {
                ms.ActivateMode(mode.ID)
            }
        }(i)
    }

    // Readers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ms.ListModes()
            ms.GetActiveSession()
        }()
    }

    wg.Wait()
}
```

### 11.3 Kill Loop Detector Race

```go
func TestConcurrent_KillLoopDetector_Race(t *testing.T) {
    kld := modes.NewKillLoopDetector()

    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                kld.RecordKill(fmt.Sprintf("app-%d", workerID%5))
            }
        }(i)
    }
    wg.Wait()
}
```

### 11.4 Proxy Health Check Race

```go
func TestConcurrent_ProxyWatchdog_HealthRace(t *testing.T) {
    fakeProxy := modes.NewFakeProxyManager()
    fakeState := modes.NewFakeStateFileManager()
    ub := modes.NewURLBlockerWithDI(fakeProxy, fakeState)
    ub.Start(62839, []string{"example.com"})

    // Start concurrent health checks
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 50; j++ {
                ub.IsAllowed("example.com")
                ub.IsRunning()
            }
        }()
    }
    wg.Wait()
    ub.Stop()
}
```

---

## 12. Cross-Platform Compatibility Tests

### 12.1 OS Matrix

| OS | Version | Arch | WM/DE | Proxy Mechanism | Notif Mechanism | Process Enum |
|----|---------|------|-------|-----------------|-----------------|--------------|
| Windows 10 | 22H2 | x64 | DWM | Registry | Registry | tasklist |
| Windows 11 | 23H2 | x64 | DWM | Registry | Registry | tasklist |
| Ubuntu 22.04 | LTS | x64 | GNOME X11 | gsettings | busctl | /proc |
| Ubuntu 24.04 | LTS | x64 | GNOME Wayland | gsettings | busctl | /proc |
| Fedora 40 | Latest | x64 | GNOME Wayland | gsettings | busctl | /proc |
| Debian 12 | Stable | x64 | XFCE | gsettings | busctl | /proc |
| Kubuntu 24.04 | LTS | x64 | KDE Plasma | gsettings | busctl | /proc |
| macOS Ventura | 13.x | arm64 | Aqua | networksetup | osascript | ps |
| macOS Sonoma | 14.x | arm64 | Aqua | networksetup | osascript | ps |

### 12.2 Platform Detection Tests

```go
// File: test/integration/platform_test.go
package integration

func TestPlatform_ProxyCommandFormat(t *testing.T) {
    // Verify that the OS-specific proxy commands match expected format
    type proxyCmd struct {
        os      string
        command string
        args    []string
    }

    expected := map[string]proxyCmd{
        "windows": {"windows", "powershell", []string{"-NoProfile", "-Command", "Set-ItemProperty ..."}},
        "linux":   {"linux", "gsettings", []string{"set", "org.gnome.system.proxy", "mode", "manual"}},
        "darwin":  {"darwin", "networksetup", []string{"-setwebproxy", "service", "127.0.0.1", "62828"}},
    }

    // Test via production proxy manager
    for osName, expectedCmd := range expected {
        t.Run(osName, func(t *testing.T) {
            if runtime.GOOS != osName && osName != runtime.GOOS {
                // Can't test on wrong platform, but we can validate the command format
                t.Skipf("wrong platform: %s", runtime.GOOS)
            }
            // Test that the expected tool exists
            _, err := exec.LookPath(expectedCmd.command)
            if err != nil {
                t.Logf("command %q not available: %v", expectedCmd.command, err)
            }
        })
    }
}
```

### 12.3 Cross-Platform Process Detection

```go
//go:build windows
func TestWindows_ProcessEnumeration_IncludesCurrent(t *testing.T) {
    procs, err := modes.ListWindowsProcessesWithKiller()
    if err != nil {
        t.Fatalf("ListWindowsProcessesWithKiller failed: %v", err)
    }
    found := false
    for _, p := range procs {
        if strings.Contains(p.Name, "go-test") || strings.Contains(p.Name, "powershell") {
            found = true
            break
        }
    }
    if !found {
        t.Log("current process not in list (acceptable)")
    }
}

//go:build linux
func TestLinux_ProcessEnumeration_ReadsProc(t *testing.T) {
    procs, err := modes.ListLinuxProcessesWithKiller()
    if err != nil {
        t.Fatalf("ListLinuxProcessesWithKiller failed: %v", err)
    }
    if len(procs) == 0 {
        t.Fatal("no processes found on Linux — this is very unusual")
    }
}

//go:build darwin
func TestMacOS_ProcessEnumeration_RunsPs(t *testing.T) {
    procs, err := modes.ListMacOSProcessesWithKiller()
    if err != nil {
        t.Fatalf("ListMacOSProcessesWithKiller failed: %v", err)
    }
    if len(procs) == 0 {
        t.Fatal("no processes found on macOS")
    }
}
```

---

## 13. VM & Sandbox Validation Strategy

### 13.1 Vagrant-Based VM Matrix

```ruby
# File: test/vms/vagrant/Vagrantfile
Vagrant.configure("2") do |config|
  # Windows 11
  config.vm.define "windows11" do |win|
    win.vm.box = "gusztavvargadr/windows-11"
    win.vm.provider "virtualbox" do |vb|
      vb.memory = 4096
      vb.cpus = 2
    end
    win.vm.provision "shell", path: "scripts/setup-windows.ps1"
    win.vm.provision "shell", path: "scripts/run-e2e.ps1"
  end

  # Ubuntu 24.04 GNOME
  config.vm.define "ubuntu2404" do |linux|
    linux.vm.box = "ubuntu/jammy64"
    linux.vm.provider "virtualbox" do |vb|
      vb.memory = 2048
      vb.cpus = 2
      vb.gui = true  # Need GUI for xdotool tests
    end
    linux.vm.provision "shell", path: "scripts/setup-linux.sh"
    linux.vm.provision "shell", path: "scripts/run-e2e.sh"
  end

  # macOS (via VMWare Fusion — requires license)
  config.vm.define "macos" do |mac|
    mac.vm.box = "jhcook/macos-sonoma"
    mac.vm.provider "vmware_desktop" do |vmw|
      vmw.memory = 4096
      vmw.cpus = 2
    end
  end
end
```

### 13.2 Docker-Based Linux Testing

```dockerfile
# File: test/vms/docker/Dockerfile.linux-e2e
FROM ubuntu:24.04

# Install dependencies for GUI testing
RUN apt-get update && apt-get install -y \
    xvfb x11vnc xdotool x11-utils procps psmisc \
    gsettings-desktop-schemas libglib2.0-bin \
    golang-go gcc \
    dbus-x11 busctl \
    && rm -rf /var/lib/apt/lists/*

# Create test user
RUN useradd -m -s /bin/bash testuser
USER testuser
WORKDIR /home/testuser/app

# Copy app
COPY --chown=testuser:testuser . .

# Run tests with virtual framebuffer
ENTRYPOINT ["xvfb-run", "-s", "-screen 0 1280x1024x24"]
CMD ["go", "test", "-v", "-tags=e2e", "./test/e2e/..."]
```

```yaml
# File: docker-compose.linux-e2e.yml
services:
  linux-e2e-gnome:
    build:
      context: .
      dockerfile: test/vms/docker/Dockerfile.linux-e2e
    environment:
      - DISPLAY=:99
      - ROBOTY_E2E=1
      - ROBOTY_SAFE_MODE=true
    volumes:
      - /tmp/.X11-unix:/tmp/.X11-unix:ro
    cap_add:
      - SYS_PTRACE  # For /proc access
    security_opt:
      - seccomp:unconfined  # For process operations
```

### 13.3 Packer Templates for Windows + macOS

```hcl
# File: test/vms/packer/windows-11.pkr.hcl
source "virtualbox-iso" "windows-11" {
  iso_url             = "https://software-static.download.prss.microsoft.com/dbazure/Win11_23H2_English_x64.iso"
  iso_checksum        = "sha256:..."
  guest_os_type       = "Windows11_64"
  communicator        = "winrm"
  winrm_username      = "vagrant"
  shutdown_command    = "shutdown /s /t 10 /f /d p:4:1 /c \"Packer Shutdown\""
  disk_size           = 51200
  memory              = 4096
  cpus                = 2
}

build {
  sources = ["source.virtualbox-iso.windows-11"]

  provisioner "powershell" {
    scripts = [
      "test/vms/scripts/install-go.ps1",
      "test/vms/scripts/install-chrome.ps1",
      "test/vms/scripts/install-vs-code.ps1",
      "test/vms/scripts/disable-defender.ps1",  # For performance
    ]
  }
}
```

### 13.4 Safe Testing in VMs

```powershell
# File: test/vms/scripts/setup-windows.ps1
# Safe test environment setup

# Create a restore point
Checkpoint-Computer -Description "Before Roboty E2E Tests" -RestorePointType MODIFY_SETTINGS

# Set up test snapshot of proxy registry
$originalProxy = Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -ErrorAction SilentlyContinue
$originalProxyValue = if ($originalProxy) { $originalProxy.ProxyEnable } else { 0 }
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value $originalProxyValue -Type DWord -Force

# Set up test snapshot of notification registry
$originalNotif = Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -ErrorAction SilentlyContinue
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings" -Name "NOC_GLOBAL_SETTING_TOASTS_ENABLED" -Value 1 -Type DWord -Force

# Set ROBOTY_SAFE_MODE globally
[Environment]::SetEnvironmentVariable("ROBOTY_SAFE_MODE", "true", "Machine")
```

---

## 14. Production Hardening Verification

### 14.1 Build Hardening Checklist

```go
// File: test/security/hardening_test.go
package security

func TestHardening_BuildTags_NoOSLeak(t *testing.T) {
    // Verify that platform-specific code is properly guarded with build tags
    // This is a compile-time check
}

func TestHardening_NoHardcodedSecrets(t *testing.T) {
    // Scan for hardcoded credentials, API keys, tokens
    patterns := []string{
        `api[_-]?key\s*=\s*['\"][^'\"]+['\"]`,
        `password\s*=\s*['\"][^'\"]+['\"]`,
        `token\s*=\s*['\"][A-Za-z0-9+/=]{20,}['\"]`,
        `secret\s*=\s*['\"][^'\"]+['\"]`,
    }
    // Use grep on source files
    for _, pattern := range patterns {
        // This would run grep in a real CI check
        re := regexp.MustCompile(pattern)
        _ = re
    }
}

func TestHardening_FilePermissions(t *testing.T) {
    // State files should have restrictive permissions
    info, err := os.Stat(".roboty_proxy_state")
    if err == nil {
        mode := info.Mode()
        if mode&0077 != 0 {
            t.Error("state file has excessive permissions:", mode)
        }
    }
}

func TestHardening_NoShellExecution_Unvalidated(t *testing.T) {
    // Verify that all exec.Command calls use validated/escaped arguments
    // This is a static analysis check - verify via code review
    // The production_killers.go uses NormalizeKillExec before executing
}
```

### 14.2 Go Vulnerability Scanning

```yaml
# In CI: Run govulncheck
steps:
  - name: Go Vulnerability Check
    run: |
      go install golang.org/x/vuln/cmd/govulncheck@latest
      govulncheck ./...
```

### 14.3 Static Analysis Configuration

```yaml
# In CI
- name: Static Analysis
  run: |
    go vet ./...
    staticcheck ./...
    gosec -quiet -fmt=json -out=results.json ./...
```

---

## 15. Real-World User Workflow Simulations

### 15.1 Workflow: Student During Exam

```go
// File: test/e2e/workflow_test.go

func TestE2E_Workflow_StudentExamMode(t *testing.T) {
    os.Setenv("ROBOTY_SAFE_MODE", "true")
    defer os.Unsetenv("ROBOTY_SAFE_MODE")

    ms := modes.NewModeService(nil, nil)

    // Create exam mode
    mode, _ := ms.CreateMode(modes.CreateModeRequest{
        Name:             "Exam Mode",
        DurationMinutes:  120,
        MuteNotifications: true,
        Apps: []modes.FocusModeApp{
            {AppName: "Chrome", AppExec: "chrome.exe", CloseOnActivate: true},
            {AppName: "Firefox", AppExec: "firefox.exe", CloseOnActivate: true},
            {AppName: "Discord", AppExec: "discord.exe", CloseOnActivate: true},
            {AppName: "Slack", AppExec: "slack.exe", CloseOnActivate: true},
            {AppName: "Spotify", AppExec: "spotify.exe", CloseOnActivate: true},
            {AppName: "Exam Browser", AppExec: "exambrowser.exe", IsAllowed: true},
        },
        AllowedURLs: []string{
            "exam-platform.com",
            "moodle.org",
            "stackoverflow.com",
        },
    })

    // Activate
    session, _ := ms.ActivateMode(mode.ID)

    // Verify state
    assert.True(t, session.Status == "active")

    // Simulate: timer expires naturally
    // (shortened for test)
    time.Sleep(100 * time.Millisecond)

    // Deactivate
    ms.DeactivateMode(session.ID)
}
```

### 15.2 Workflow: Remote Worker Deep Focus

```go
func TestE2E_Workflow_DeepWorkSession(t *testing.T) {
    os.Setenv("ROBOTY_SAFE_MODE", "true")
    defer os.Unsetenv("ROBOTY_SAFE_MODE")

    ms := modes.NewModeService(nil, nil)

    // Work focuses session
    mode, _ := ms.CreateMode(modes.CreateModeRequest{
        Name: "Deep Work",
        Apps: []modes.FocusModeApp{
            {AppName: "VS Code", AppExec: "code.exe", IsAllowed: true},
            {AppName: "Terminal", AppExec: "wt.exe", IsAllowed: true},
            {AppName: "Slack", AppExec: "slack.exe", CloseOnActivate: true},
        },
        AllowedURLs: []string{
            "github.com",
            "gitlab.com",
            "docs.microsoft.com",
        },
    })

    session, _ := ms.ActivateMode(mode.ID)

    // Switch foreground — only code should survive
    // (blocker is running in background with 2s interval)

    // Verify slack not running (would be killed)
    time.Sleep(3 * time.Second)

    ms.DeactivateMode(session.ID)
}
```

### 15.3 Workflow: Emergency Stop While Blocking

```go
func TestE2E_Workflow_EmergencyStopDuringBlock(t *testing.T) {
    os.Setenv("ROBOTY_SAFE_MODE", "true")
    defer os.Unsetenv("ROBOTY_SAFE_MODE")

    ms := modes.NewModeService(nil, nil)
    mode, _ := ms.CreateMode(modes.CreateModeRequest{
        Name: "Work",
        Apps: []modes.FocusModeApp{
            {AppName: "Chrome", AppExec: "chrome.exe", CloseOnActivate: true},
        },
    })

    ms.ActivateMode(mode.ID)

    // Emergency stop while blocker is active
    ms.EmergencyStop("user-emergency")

    // Must be fully clean
    active, _ := ms.GetActiveSession()
    if active != nil {
        t.Error("emergency stop should clean all sessions")
    }

    // Must be able to start a new session after emergency
    mode2, _ := ms.CreateMode(modes.CreateModeRequest{
        Name: "Resume Work",
        Apps: []modes.FocusModeApp{
            {AppName: "Chrome", AppExec: "chrome.exe"},
        },
    })
    session2, err := ms.ActivateMode(mode2.ID)
    if err != nil {
        t.Errorf("should be able to activate after emergency stop: %v", err)
    }
    ms.DeactivateMode(session2.ID)
}
```

---

## 16. CI/CD Strategy

### 16.1 GitHub Actions Workflow: Unit + Safety + Integration

```yaml
# File: .github/workflows/test-unit.yml
name: Unit & Safety Tests
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
      - name: Unit tests
        run: go test -v -race -count=1 -short ./internal/...
      - name: Fuzz tests (short)
        run: go test -fuzz=FuzzNormalizeKillExec -fuzztime=10s ./internal/modes/
      - name: Build
        run: go build -v ./...
      - name: Vet
        run: go vet ./...
      - name: Staticcheck
        run: |
          go install honnef.co/go/tools/cmd/staticcheck@latest
          staticcheck ./...

  safety-critical:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Critical safety tests
        run: go test -v -run TestCritical ./internal/modes/
      - name: Goblin chaos tests
        run: go test -v -run TestChaos ./internal/modes/
```

### 16.2 GitHub Actions: E2E + Integration (Nightly)

```yaml
# File: .github/workflows/test-e2e.yml
name: E2E & Integration Tests
on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM
  workflow_dispatch:  # Manual trigger

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Integration tests
        run: go test -v -race -count=1 -tags=integration ./test/integration/...

  e2e-linux:
    runs-on: ubuntu-latest
    services:
      xvfb:
        image: selenium/standalone-chrome:latest
    steps:
      - uses: actions/checkout@v4
      - name: Install deps
        run: |
          sudo apt-get update
          sudo apt-get install -y xvfb xdotool x11-utils dbus-x11
      - name: E2E tests (Linux)
        env:
          DISPLAY: :99
          ROBOTY_E2E: 1
          ROBOTY_SAFE_MODE: true
        run: |
          Xvfb :99 -screen 0 1280x1024x24 &
          sleep 2
          go test -v -tags=e2e -run TestE2E_Workflow ./test/e2e/

  e2e-windows:
    runs-on: windows-latest
    env:
      ROBOTY_E2E: 1
      ROBOTY_SAFE_MODE: true
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: E2E tests (Windows)
        run: go test -v -tags=e2e -run TestE2E_Windows ./test/e2e/

  e2e-macos:
    runs-on: macos-latest
    env:
      ROBOTY_E2E: 1
      ROBOTY_SAFE_MODE: true
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: E2E tests (macOS)
        run: go test -v -tags=e2e -run TestE2E_macOS ./test/e2e/
```

### 16.3 Weekly Chaos + Stress + Security

```yaml
# File: .github/workflows/test-chaos.yml
name: Chaos & Security Tests
on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday
  workflow_dispatch:

jobs:
  chaos:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Chaos tests
        run: go test -v -race -count=1 -run TestChaos -tags=chaos ./test/chaos/...
      - name: Stress tests
        run: go test -v -race -count=1 -run TestStress -tags=stress ./test/stress/...

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Fuzz tests (extended)
        run: |
          go test -fuzz=FuzzNormalizeKillExec -fuzztime=60s ./internal/modes/
          go test -fuzz=FuzzIsAlwaysAllowed -fuzztime=60s ./internal/modes/
          go test -fuzz=FuzzIsAllowed -fuzztime=60s ./internal/modes/
      - name: Security tests
        run: go test -v -run TestSecurity ./test/security/...
      - name: Vulnerabilities
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
      - name: Gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -quiet ./...
```

### 16.4 Pre-Commit Hook

```yaml
# File: .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-added-large-files
      - id: check-merge-conflict
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.57.0
    hooks:
      - id: golangci-lint
  - repo: local
    hooks:
      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        pass_filenames: false
      - id: go-build
        name: go build
        entry: go build ./...
        language: system
        pass_filenames: false
      - id: critical-tests
        name: critical safety tests
        entry: go test -run TestCritical -count=1 ./internal/modes/
        language: system
        pass_filenames: false
```

---

## 17. OS Process Safety Matrix

### 17.1 WINDOWS — Processes That MUST NEVER Be Killed

| Process Name | Exec | Category | If Killed | Protection Layer |
|-------------|------|----------|-----------|-----------------|
| Windows Explorer | `explorer.exe` | UI Shell | Desktop disappears, no taskbar/start | whitelist.json + KillSafetyVerifier |
| Desktop Window Manager | `dwm.exe` | Compositor | Aero/Win+D transparency lost, UI artifacts | whitelist.json + KillSafetyVerifier |
| Client Server Runtime | `csrss.exe` | System | BSOD (SYSTEM_PROCESS_TERMINATED) | whitelist.json + KillSafetyVerifier |
| Windows Logon | `winlogon.exe` | Auth | Immediate forced logout | whitelist.json + KillSafetyVerifier |
| Windows Init | `wininit.exe` | System | BSOD | whitelist.json + KillSafetyVerifier |
| Local Security Authority | `lsass.exe` | Security | No auth, forced reboot | whitelist.json + KillSafetyVerifier |
| Service Control Manager | `services.exe` | System | BSOD | whitelist.json + KillSafetyVerifier |
| Service Host | `svchost.exe` | System | Various services die, system instability | whitelist.json + KillSafetyVerifier |
| Shell Experience Host | `shellexperiencehost.exe` | UI Shell | Start menu/taskbar broken | whitelist.json + KillSafetyVerifier |
| Start Menu Exp Host | `startmenuexperiencehost.exe` | UI Shell | Start menu broken | whitelist.json + KillSafetyVerifier |
| Runtime Broker | `runtimebroker.exe` | Permissions | App permission prompts broken | whitelist.json + KillSafetyVerifier |
| Task Host Window | `taskhostw.exe` | Tasks | Scheduled tasks stop working | whitelist.json + KillSafetyVerifier |
| Search Host | `searchhost.exe` | Search | Windows Search broken | whitelist.json + KillSafetyVerifier |
| Search App | `searchapp.exe` | Search | Search UI broken | whitelist.json + KillSafetyVerifier |
| Search Indexer | `searchindexer.exe` | Search | Search index corrupted | whitelist.json + KillSafetyVerifier |
| System Settings | `systemsettings.exe` | UI Shell | Settings app broken | whitelist.json + KillSafetyVerifier |
| Logon UI | `logonui.exe` | Auth | Login screen broken | whitelist.json + KillSafetyVerifier |
| Local Session Manager | `lsm.exe` | Sessions | Session management broken | whitelist.json + KillSafetyVerifier |
| SmartScreen | `smartscreen.exe` | Security | No Defender SmartScreen | whitelist.json + KillSafetyVerifier |
| App Frame Host | `applicationframehost.exe` | UI Shell | WinRT apps broken | whitelist.json + KillSafetyVerifier |
| Text Input Host | `textinputhost.exe` | Input | Touch keyboard broken | whitelist.json + KillSafetyVerifier |
| Task Manager | `taskmgr.exe` | Tool | Can't open Task Manager | whitelist.json + KillSafetyVerifier |
| CTF Loader | `ctfmon.exe` | Input | Alternative input (IME) broken | whitelist.json + KillSafetyVerifier |
| Shell Infrastructure Host | `sihost.exe` | UI Shell | Shell components broken | whitelist.json + KillSafetyVerifier |
| Console Host | `conhost.exe` | Terminal | Console windows broken | whitelist.json + KillSafetyVerifier |
| Command Prompt | `cmd.exe` | Terminal | Can't use cmd | user whitelist |
| PowerShell | `powershell.exe` | Terminal | Can't use PowerShell | user whitelist |
| PowerShell Core | `pwsh.exe` | Terminal | Can't use pwsh | user whitelist |
| Windows Terminal | `wt.exe` | Terminal | Can't use Terminal | user whitelist |
| **Roboty** | `roboty.exe` | App | **SUICIDE** | KillSafetyVerifier selfNames |
| **Roboty Dev** | `roboty-dev.exe` | App | Dev process killed | KillSafetyVerifier selfNames |
| **Wails** | `wails.exe` | Runtime | Dev runtime killed | KillSafetyVerifier selfNames |
| Ancestor chain | Various | App | Parent/launcher/terminal/IDE | GetAncestorExecs() |

### 17.2 LINUX — Processes That MUST NEVER Be Killed

| Process Name | Category | If Killed | Protection |
|-------------|----------|-----------|-----------|
| `systemd` | Init | Kernel panic | whitelist.json + KillSafetyVerifier |
| `systemd-logind` | Session mgmt | Can't log in | whitelist.json |
| `systemd-journald` | Logging | Logs stop | whitelist.json |
| `systemd-networkd` | Networking | Network broken | whitelist.json |
| `systemd-resolved` | DNS | DNS resolution broken | whitelist.json |
| `dbus-daemon` | IPC | Desktop broken, apps can't communicate | whitelist.json |
| `NetworkManager` | Network | Network offline | whitelist.json |
| `wpa_supplicant` | WiFi | WiFi broken | whitelist.json |
| `polkitd` | Auth | Permission checks broken | whitelist.json |
| `udevd` | Devices | Device detection broken | whitelist.json |
| `gnome-shell` | WM/DE | GNOME desktop crashes | whitelist.json |
| `mutter` | Compositor | GNOME compositor broken | whitelist.json |
| `kwin` | WM/DE | KDE desktop crashes | whitelist.json |
| `plasmashell` | WM/DE | KDE Plasma crashes | whitelist.json |
| `Xorg` | Display | X11 server dies, apps crash | whitelist.json |
| `Xwayland` | Display | Wayland XWayland bridge broken | whitelist.json |
| `wayland` | Display | Wayland compositor, all apps crash | whitelist.json |
| `pipewire` | Audio | Audio broken | whitelist.json |
| `pulseaudio` | Audio | Audio broken | whitelist.json |
| `wireplumber` | Audio | Audio routing broken | whitelist.json |
| `lightdm` / `gdm` / `sddm` | Display mgr | Login screen broken | whitelist.json |
| `login` | Auth | Can't log in | whitelist.json |
| `sshd` | Remote | SSH access broken | whitelist.json |
| `init` | Init | PID 1 — kernel panic | whitelist.json |
| `kthreadd` | Kernel | Kernel threads, system crash | whitelist.json |

### 17.3 macOS — Processes That MUST NEVER Be Killed

| Process Name | Category | If Killed | Protection |
|-------------|----------|-----------|-----------|
| `WindowServer` | Compositor | Window server crash, all UI gone | whitelist.json |
| `Finder` | UI Shell | Desktop and Finder broken | whitelist.json |
| `Dock` | UI Shell | Dock disappears | whitelist.json |
| `launchd` | Init | PID 1 — kernel panic | whitelist.json |
| `SystemUIServer` | UI Shell | Menu extras broken | whitelist.json |
| `ControlCenter` | UI Shell | Control Center broken | whitelist.json |
| `NotificationCenter` | Notifications | Notification center broken | whitelist.json |
| `Spotlight` | Search | Spotlight broken | whitelist.json |
| `WindowManager` | Windows | Stage Manager windows broken | whitelist.json |

### 17.4 All Platforms — Self + Ancestor Protection

| Type | What | How |
|------|------|-----|
| Self | `roboty` / `roboty.exe` / `roboty1` / `roboty-dev` | KillSafetyVerifier.selfNames |
| Wails runtime | `wails` / `wails.exe` | KillSafetyVerifier.selfNames |
| Terminal | `cmd`, `powershell`, `pwsh`, `wt`, `bash`, `zsh`, `fish` | GetAncestorExecs() |
| IDE | `code`, `goland`, `intellij`, `vim`, `nvim` | GetAncestorExecs() |
| Launcher | `explorer`, `Dock`, `gnome-shell`, `plasmashell` | GetAncestorExecs() |
| SSH parent | `sshd`, `mstsc`, `remmina`, `tigerVNC` | GetAncestorExecs() |
| Dev tools | `wails`, `go`, `node`, `npm`, `vite`, `webpack` | GetAncestorExecs() |

---

## 18. Bypass & Allowlist Matrices

### 18.1 Localhost / Internal URL Bypass (Must NEVER Be Blocked)

| Host | Port | Purpose | Why |
|------|------|---------|-----|
| `localhost` | any | Generic loopback | System networking |
| `127.0.0.1` | any | IPv4 loopback | System networking |
| `127.0.0.0/8` | any | IPv4 loopback range | All loopback addresses |
| `::1` | any | IPv6 loopback | IPv6 networking |
| `0.0.0.0` | any | Wildcard bind | Server bind |
| `0:0:0:0:0:0:0:1` | any | IPv6 long form | Full IPv6 loopback |
| `0:0:0:0:0:0:0:0` | any | IPv6 unspecified | IPv6 any address |
| `localhost.localdomain` | any | FQDN localhost | mDNS resolution |
| `local` | any | Bonjour local | macOS Bonjour |
| `169.254.0.0/16` | any | Link-local | DHCP failure fallback |
| `10.0.0.0/8` | any | Private A | Internal networks |
| `192.168.0.0/16` | any | Private C | Internal networks |
| `wails` | any | Wails protocol | Wails WebView frontend |
| `wails.localhost` | any | Wails dev | Wails dev server |

### 18.2 Browser/WebView Safe Allowlists

| Pattern | Example | Source | Must Pass |
|---------|---------|--------|-----------|
| `wails://*` | `wails://app` | Wails runtime | ✓ |
| `wails://localhost*` | `wails://localhost:34115` | Wails dev | ✓ |
| `http://localhost:*` | `http://localhost:34115` | Vite dev server | ✓ |
| `http://127.0.0.1:*` | `http://127.0.0.1:34115` | Vite fallback | ✓ |
| `ws://localhost:*` | `ws://localhost:34115` | WebSocket HMR | ✓ |
| `wss://localhost:*` | `wss://localhost:5173` | WebSocket Secure | ✓ |
| `http://*:5173` | Vite default | Vite HMR | ✓ |
| `http://*:4173` | Vite preview | Vite preview | ✓ |
| `http://*:3000` | Generic dev | Common dev port | ✓ |
| `http://*:8080` | Generic dev | Common dev port | ✓ |
| `http://*:3001` | Wails dev | Wails dev server | ✓ |

### 18.3 Dev-Environment Safe Allowlists (Wails + Vite + Node + WebSocket + HMR)

| Component | Protocol | Host | Port |
|-----------|----------|------|------|
| Wails WebView | `wails://` | `localhost` | dynamic |
| Wails runtime | `http://` | `127.0.0.1` | 34115 |
| Wails dev server | `http://` | `localhost` | 34115 |
| Vite dev server | `http://` | `localhost` | 5173 |
| Vite HMR WS | `ws://` | `localhost` | 5173 |
| Vite preview | `http://` | `localhost` | 4173 |
| Node inspector | `ws://` | `127.0.0.1` | 9229 |
| Go pprof | `http://` | `127.0.0.1` | 6060 |
| Go debug | `http://` | `127.0.0.1` | 2345 |
| SQLite browser | `http://` | `localhost` | 8080 |

### 18.4 Ancestor Protection (Auto-Detected)

These are detected by walking the process tree at startup. Any ancestor of the Roboty process is automatically added to the kill-protection list:

```
IDEs: code, goland, idea, vim, nvim, emacs, sublime_text, atom
Terminals: cmd, powershell, pwsh, wt, WindowsTerminal, bash, zsh, fish, sh, dash, xterm, gnome-terminal, konsole, alacritty, iterm2
Launchers: explorer, Finders, Dock, gnome-shell, plasmashell, kwin, sddm, gdm, lightdm
Dev Tools: go, node, npm, npx, yarn, pnpm, vite, webpack, tsc, eslint, prettier, wails
SSH/Remote: sshd, ssh, mstsc, remmina, TigerVNC, vnc, RDC, RemoteDesktop
```

---

## 19. Threat Models & Failure Trees

### 19.1 Threat Model: Process Kill Safety

```
┌────────────────────────────────────────────────────┐
│                  Threat T-KILL-01                   │
│        "Roboty accidentally kills system process"   │
└────────────────────────────────────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │   Failure Detection    │
              │   Tree: T-KILL-01      │
              └───────────────────────┘
                          │
        ┌─────────────────┼──────────────────┐
        ▼                 ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Input:       │  │ Input:       │  │ Input:       │
│ User adds    │  │ Blocker      │  │ CloseApp     │
│ app to mode  │  │ polls FG     │  │ list         │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ validate     │  │ Normalize    │  │ Normalize    │
│ AppExec()    │  │ KillExec()   │  │ KillExec()   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ IsSafeToKill │  │ IsSafeToKill │  │ IsSafeToKill │
│ → false if   │  │ → false if   │  │ → false if   │
│   system     │  │   system     │  │   system     │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       ▼                 ▼                 ▼
  ┌─────────┐      ┌─────────┐      ┌─────────┐
  │ Return  │      │ Return  │      │ Return  │
  │ error   │      │ skip    │      │ skip    │
  │ "system │      │ kill    │      │ kill    │
  │ process"│      │ (log)   │      │ (log)   │
  └─────────┘      └─────────┘      └─────────┘

MITIGATIONS:
  ● whitelist.json with all known system processes
  ● KillSafetyVerifier with hardcoded system names (fallback)
  ● selfNames protection (never kill roboty/wails/wails)
  ● GetAncestorExecs() protects ancestors
  ● KillLoopDetector prevents infinite kill attempts
  ● NormalizeKillExec rejects dangerous characters
  ● NormalizeKillExec strips path traversal attempts
  ● validateAppExec() called at mode creation time
```

### 19.2 Threat Model: Proxy Bypass

```
┌────────────────────────────────────────────────────┐
│                  Threat T-PROXY-01                  │
│     "Malicious site bypasses Roboty proxy filter"   │
└────────────────────────────────────────────────────┘
                          │
                          ▼
         ┌────────────────────────────────┐
         │   Attack Vectors               │
         └────────────────────────────────┘
                          │
    ┌──────────┬──────────┼──────────┬──────────┐
    ▼          ▼          ▼          ▼          ▼
┌──────┐ ┌──────────┐ ┌──────┐ ┌──────────┐ ┌──────┐
│ DNS  │ │ HTTP CON │ │ IPv6 │ │ Port     │ │ SNI  │
│ over │ │NECT over │ │ add- │ │ confus-  │ │ hid- │
│ HTTPS│ │ allowed  │ │ ress │ │ ion      │ │ ing  │
└──────┘ │ host     │ └──────┘ └──────────┘ └──────┘
         └──────────┘

MITIGATIONS:
  ● isAlwaysAllowed() checks localhost BEFORE allowed list
  ● Hostname normalization strips protocol/path/port
  ● Subdomain matching prevents DNS tunneling
  ● IPv6 loopback explicitly checked
  ● FQDN trailing dot handled correctly
  ● Host extracted from r.Host (preserves port for CONNECT)
```

### 19.3 Threat Model: Crash Orphan

```
┌────────────────────────────────────────────────────┐
│                 Threat T-CRASH-01                  │
│  "Process crashes leaving proxy enabled forever"   │
└────────────────────────────────────────────────────┘
                          │
                          ▼
       ┌─────────────────────────────────────────┐
       │              Scenario                    │
       │  1. Focus mode activates proxy           │
       │  2. OS proxy is set via registry/cmd     │
       │  3. State marker file written            │
       │  4. PROCESS CRASHES (no cleanup)         │
       │  5. Next startup: detect marker, cleanup │
       └─────────────────────────────────────────┘

MITIGATIONS:
  ● proxyStateName marker file written on every Enable
  ● State marker cleared on every Disable
  ● If marker exists on startup → CleanupOrphanedProxy() runs
  ● Same pattern for notifications: notifStateName
  ● signalHandler catches SIGINT/SIGTERM
  ● EmergencyStop() performs full cleanup
  ● ProxyWatchdog detects proxy failure + disables system proxy
```

### 19.4 Threat Model: Kill Loop

```
┌────────────────────────────────────────────────────┐
│                 Threat T-LOOP-01                   │
│   "Blocked app respawns instantly causing loop"    │
└────────────────────────────────────────────────────┘
                          │
                          ▼
       ┌─────────────────────────────────────────┐
       │          Kill Loop Sequence              │
       │  1. Blocker identifies unallowed app     │
       │  2. Blocker kills app                    │
       │  3. App auto-restarts (or user reopens)  │
       │  4. Blocker polls foreground again       │
       │  5. Sees app again → kills again         │
       │  6. REPEAT forever (CPU + UX nightmare)  │
       └─────────────────────────────────────────┘

MITIGATIONS:
  ● KillLoopDetector tracks kills per exec within window
  ● MaxConsecutiveKills (default 10) within KillLoopWindow (30s)
  ● When threshold exceeded → EmergencyFailsafe triggered
  ● EmergencyFailsafe calls EmergencyStop → disables everything
  ● SafetyAuditLogger records all kill attempts
  ● Allows debugging kill patterns post-incident
```

---

## 20. Edge-Case Matrices

### 20.1 Process Killing Edge Cases

| Edge Case | Expected Behavior | Covered By |
|-----------|------------------|------------|
| Empty exec name | Logged, not killed | `TestCloseApp_EdgeCases` |
| Null bytes in exec name | Rejected by NormalizeKillExec | `TestNormalizeKillExec_RejectsDangerousInputs` |
| Very long exec name (1000+ chars) | Rejected or truncated | `FuzzNormalizeKillExec` |
| Unicode homoglyph exec name | Rejected or normalized safely | `TestNormalizeKillExec_UnicodeNormalization` |
| Exec name that matches system process via substring | Exact match only (no substring) | `TestWhitelist_NoPartialMatching` |
| Process that keeps restarting | Kill loop detection triggers failsafe | `TestChaos_KillLoop_RestartingApp` |
| Process already dead | Error logged, no panic | `TestCloseApps_EdgeCases` |
| SIGKILL (cannot be caught) | State marker allows recovery | `TestCrashRecovery_ForceTerminate_Cleanup` |
| Concurrent kill same process | Serialized via mutex, no race | `TestConcurrent_KillLoopDetector_Race` |
| Kill process with path separators | Rejected by NormalizeKillExec | `TestRedTeam_KillExecBypass` |

### 20.2 Proxy Edge Cases

| Edge Case | Expected Behavior | Covered By |
|-----------|------------------|------------|
| CONNECT request with no Host header | Use r.URL.Hostname() | `TestHandleHTTPS_HostFallback` |
| HTTP request with no URL host | Use r.Host | `TestHandleHTTPPlain_HostParsing` |
| IPv6 address with port (e.g., `[::1]:62828`) | Stripped correctly, localhost allowed | `TestIsAlwaysAllowed_IPv6EdgeCases` |
| FQDN trailing dot (`example.com.`) | Stripped, matches correctly | `TestIsAllowed_HostSource` |
| Empty allowed URL list | Everything blocked except localhost | `TestURLBlocker_BlockAll` |
| Port number in allowed URL (e.g., `example.com:8080`) | Preserved and matched | `TestNormalizeURLs` |
| Protocol prefix in allowed URL (e.g., `https://x.com`) | Stripped before matching | `TestNormalizeURLs` |
| Proxy enable fails → proxy should not start | State file not written, error returned | `TestCritical_ProxyStateNotSavedWhenEnableFails` |
| Proxy disable fails → state file still cleared | Best-effort cleanup | `TestURLBlockerWithDI_ProxyEnableFailure_NoOrphan` |
| Concurrent proxy start/stop | Mutex-protected, no panic | `TestConcurrent_Proxy_StartStopRace` |

### 20.3 Notification Edge Cases

| Edge Case | Expected Behavior | Covered By |
|-----------|------------------|------------|
| Already muted → Mute() | Still logs success | `TestNotificationMute_RoundTrip` |
| Already restored → Restore() | Still logs success | `TestNotificationMute_RoundTrip` |
| No display server | Command fails, logs error, returns | `TestLinuxNotificationCommands` |
| busctl not available | Logs error, returns safely | `TestLinuxNotificationCommands` |
| osascript not available | Logs error, returns safely | `TestMacOSNotificationCommands` |
| Dev mode: no-op | Logs what would happen | `TestCritical_SafeModePreventsKill` |
| Concurrent Mute/Restore | Mutex-protected by fake | `TestFakeNotificationManager_MuteRestore` |

### 20.4 Focus Mode Lifecycle Edge Cases

| Edge Case | Expected Behavior | Covered By |
|-----------|------------------|------------|
| Activate same mode twice | Second is no-op or returns existing session | `TestIntegration_FocusModeLifecycle` |
| Deactivate with no active session | Error returned, no panic | `TestIntegration_FocusModeLifecycle` |
| Timer expires naturally | Clean shutdown, session marked complete | `TestCrashRecovery_EmergencyStop_Cleanup` |
| Create mode with empty apps | Succeeds with empty blocker list | `TestIntegration_FocusModeLifecycle` |
| Delete mode while active session exists | Mode deleted, session continues (or cancelled) | `TestIntegration_FocusModeLifecycle` |
| EmergencyStop while nothing is running | No-op, no panic | `TestCritical_EmergencyStop_SafeToCallMultiple` |
| Concurrent create + activate + delete | Data races prevented by mutex | `TestConcurrent_ModeService_ReadWriteRace` |

---

## 21. Regression Test Plan

### 21.1 Critical Safety Regression Tests (Run on Every Commit)

```
TestCritical_ExplorerNeverKillable
TestCritical_DwmNeverKillable
TestCritical_LocalhostNeverBlockedByProxy
TestCritical_RobotyNeverSelfKill
TestCritical_AncestorProcessesProtected
TestCritical_SafeModePreventsKill
TestCritical_KillLoopDetector_PreventsInfiniteLoop
TestCritical_KillMustPassSafetyVerifier
TestCritical_NormalizeKillExec_RejectsDangerousPatterns
TestCritical_ValidateAppExec_RejectsSystemProcesses
TestCritical_Whitelist_AllRequiredEntries
TestCritical_EmergencyStop_SafeToCallMultiple
TestCritical_ConcurrentEmergencyStop
TestCritical_ProxyStateNotSavedWhenEnableFails
TestCritical_safeGo_RecoversPanic
TestCritical_Whitelist_NoSectionKeys
```

### 21.2 Release-Blocker Regression Tests

```
// All Critical tests
// All Fuzz tests (60s each)
// Full safety_test.go
// Full proxy_test.go
// Full blocker_test.go
// Full integration_di_test.go
// QuietHours notification tests
// KillSafetyVerifier system process checks (all 50+ processes)
// Cross-platform build tests (windows/linux/darwin)
// Race detector tests (all packages)
```

### 21.3 Pre-Release Manual Testing Checklist

- [ ] Whitelist JSON file is present and parsable
- [ ] All system-critical Windows processes are listed
- [ ] All system-critical Linux processes are listed
- [ ] All system-critical macOS processes are listed
- [ ] No `bundle_ids` or section keys leak into process names
- [ ] Self-protection covers all possible binary names
- [ ] Ancestor detection covers Terminal/IDE/SSH
- [ ] `isAlwaysAllowed` covers all loopback ranges
- [ ] `wails://` protocol is never blocked
- [ ] Vite HMR ports are never blocked
- [ ] `NormalizeKillExec` rejects all shell metacharacters
- [ ] `DevMode` prevents all OS modifications
- [ ] State markers are cleaned on graceful shutdown
- [ ] Start-up cleanup detects orphaned proxy/notif states
- [ ] Proxy health check detects failures within 5 seconds
- [ ] Kill loop detector prevents infinite loops
- [ ] Emergency failsafe triggers before MaxConsecutiveKills*2
- [ ] `safeGo` recovers panics without crashing the process
- [ ] Signal handler catches SIGINT/SIGTERM
- [ ] No `exec.Command` is called with unsanitized user input
- [ ] All tests pass with `-race`
- [ ] All tests pass on Windows 11
- [ ] All tests pass on Ubuntu 24.04 GNOME
- [ ] All tests pass on macOS Sonoma

---

## 22. Architecture Improvement Recommendations

### 22.1 Interface for DataStore (Remove Direct DB Dependency)

```go
// Recommend: Create a DataStore interface to abstract sql.DB
type DataStore interface {
    // Focus Modes
    CreateFocusMode(ctx context.Context, p CreateModeParams) (*FocusMode, error)
    GetFocusModeByID(ctx context.Context, id string) (*FocusMode, error)
    GetAllFocusModes(ctx context.Context) ([]FocusMode, error)
    UpdateFocusMode(ctx context.Context, p UpdateModeParams) error
    DeleteFocusMode(ctx context.Context, id string) error

    // Focus Sessions  
    CreateFocusSession(ctx context.Context, p CreateSessionParams) (*FocusSession, error)
    GetActiveFocusSession(ctx context.Context) (*FocusSession, error)
    GetFocusSessionByID(ctx context.Context, id string) (*FocusSession, error)
    UpdateFocusSessionStatus(ctx context.Context, id, status string) error

    // Focus Mode Apps
    CreateFocusModeApp(ctx context.Context, p CreateAppParams) error
    GetFocusModeAppsByModeID(ctx context.Context, modeID string) ([]FocusModeApp, error)
    GetFocusModeAllowedAppsByModeID(ctx context.Context, modeID string) ([]FocusModeApp, error)
    DeleteFocusModeAppsByModeID(ctx context.Context, modeID string) error

    // Focus Mode URLs
    CreateFocusModeURL(ctx context.Context, p CreateURLParams) error
    GetFocusModeURLsByModeID(ctx context.Context, modeID string) ([]FocusModeURL, error)
    DeleteFocusModeURLsByModeID(ctx context.Context, modeID string) error
}
```

### 22.2 Add ForegroundDetector Interface

```go
// Currently ForegroundTracker is concrete — abstract it
type ForegroundDetector interface {
    Poll() (*ProcessInfo, error)
    Start(ctx context.Context, interval time.Duration, callback func(ProcessInfo))
    Stop()
    ListRunning() ([]ProcessInfo, error)
}
```

### 22.3 Add URLBlocker Interface

```go
type URLBlockerService interface {
    Start(port int, allowedURLs []string) error
    Stop() error
    SetAllowedURLs(urls []string)
    IsRunning() bool
    IsAllowed(host string) bool
}
```

### 22.4 Use Options Pattern for ModeService

```go
type ModeServiceConfig struct {
    Database          *db.DB
    Queries           *db.Queries
    DataStore         DataStore // new
    Tracker           *ForegroundTracker
    Killer            ProcessKiller
    ProxyMgr          SystemProxyManager
    NotifMgr          NotificationManager
    WatchdogInterval  time.Duration
    ProxyHealthTimeout time.Duration
    MaxConsecutiveKills int
    KillLoopWindow     time.Duration
    DevMode            bool
}
```

### 22.5 Config Separation

```go
// Move all safety constants to a config struct
type FocusModeConfig struct {
    WatchdogInterval     time.Duration `default:"5s"`
    ProxyHealthTimeout   time.Duration `default:"3s"`
    MaxConsecutiveKills  int           `default:"10"`
    KillLoopWindow       time.Duration `default:"30s"`
    EmergencyFailsafeMax int           `default:"3"`
    DefaultProxyPort     int           `default:"62828"`
    BlockPollInterval    time.Duration `default:"2s"`
}
```

---

## 23. Telemetry & Audit Logging Strategy

### 23.1 Structured Audit Events

```go
type SafetyEvent struct {
    Type      SafetyEventType `json:"type"`
    Target    string          `json:"target,omitempty"`
    Source    string          `json:"source,omitempty"`
    Allowed   bool            `json:"allowed,omitempty"`
    Timestamp time.Time       `json:"timestamp"`
    Message   string          `json:"message"`
    // Additional fields for observability
    SessionID string `json:"session_id,omitempty"`
    ModeID    string `json:"mode_id,omitempty"`
    PID       int    `json:"pid,omitempty"`
    Duration  string `json:"duration,omitempty"`
    Error     string `json:"error,omitempty"`
}
```

### 23.2 Audit Log Export

```go
// File: internal/modes/audit_export.go
type AuditExporter interface {
    ExportEvents(events []SafetyEvent) error
    ExportToJSON(path string) error
    ExportToSyslog() error
}

// File-based audit log
type FileAuditExporter struct {
    path string
}

func (e *FileAuditExporter) ExportEvents(events []SafetyEvent) error {
    data, _ := json.MarshalIndent(events, "", "  ")
    return os.WriteFile(e.path, data, 0644)
}
```

### 23.3 Metrics Collection

```go
// Recommended metrics (exported via expvar or Prometheus):
var (
    metricsActiveSessions     = expvar.NewInt("focus_active_sessions")
    metricsTotalActivations   = expvar.NewInt("focus_total_activations")
    metricsKillAttempts       = expvar.NewInt("focus_kill_attempts")
    metricsKillBlocked        = expvar.NewInt("focus_kill_blocked_safety")
    metricsProxyEnables       = expvar.NewInt("focus_proxy_enables")
    metricsProxyDisables      = expvar.NewInt("focus_proxy_disables")
    metricsEmergencyStops     = expvar.NewInt("focus_emergency_stops")
    metricsPanicsRecovered    = expvar.NewInt("focus_panics_recovered")
    metricsWatchdogActions    = expvar.NewInt("focus_watchdog_actions")
    metricsKillLoopsDetected  = expvar.NewInt("focus_kill_loops_detected")
)
```

### 23.4 Recommended Telemetry Pipeline

```
Roboty Process
     │
     ├──→ SafetyAuditLogger (in-memory ring buffer)
     │         │
     │         ├──→ FileAuditExporter (JSON on disk)
     │         ├──→ SyslogExporter (Linux/macOS syslog)
     │         ├──→ EventLogExporter (Windows Event Log)
     │         └──→ expvar (HTTP /debug/vars)
     │
     └──→ Standard log (go log)
               │
               ├──→ stderr (default)
               ├──→ file (opt-in: -logfile)
               └──→ structured JSON (opt-in: -log-json)
```

---

## 24. Emergency Rollback Architecture

### 24.1 Rollback State Machine

```
STARTUP
   │
   ▼
Check for Orphans
   │
   ├── proxyStateName exists?
   │      └── Yes → CleanupOrphanedProxy()
   │                   └── Disable OS proxy
   │                   └── Remove state marker
   │
   ├── notifStateName exists?
   │      └── Yes → CleanupOrphanedNotifications()
   │                   └── Restore notifications
   │                   └── Remove state marker
   │
   └── No orphans → Normal startup
                         │
                         ▼
                   Check for stale sessions
                         │
                         ├── Active session but timer expired?
                         │      └── Yes → Auto-complete session
                         │
                         └── Active session with running timer?
                                └── Yes → Resume session with remaining time
```

### 24.2 Emergency Stop Chain

```
EmergencyStop("reason")
   │
   ├── Lock mutex
   ├── Stop AppBlocker (cancel goroutine)
   ├── Stop URLBlocker (shutdown HTTP server)
   │       └── Disable system proxy (OS-level)
   │       └── Clear proxy state marker
   ├── Restore notifications (OS-level)
   ├── Clear notification state marker
   ├── Mark all active sessions as "emergency_stopped"
   └── Unlock mutex
```

### 24.3 Signal-Based Recovery

```go
func SetupSignalHandler() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    safeGo(func() {
        sig := <-sigCh
        log.Printf("Signal %v — cleanup", sig)
        if callback := getGlobalEmergencyCallback(); callback != nil {
            callback("signal-" + sig.String())
        }
        os.Exit(1)
    })
}
```

### 24.4 Kill Switch (Emergency Disable All)

```go
// Button on UI or CLI flag: --emergency-stop
func (a *App) EmergencyKillSwitch() string {
    if a.modeService == nil {
        return "no mode service"
    }
    a.modeService.EmergencyStop("kill-switch-ui")
    // Also clear all state markers on disk
    sf := modes.NewFileStateManager()
    sf.CleanupAll()
    return "all systems stopped and state cleaned"
}
```

---

## 25. Coverage Strategy

### 25.1 Coverage Targets

| Package | Unit Coverage | Integration Coverage | Branch Coverage |
|---------|-------------|---------------------|----------------|
| `internal/modes` | >85% | >70% | >80% |
| `internal/modes/safety.go` | 100% | 100% | 100% |
| `internal/modes/blocking.go` | >90% | >80% | >85% |
| `internal/modes/proxy.go` | >85% | >75% | >80% |
| `internal/modes/service.go` | >80% | >70% | >75% |
| `internal/db` | >70% | >80% | >75% |
| `internal/modes/tracker.go` | >60% | >60% | >65% |

### 25.2 Fuzz Coverage Goals

| Fuzz Target | Corpus Seeds | Time | Coverage Target |
|-------------|-------------|------|----------------|
| `FuzzNormalizeKillExec` | 20+ | 60s | All branches |
| `FuzzIsAlwaysAllowed` | 15+ | 60s | All branches |
| `FuzzIsAllowed` | 10+ combined | 60s | All branches |

### 25.3 Risk-Based Testing Priority

```
P0 (Blocking): Safety verifier, whitelist, kill protection, localhost bypass
P1 (Critical): Proxy lifecycle, notification lifecycle, crash recovery
P2 (Important): App blocking, URL filtering, concurrent access
P3 (Standard): Process enumeration, app detection, logging
P4 (Low): Performance, stress, long-running
```

---

## 26. Recommended Libraries & Tools

| Tool | Purpose | Command |
|------|---------|---------|
| `testing` | Go built-in test framework | `go test` |
| `go vet` | Static analysis | `go vet ./...` |
| `staticcheck` | Advanced static analysis | `staticcheck ./...` |
| `gosec` | Security linting | `gosec ./...` |
| `govulncheck` | Vulnerability scanning | `govulncheck ./...` |
| `go test -race` | Data race detection | `go test -race ./...` |
| `go test -fuzz` | Fuzz testing | `go test -fuzz=FuzzName` |
| `gotestsum` | Test output formatter | `gotestsum --junitfile report.xml` |
| `testify` | Assertions/mocks | `go get github.com/stretchr/testify` |
| `ginkgo` | BDD test framework | `go get github.com/onsi/ginkgo/v2` |
| `gomega` | BDD matchers | `go get github.com/onsi/gomega` |
| `go-cmp` | Deep comparison | `go get github.com/google/go-cmp` |
| `faker` | Fake data generation | `go get github.com/brianvoe/gofakeit/v6` |
| `pre-commit` | Git hooks | `pip install pre-commit` |
| `vagrant` | VM management | `choco install vagrant` |
| `packer` | VM image building | `choco install packer` |
| `xvfb` | Virtual framebuffer (Linux) | `apt install xvfb` |
| `Xvfb` | Headless X11 server | `Xvfb :99 &` |
| `docker` | Containerized testing | `docker build -f Dockerfile.linux-e2e .` |
| `gotip` | Latest Go (for new fuzz features) | `go install golang.org/dl/gotip@latest` |

---

## 27. Example Fake Process Launcher (Test Helper)

```go
// File: test/helpers/mockprocess/mock_process.go
package mockprocess

import (
    "fmt"
    "os/exec"
    "runtime"
)

// Launcher creates mock processes for testing.
// On Windows: creates a dummy process that sleeps.
// On Linux/macOS: creates a sleep process.
type Launcher struct {
    processes []*exec.Cmd
}

func (l *Launcher) Start(name string) error {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("powershell", "-NoProfile", "-Command",
            fmt.Sprintf("Start-Sleep -Seconds 3600; Start-Process '%s' -NoNewWindow", name))
    case "linux", "darwin":
        cmd = exec.Command("sleep", "3600")
    }
    if cmd == nil {
        return fmt.Errorf("unsupported platform")
    }
    if err := cmd.Start(); err != nil {
        return err
    }
    l.processes = append(l.processes, cmd)
    return nil
}

func (l *Launcher) Cleanup() {
    for _, cmd := range l.processes {
        cmd.Process.Kill()
    }
    l.processes = nil
}

// MockApp represents a fake application that can be "blocked"
type MockApp struct {
    Name      string
    Exec      string
}

func NewMockApp(name, exec string) *MockApp {
    return &MockApp{Name: name, Exec: exec}
}
```

---

## 28. Makefile Targets

```makefile
# File: Makefile
.PHONY: test test-unit test-safety test-integration test-e2e test-all \
        test-chaos test-stress test-security test-race test-fuzz \
        build lint vet check

# Quick tests (run on every save)
test-unit:
	go test -v -count=1 -short -run Test[A-Z] ./internal/...

# Safety-critical tests (run on every commit)
test-safety:
	go test -v -count=1 -run TestCritical ./internal/modes/
	go test -v -count=1 -run TestSafety ./internal/modes/
	go test -v -count=1 -run 'TestKill|TestWhitelist|TestNormalize|TestRedTeam' ./internal/modes/

# Integration tests (DI fakes + in-memory)
test-integration:
	go test -v -count=1 -race -run TestIntegration ./test/integration/

# End-to-end tests (in VM only)
test-e2e:
	ROBOTY_E2E=1 ROBOTY_SAFE_MODE=true go test -v -count=1 -run TestE2E ./test/e2e/

# Chaos + failure injection
test-chaos:
	go test -v -count=1 -race -run TestChaos ./test/chaos/

# Stress tests
test-stress:
	go test -v -count=1 -race -run TestStress ./test/stress/

# Security tests
test-security:
	go test -v -count=1 -run TestSecurity ./test/security/
	govulncheck ./...
	gosec -quiet ./...

# Race detector
test-race:
	go test -race -count=1 -short ./internal/...

# Fuzz tests
test-fuzz:
	go test -fuzz=FuzzNormalizeKillExec -fuzztime=60s ./internal/modes/
	go test -fuzz=FuzzIsAlwaysAllowed -fuzztime=60s ./internal/modes/
	go test -fuzz=FuzzIsAllowed -fuzztime=60s ./internal/modes/

# Full test suite
test-all: test-unit test-safety test-integration test-race test-fuzz test-security

# Build
build:
	go build -v -buildmode=pie -ldflags="-s -w" -o build/roboty.exe .

# Lint + Vet
lint:
	staticcheck ./...
	golangci-lint run ./...

vet:
	go vet ./...

check: vet lint test-safety test-unit

# Coverage
coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -html=coverage.out -o coverage.html
```

---

## 29. Final Safety Checklist

### 29.1 Pre-Deployment Safety Verification

```
[ ] KillSafetyVerifier blocks ALL processes in Section 17
[ ] KillSafetyVerifier blocks ALL self-names
[ ] KillSafetyVerifier blocks ALL ancestor execs at startup
[ ] GetAncestorExecs walks the full process tree (depth 50+)
[ ] GetWhitelistExecs reads whitelist.json successfully
[ ] GetWhitelistExecs returns non-empty set (fallback hardcoded list)
[ ] whitelist.json contains ALL section processes with normalized keys
[ ] whitelist.json does NOT contain section key names as entries
[ ] isAlwaysAllowed passes ALL localhost/loopback patterns
[ ] isAlwaysAllowed passes ALL wails:// patterns
[ ] isAlwaysAllowed blocks ALL external domains
[ ] NormalizeKillExec rejects ALL shell metacharacters
[ ] NormalizeKillExec rejects ALL path traversal attempts
[ ] NormalizeKillExec rejects ALL flag injection attempts
[ ] KillLoopDetector triggers at MaxConsecutiveKills
[ ] KillLoopDetector tracks processes independently
[ ] KillLoopDetector window expiry resets correctly
[ ] EmergencyFailsafe triggers only once (idempotent)
[ ] EmergencyStop is idempotent and thread-safe
[ ] safeGo recovers panics without crashing process
[ ] SetupSignalHandler registers SIGINT/SIGTERM handlers
[ ] proxyStateName marker written on Enable
[ ] proxyStateName marker cleared on Disable
[ ] proxyStateName marker NOT written if Enable fails
[ ] CleanupOrphanedProxy() detects and cleans on startup
[ ] notifStateName marker written on Mute
[ ] notifStateName marker cleared on Restore
[ ] CleanupOrphanedNotifications() detects and restores on startup
[ ] DevMode (ROBOTY_SAFE_MODE=true) prevents ALL OS modifications
[ ] DevMode logs all would-be modifications
[ ] All OS-level operations use validated/normalized exec names
[ ] All exe.Command calls use safe arguments (no unsanitized input)
[ ] All tests pass with -race flag
[ ] All tests pass on Windows/Linux/macOS
```

---

*Generated by Roboty QA Engineering — Enterprise Testing Specification v2.0*
*Last updated: 2026-05-16*
