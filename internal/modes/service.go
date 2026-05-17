package modes

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"Roboty/internal/db"

	"github.com/google/uuid"
)

type ModeService struct {
	database     *db.DB
	queries     FocusDataStore
	ctx         context.Context
	mu          sync.Mutex
	timers      map[string]*time.Timer
	notifMuted  bool
	tracker     *ForegroundTracker
	appBlocker  *AppBlocker
	urlBlocker  *URLBlocker
	killer      ProcessKiller
	proxyMgr    SystemProxyManager
	notifMgr    NotificationManager
}

func NewModeService(database *db.DB, queries FocusDataStore) *ModeService {
	tracker := NewForegroundTracker()
	return &ModeService{
		database:    database,
		queries:    queries,
		timers:     make(map[string]*time.Timer),
		tracker:    tracker,
		appBlocker: NewAppBlocker(tracker),
		urlBlocker: NewURLBlocker(),
		killer:     NewRealProcessKiller(),
		proxyMgr:   NewRealProxyManager(),
		notifMgr:   NewRealNotificationManager(),
	}
}

func NewModeServiceWithDI(
	database *db.DB,
	queries FocusDataStore,
	tracker *ForegroundTracker,
	killer ProcessKiller,
	proxyMgr SystemProxyManager,
	notifMgr NotificationManager,
) *ModeService {
	return &ModeService{
		database:    database,
		queries:    queries,
		timers:     make(map[string]*time.Timer),
		tracker:    tracker,
		appBlocker: NewAppBlockerWithDI(tracker, killer),
		urlBlocker: NewURLBlockerWithDI(proxyMgr, NewFileStateManager()),
		killer:     killer,
		proxyMgr:   proxyMgr,
		notifMgr:   notifMgr,
	}
}

func (s *ModeService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *ModeService) InitFocusSchema() error {
	// Add is_allowed column if it doesn't exist yet
	alterSQL := `ALTER TABLE focus_mode_apps ADD COLUMN is_allowed INTEGER DEFAULT 0`
	if _, err := s.database.DB().Exec(alterSQL); err != nil {
		// Column likely already exists — ignore error
		log.Println("[modes] is_allowed column may already exist (safe to ignore)")
	}

	// Create focus_mode_urls table
	urlsSchema := `
	CREATE TABLE IF NOT EXISTS focus_mode_urls (
		id TEXT PRIMARY KEY,
		mode_id TEXT NOT NULL REFERENCES focus_modes(id) ON DELETE CASCADE,
		url TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(mode_id, url)
	);
	CREATE INDEX IF NOT EXISTS idx_focus_mode_urls_mode_id ON focus_mode_urls(mode_id);`
	if _, err := s.database.DB().Exec(urlsSchema); err != nil {
		return fmt.Errorf("init focus urls schema: %w", err)
	}

	log.Println("[modes] Focus URL schema initialized")
	return nil
}

// EmergencyStop performs an immediate, safe shutdown of all active focus mode
// subsystems. Called by the watchdog failsafe or when critical failures are detected.
func (s *ModeService) EmergencyStop(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[modes] 🚨 EMERGENCY STOP triggered: %s", reason)

	// Stop app blocker immediately
	s.appBlocker.Stop()

	// Stop URL blocker and disable system proxy
	if err := s.urlBlocker.Stop(); err != nil {
		log.Printf("[modes] ⚠️ EmergencyStop: URLBlocker stop warning: %v", err)
	}

	// Restore notifications
	if s.notifMuted {
		RestoreNotifications()
		s.notifMuted = false
	}

	// Complete all active sessions
	if s.queries != nil {
		sessions, err := s.queries.GetAllFocusSessions(context.Background())
		if err == nil {
			for _, session := range sessions {
				if session.Status == "active" {
					modeName := session.ModeID
					if mode, err := s.queries.GetFocusModeByID(context.Background(), session.ModeID); err == nil && mode != nil {
						modeName = mode.Name
					}
					s.queries.UpdateFocusSessionStatus(context.Background(), session.ID, "emergency_stopped")
					log.Printf("[modes] Emergency stop: terminated session %s (mode: %q)", session.ID, modeName)
				}
			}
		}
	}

	log.Printf("[modes] ✅ Emergency stop complete: %s", reason)
}

func (s *ModeService) CheckResumeSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean up orphaned proxy settings from previous crash
	CleanupOrphanedProxy()

	session, err := s.queries.GetActiveFocusSession(context.Background())
	if err != nil || session == nil {
		return
	}

	modeName := session.ModeID
	if mode, err := s.queries.GetFocusModeByID(context.Background(), session.ModeID); err == nil && mode != nil {
		modeName = mode.Name
	}
	log.Printf("[modes] 🔄 Found active session %s (mode: %q), checking status...", session.ID, modeName)

	if session.EndsAt != nil {
		endTime, err := time.Parse("2006-01-02 15:04:05", session.EndsAt.String())
		if err == nil && time.Now().After(endTime) {
			log.Printf("[modes] ⏰ Session timer already expired for %q — auto-deactivating", modeName)
			s.queries.UpdateFocusSessionStatus(context.Background(), session.ID, "completed")
			s.appBlocker.Stop()
			s.urlBlocker.Stop()
			if s.notifMuted {
				RestoreNotifications()
				s.notifMuted = false
			}
			return
		}
		remaining := time.Until(endTime)
		if remaining > 0 {
			log.Printf("[modes] 🔄 Resuming session timer for %q, %v remaining", modeName, remaining)
			s.startTimer(session.ID, session.ModeID, remaining)
		}
	}
}

func (s *ModeService) ListModes() ([]FocusMode, error) {
	modes, err := s.queries.GetAllFocusModes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list modes: %w", err)
	}
	result := make([]FocusMode, len(modes))
	for i, m := range modes {
		result[i] = s.toFocusMode(&m)
		apps, _ := s.queries.GetFocusModeAppsByModeID(context.Background(), m.ID)
		for _, a := range apps {
			result[i].Apps = append(result[i].Apps, s.toFocusModeApp(&a))
		}
		urls, _ := s.queries.GetFocusModeURLsByModeID(context.Background(), m.ID)
		for _, u := range urls {
			result[i].AllowedURLs = append(result[i].AllowedURLs, u.URL)
		}
	}
	return result, nil
}

func (s *ModeService) GetMode(id string) (*FocusMode, error) {
	m, err := s.queries.GetFocusModeByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("get mode %s: %w", id, err)
	}
	if m == nil {
		return nil, nil
	}
	result := s.toFocusMode(m)
	apps, _ := s.queries.GetFocusModeAppsByModeID(context.Background(), id)
	for _, a := range apps {
		result.Apps = append(result.Apps, s.toFocusModeApp(&a))
	}
	urls, _ := s.queries.GetFocusModeURLsByModeID(context.Background(), id)
	for _, u := range urls {
		result.AllowedURLs = append(result.AllowedURLs, u.URL)
	}
	return &result, nil
}

// validateAppExec checks that an app exec name is safe to store and potentially kill.
// Returns a user-friendly error if the name is dangerous or system-critical.
func validateAppExec(appExec, appName string) error {
	safeExec := NormalizeKillExec(appExec)
	if safeExec == "" {
		return fmt.Errorf("app exec %q is invalid or contains dangerous characters", appExec)
	}
	safe, reason := GetGlobalSafetyVerifier().IsSafeToKill(safeExec)
	if !safe {
		return fmt.Errorf("cannot add %q (%s): %s", appName, appExec, reason)
	}
	return nil
}

func (s *ModeService) CreateMode(req CreateModeRequest) (*FocusMode, error) {
	// Validate all app execs before creating
	for _, app := range req.Apps {
		if err := validateAppExec(app.AppExec, app.AppName); err != nil {
			return nil, err
		}
	}

	id := uuid.New().String()
	params := db.CreateFocusModeParams{
		ID:                id,
		Name:             req.Name,
		Description:      req.Description,
		DurationMinutes:   req.DurationMinutes,
		MuteNotifications: req.MuteNotifications,
		Enabled:          req.Enabled,
		Icon:             req.Icon,
		Color:            req.Color,
	}
	_, err := s.queries.CreateFocusMode(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("create mode: %w", err)
	}
	if req.Enabled {
		s.disableOtherModes(id)
	}

	for _, app := range req.Apps {
		appParams := db.CreateFocusModeAppParams{
			ID:               uuid.New().String(),
			ModeID:          id,
			AppName:         app.AppName,
			AppExec:         app.AppExec,
			CloseOnActivate: app.CloseOnActivate,
			IsAllowed:       app.IsAllowed,
		}
		_, err := s.queries.CreateFocusModeApp(context.Background(), appParams)
		if err != nil {
			log.Printf("[modes] Failed to add app %s: %v", app.AppName, err)
		}
	}

	for _, url := range req.AllowedURLs {
		urlParams := db.CreateFocusModeURLParams{
			ID:      uuid.New().String(),
			ModeID:  id,
			URL:     url,
		}
		_, err := s.queries.CreateFocusModeURL(context.Background(), urlParams)
		if err != nil {
			log.Printf("[modes] Failed to add URL %s: %v", url, err)
		}
	}

	return s.GetMode(id)
}

func (s *ModeService) UpdateMode(id string, req UpdateModeRequest) (*FocusMode, error) {
	updateParams := db.UpdateFocusModeParams{
		ID:                id,
		Name:             req.Name,
		Description:      req.Description,
		DurationMinutes:   req.DurationMinutes,
		MuteNotifications: req.MuteNotifications,
		Enabled:          req.Enabled,
		Icon:             req.Icon,
		Color:            req.Color,
	}
	_, err := s.queries.UpdateFocusMode(context.Background(), updateParams)
	if err != nil {
		return nil, fmt.Errorf("update mode %s: %w", id, err)
	}

	// Replace apps
	if err := s.queries.DeleteFocusModeAppsByModeID(context.Background(), id); err != nil {
		log.Printf("[modes] Failed to clear apps for mode %s: %v", id, err)
	}
	for _, app := range req.Apps {
		appParams := db.CreateFocusModeAppParams{
			ID:               uuid.New().String(),
			ModeID:          id,
			AppName:         app.AppName,
			AppExec:         app.AppExec,
			CloseOnActivate: app.CloseOnActivate,
			IsAllowed:       app.IsAllowed,
		}
		_, err := s.queries.CreateFocusModeApp(context.Background(), appParams)
		if err != nil {
			log.Printf("[modes] Failed to add app %s: %v", app.AppName, err)
		}
	}

	// Replace URLs
	if err := s.queries.DeleteFocusModeURLsByModeID(context.Background(), id); err != nil {
		log.Printf("[modes] Failed to clear URLs for mode %s: %v", id, err)
	}
	for _, url := range req.AllowedURLs {
		urlParams := db.CreateFocusModeURLParams{
			ID:      uuid.New().String(),
			ModeID:  id,
			URL:     url,
		}
		_, err := s.queries.CreateFocusModeURL(context.Background(), urlParams)
		if err != nil {
			log.Printf("[modes] Failed to add URL %s: %v", url, err)
		}
	}

	if req.Enabled {
		s.disableOtherModes(id)
	}

	return s.GetMode(id)
}

func (s *ModeService) DeleteMode(id string) error {
	s.mu.Lock()
	if timer, ok := s.timers[id]; ok {
		timer.Stop()
		delete(s.timers, id)
	}
	s.mu.Unlock()

	if err := s.queries.DeleteFocusMode(context.Background(), id); err != nil {
		return fmt.Errorf("delete mode %s: %w", id, err)
	}
	return nil
}

func (s *ModeService) ToggleMode(id string, enabled bool) error {
	m, err := s.queries.GetFocusModeByID(context.Background(), id)
	if err != nil || m == nil {
		return fmt.Errorf("mode %s not found", id)
	}

	if enabled {
		s.disableOtherModes(id)
	} else {
		s.deactivateIfActive(id)
	}

	updateParams := db.UpdateFocusModeParams{
		ID:                id,
		Name:             m.Name,
		Description:      m.Description,
		DurationMinutes:   m.DurationMinutes,
		MuteNotifications: m.MuteNotifications,
		Enabled:          enabled,
		Icon:             m.Icon,
		Color:            m.Color,
	}
	_, err = s.queries.UpdateFocusMode(context.Background(), updateParams)
	return err
}

func (s *ModeService) ActivateMode(modeID string) (*FocusSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.queries.GetFocusModeByID(context.Background(), modeID)
	if err != nil || m == nil {
		return nil, fmt.Errorf("mode %s not found", modeID)
	}

	// Get allowed apps (is_allowed = true)
	allowedApps, err := s.queries.GetFocusModeAllowedAppsByModeID(context.Background(), modeID)
	if err != nil {
		return nil, fmt.Errorf("get allowed apps for mode %s: %w", modeID, err)
	}

	// Get all apps for close-on-activate
	allApps, err := s.queries.GetFocusModeAppsByModeID(context.Background(), modeID)
	if err != nil {
		return nil, fmt.Errorf("get apps for mode %s: %w", modeID, err)
	}

	// Build exec name lists
	allowedExecs := make([]string, 0, len(allowedApps))
	for _, a := range allowedApps {
		allowedExecs = append(allowedExecs, strings.ToLower(a.AppExec))
	}

	if len(allowedExecs) > 0 {
		log.Printf("[modes] Allowed apps: %v", allowedExecs)
	} else {
		log.Printf("[modes] ⚠️ No allowed apps — all non-whitelisted apps will be blocked")
	}

	var closeAppsList []string
	for _, a := range allApps {
		if a.CloseOnActivate {
			// SAFETY: Verify the exec name is safe before adding to close list
			safeExec := NormalizeKillExec(a.AppExec)
			if safeExec == "" {
				log.Printf("[modes] SKIP closeOnActivate for %q: rejected by NormalizeKillExec", a.AppExec)
				continue
			}
			safe, reason := GetGlobalSafetyVerifier().IsSafeToKill(safeExec)
			if !safe {
				log.Printf("[modes] SAFETY BLOCKED closeOnActivate for %q: %s", a.AppExec, reason)
				continue
			}
			closeAppsList = append(closeAppsList, a.AppExec)
			// Always allow apps that close on activate (they get killed and stay dead)
		}
	}

	// Close apps marked for closing
	if len(closeAppsList) > 0 {
		CloseApps(closeAppsList)
	}

	// Stop any existing blocker/proxy before (re)starting with new config
	s.appBlocker.Stop()
	s.appBlocker.Start(allowedExecs, closeAppsList, 2*time.Second)

	// Start URL blocker (whitelist mode — block ALL except allowed)
	allowedURLs, err := s.queries.GetFocusModeURLsByModeID(context.Background(), modeID)
	if err != nil {
		return nil, fmt.Errorf("get allowed URLs for mode %s: %w", modeID, err)
	}
	urlStrs := make([]string, 0, len(allowedURLs))
	for _, u := range allowedURLs {
		urlStrs = append(urlStrs, u.URL)
	}
	if len(urlStrs) > 0 {
		log.Printf("[modes] Allowed URLs: %v", urlStrs)
	} else {
		log.Printf("[modes] No allowed URLs configured — no URL blocking")
	}
	// Always stop existing proxy first to ensure clean state + port reuse
	s.urlBlocker.Stop()
	if len(urlStrs) > 0 {
		if err := s.urlBlocker.Start(DefaultProxyPort, urlStrs); err != nil {
			log.Printf("[modes] Failed to start URL blocker: %v — app blocking still active", err)
		}
	}

	// Mute notifications if configured
	if m.MuteNotifications && !s.notifMuted {
		MuteNotifications()
		s.notifMuted = true
	}

	// Create session and timer
	var endsAt *string
	if m.DurationMinutes > 0 {
		endTime := time.Now().Add(time.Duration(m.DurationMinutes) * time.Minute)
		formatted := endTime.Format("2006-01-02 15:04:05")
		endsAt = &formatted
	}

	sessionID := uuid.New().String()
	sessionParams := db.CreateFocusSessionParams{
		ID:      sessionID,
		ModeID:  modeID,
		EndsAt:  endsAt,
		Status:  "active",
	}
	session, err := s.queries.CreateFocusSession(context.Background(), sessionParams)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if endsAt != nil {
		duration := time.Duration(m.DurationMinutes) * time.Minute
		s.startTimer(sessionID, modeID, duration)
	}

	durationStr := "no limit"
	if m.DurationMinutes > 0 {
		durationStr = fmt.Sprintf("%dm", m.DurationMinutes)
	}
	log.Printf("[modes] 🔵 MODE STARTED: %q (session: %s, apps: %d, urls: %d, duration: %s)",
		m.Name, sessionID, len(allowedExecs), len(urlStrs), durationStr)

	return &FocusSession{
		ID:        session.ID,
		ModeID:    session.ModeID,
		StartedAt: session.StartedAt.String(),
		EndsAt:    endsAt,
		Status:    session.Status,
	}, nil
}

func (s *ModeService) DeactivateMode(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.queries.GetFocusSessionByID(context.Background(), sessionID)
	if err != nil || session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Get mode name before cleanup for logging
	modeName := session.ModeID
	if mode, err := s.queries.GetFocusModeByID(context.Background(), session.ModeID); err == nil && mode != nil {
		modeName = mode.Name
	}

	log.Printf("[modes] 🟢 MODE ENDING: %q (session: %s)", modeName, sessionID)

	_, err = s.queries.UpdateFocusSessionStatus(context.Background(), sessionID, "completed")
	if err != nil {
		log.Printf("[modes] Failed to update session status: %v", err)
	}

	// Stop timers
	if timer, ok := s.timers[session.ModeID]; ok {
		timer.Stop()
		delete(s.timers, session.ModeID)
	}

	// Stop continuous app blocker
	s.appBlocker.Stop()

	// Stop URL blocker — always attempt, log failures but don't abort cleanup
	if err := s.urlBlocker.Stop(); err != nil {
		log.Printf("[modes] ⚠️ URL blocker stop warning: %v", err)
	}

	// Restore notifications
	if s.notifMuted {
		RestoreNotifications()
		s.notifMuted = false
	}

	log.Printf("[modes] 🟢 MODE ENDED: %q (session: %s)", modeName, sessionID)
	return nil
}

func (s *ModeService) GetActiveSession() (*FocusSession, error) {
	session, err := s.queries.GetActiveFocusSession(context.Background())
	if err != nil || session == nil {
		return nil, nil
	}
	result := &FocusSession{
		ID:        session.ID,
		ModeID:    session.ModeID,
		StartedAt: session.StartedAt.String(),
		Status:    session.Status,
	}
	if session.EndsAt != nil {
		ends := session.EndsAt.String()
		result.EndsAt = &ends
	}
	return result, nil
}

// GetInstalledApps returns apps from app-mappings that also exist on user's PC
func (s *ModeService) GetInstalledApps() ([]InstalledApp, error) {
	// 1. Apps from app-mappings.json (known apps catalog)
	mappedApps, _ := GetAppsFromMappings()

	// 2. Apps installed on system (.lnk / .desktop)
	systemApps, err := GetInstalledApps()
	if err != nil {
		log.Printf("[modes] failed to list installed apps: %v", err)
		systemApps = nil
	}

	// 3. Currently running processes
	runningApps, err := s.tracker.ListRunningProcesses()
	if err != nil {
		log.Printf("[modes] failed to list running processes: %v", err)
		runningApps = nil
	}

	// Build set of apps actually on user PC (installed + running)
	pcSet := make(map[string]bool)
	for _, a := range systemApps {
		pcSet[strings.ToLower(a.Exec)] = true
	}
	for _, a := range runningApps {
		pcSet[strings.ToLower(a.Exec)] = true
	}

	// Filter out whitelisted apps (never show / never block)
	whitelist := GetWhitelistExecs()

	// Only show mapped apps that also exist on the user's PC
	merged := make([]InstalledApp, 0, len(mappedApps))
	for _, app := range mappedApps {
		key := strings.ToLower(app.Exec)
		if whitelist[key] {
			continue
		}
		if pcSet[key] {
			merged = append(merged, app)
		}
	}

	return merged, nil
}

// CheckAppOnPC checks if a given app exec name exists on the user's PC
func (s *ModeService) CheckAppOnPC(appExec string) bool {
	key := strings.ToLower(appExec)

	// Check in installed apps
	systemApps, err := GetInstalledApps()
	if err == nil {
		for _, a := range systemApps {
			if strings.ToLower(a.Exec) == key {
				return true
			}
		}
	}

	// Check in running processes
	runningApps, err := s.tracker.ListRunningProcesses()
	if err == nil {
		for _, a := range runningApps {
			if strings.ToLower(a.Exec) == key {
				return true
			}
		}
	}

	return false
}

// AddToAppMappings persists an app to app-mappings.json (Velosi format)
func (s *ModeService) AddToAppMappings(appName, appExec string, category string) error {
	// Use a default category if none provided
	if category == "" {
		category = "productive"
	}
	// This updates the app-mappings.json file
	return addAppToMappingsFile(appName, appExec, category)
}

func (s *ModeService) toFocusMode(m *db.FocusMode) FocusMode {
	return FocusMode{
		ID:                m.ID,
		Name:             m.Name,
		Description:      m.Description,
		DurationMinutes:   m.DurationMinutes,
		MuteNotifications: m.MuteNotifications,
		Enabled:          m.Enabled,
		Icon:             m.Icon,
		Color:            m.Color,
		CreatedAt:        m.CreatedAt.String(),
		UpdatedAt:        m.UpdatedAt.String(),
	}
}

func (s *ModeService) toFocusModeApp(a *db.FocusModeApp) FocusModeApp {
	return FocusModeApp{
		ID:               a.ID,
		ModeID:          a.ModeID,
		AppName:         a.AppName,
		AppExec:         a.AppExec,
		CloseOnActivate: a.CloseOnActivate,
		IsAllowed:       a.IsAllowed,
	}
}

func (s *ModeService) getModeAppExecs(modeID string) ([]string, error) {
	apps, err := s.queries.GetFocusModeAppsByModeID(context.Background(), modeID)
	if err != nil {
		return nil, err
	}
	execs := make([]string, len(apps))
	for i, a := range apps {
		execs[i] = a.AppExec
	}
	return execs, nil
}

func (s *ModeService) disableOtherModes(keepID string) {
	modes, err := s.queries.GetAllFocusModes(context.Background())
	if err != nil {
		return
	}
	for _, m := range modes {
		if m.ID != keepID && m.Enabled {
			upd := db.UpdateFocusModeParams{
				ID:                m.ID,
				Name:             m.Name,
				Description:      m.Description,
				DurationMinutes:   m.DurationMinutes,
				MuteNotifications: m.MuteNotifications,
				Enabled:          false,
				Icon:             m.Icon,
				Color:            m.Color,
			}
			s.queries.UpdateFocusMode(context.Background(), upd)
			s.deactivateIfActive(m.ID)
		}
	}
}

func (s *ModeService) deactivateIfActive(modeID string) {
	session, err := s.queries.GetActiveFocusSession(context.Background())
	if err == nil && session != nil && session.ModeID == modeID {
		modeName := modeID
		if mode, err := s.queries.GetFocusModeByID(context.Background(), modeID); err == nil && mode != nil {
			modeName = mode.Name
		}
		log.Printf("[modes] 🟢 MODE CANCELLED: %q (session: %s)", modeName, session.ID)
		s.queries.UpdateFocusSessionStatus(context.Background(), session.ID, "cancelled")
		if timer, ok := s.timers[modeID]; ok {
			timer.Stop()
			delete(s.timers, modeID)
		}
		s.appBlocker.Stop()
		if err := s.urlBlocker.Stop(); err != nil {
			log.Printf("[modes] ⚠️ URL blocker stop warning on deactivate: %v", err)
		}
		if s.notifMuted {
			RestoreNotifications()
			s.notifMuted = false
		}
	}
}

func (s *ModeService) startTimer(sessionID, modeID string, duration time.Duration) {
	if timer, ok := s.timers[modeID]; ok {
		timer.Stop()
	}
	s.timers[modeID] = time.AfterFunc(duration, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Get mode name for logging
		modeName := modeID
		existing, err := s.queries.GetFocusModeByID(context.Background(), modeID)
		if err == nil && existing != nil {
			modeName = existing.Name
		}

		log.Printf("[modes] ⏰ MODE TIMER EXPIRED: %q (session: %s)", modeName, sessionID)

		_, err = s.queries.UpdateFocusSessionStatus(context.Background(), sessionID, "completed")
		if err != nil {
			log.Printf("[modes] Failed to complete session %s: %v", sessionID, err)
		}

		s.appBlocker.Stop()
		if err := s.urlBlocker.Stop(); err != nil {
			log.Printf("[modes] ⚠️ URL blocker stop warning on timer expiry: %v", err)
		}

		if s.notifMuted {
			RestoreNotifications()
			s.notifMuted = false
		}

		upd := db.UpdateFocusModeParams{
			ID:                modeID,
			Name:             "",
			Description:      "",
			DurationMinutes:   0,
			MuteNotifications: false,
			Enabled:          false,
			Icon:             "",
			Color:            "",
		}
		if existing != nil {
			upd.Name = existing.Name
			upd.Description = existing.Description
			upd.DurationMinutes = existing.DurationMinutes
			upd.MuteNotifications = existing.MuteNotifications
			upd.Icon = existing.Icon
			upd.Color = existing.Color
			upd.Enabled = false
			s.queries.UpdateFocusMode(context.Background(), upd)
		}

		delete(s.timers, modeID)
		log.Printf("[modes] ✅ Mode auto-deactivated: %q", modeName)
	})
}

func (s *ModeService) GetURLBlockerRunning() bool {
	return s.urlBlocker.IsRunning()
}
