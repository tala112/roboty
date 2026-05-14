package modes

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"Roboty/internal/db"

	"github.com/google/uuid"
)

type ModeService struct {
	database *db.DB
	queries  *db.Queries
	ctx      context.Context
	mu       sync.Mutex
	timers   map[string]*time.Timer
	notifMuted bool
}

func NewModeService(database *db.DB, queries *db.Queries) *ModeService {
	return &ModeService{
		database: database,
		queries:  queries,
		timers:   make(map[string]*time.Timer),
	}
}

func (s *ModeService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *ModeService) InitFocusSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS focus_modes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		duration_minutes INTEGER DEFAULT 0,
		mute_notifications INTEGER DEFAULT 0,
		enabled INTEGER DEFAULT 0,
		icon TEXT DEFAULT 'shield',
		color TEXT DEFAULT '#6366f1',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS focus_mode_apps (
		id TEXT PRIMARY KEY,
		mode_id TEXT NOT NULL REFERENCES focus_modes(id) ON DELETE CASCADE,
		app_name TEXT NOT NULL,
		app_exec TEXT NOT NULL,
		close_on_activate INTEGER DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(mode_id, app_exec)
	);
	CREATE TABLE IF NOT EXISTS focus_mode_sessions (
		id TEXT PRIMARY KEY,
		mode_id TEXT NOT NULL REFERENCES focus_modes(id) ON DELETE CASCADE,
		started_at TEXT NOT NULL DEFAULT (datetime('now')),
		ends_at TEXT,
		finished_at TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_focus_mode_apps_mode_id ON focus_mode_apps(mode_id);
	CREATE INDEX IF NOT EXISTS idx_focus_mode_sessions_mode_id ON focus_mode_sessions(mode_id);
	CREATE INDEX IF NOT EXISTS idx_focus_mode_sessions_status ON focus_mode_sessions(status);
	CREATE TRIGGER IF NOT EXISTS update_focus_modes_updated_at
	AFTER UPDATE ON focus_modes
	BEGIN
		UPDATE focus_modes SET updated_at = datetime('now') WHERE id = NEW.id;
	END;`

	_, err := s.database.DB().Exec(schema)
	if err != nil {
		return fmt.Errorf("init focus schema: %w", err)
	}
	log.Println("[modes] Focus schema initialized")
	return nil
}

func (s *ModeService) CheckResumeSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.queries.GetActiveFocusSession(context.Background())
	if err != nil || session == nil {
		return
	}

	log.Printf("[modes] Found active session %s for mode %s, checking status...", session.ID, session.ModeID)

	if session.EndsAt != nil {
		endTime, err := time.Parse("2006-01-02 15:04:05", session.EndsAt.String())
		if err == nil && time.Now().After(endTime) {
			log.Printf("[modes] Session timer expired, auto-deactivating")
			s.queries.UpdateFocusSessionStatus(context.Background(), session.ID, "completed")
			modeApps, _ := s.getModeAppExecs(session.ModeID)
			UnblockApps(modeApps)
			if s.notifMuted {
				RestoreNotifications()
				s.notifMuted = false
			}
			return
		}
		remaining := time.Until(endTime)
		if remaining > 0 {
			log.Printf("[modes] Resuming session timer, %v remaining", remaining)
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
	return &result, nil
}

func (s *ModeService) CreateMode(req CreateModeRequest) (*FocusMode, error) {
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
		}
		_, err := s.queries.CreateFocusModeApp(context.Background(), appParams)
		if err != nil {
			log.Printf("[modes] Failed to add app %s: %v", app.AppName, err)
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
		}
		_, err := s.queries.CreateFocusModeApp(context.Background(), appParams)
		if err != nil {
			log.Printf("[modes] Failed to add app %s: %v", app.AppName, err)
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

	apps, err := s.queries.GetFocusModeAppsByModeID(context.Background(), modeID)
	if err != nil {
		return nil, fmt.Errorf("get apps for mode %s: %w", modeID, err)
	}

	execNames := make([]string, 0, len(apps))
	for _, a := range apps {
		execNames = append(execNames, a.AppExec)
	}

	var closeAppsList []string
	for _, a := range apps {
		if a.CloseOnActivate {
			closeAppsList = append(closeAppsList, a.AppExec)
		}
	}
	if len(closeAppsList) > 0 {
		CloseApps(closeAppsList)
	}

	BlockApps(execNames)

	var endsAt *string
	if m.DurationMinutes > 0 {
		endTime := time.Now().Add(time.Duration(m.DurationMinutes) * time.Minute)
		formatted := endTime.Format("2006-01-02 15:04:05")
		endsAt = &formatted
	}

	if m.MuteNotifications && !s.notifMuted {
		MuteNotifications()
		s.notifMuted = true
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

	log.Printf("[modes] Activated mode %s, session %s, blocks %d apps", m.Name, sessionID, len(apps))

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

	_, err = s.queries.UpdateFocusSessionStatus(context.Background(), sessionID, "completed")
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	if timer, ok := s.timers[session.ModeID]; ok {
		timer.Stop()
		delete(s.timers, session.ModeID)
	}

	apps, _ := s.queries.GetFocusModeAppsByModeID(context.Background(), session.ModeID)
	execNames := make([]string, 0, len(apps))
	for _, a := range apps {
		execNames = append(execNames, a.AppExec)
	}
	UnblockApps(execNames)

	if s.notifMuted {
		RestoreNotifications()
		s.notifMuted = false
	}

	log.Printf("[modes] Deactivated session %s", sessionID)
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

func (s *ModeService) GetInstalledApps() ([]InstalledApp, error) {
	return GetInstalledApps()
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
		s.queries.UpdateFocusSessionStatus(context.Background(), session.ID, "cancelled")
		if timer, ok := s.timers[modeID]; ok {
			timer.Stop()
			delete(s.timers, modeID)
		}
		execs, _ := s.getModeAppExecs(modeID)
		UnblockApps(execs)
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
		log.Printf("[modes] Timer expired for mode %s session %s", modeID, sessionID)
		s.mu.Lock()
		defer s.mu.Unlock()

		_, err := s.queries.UpdateFocusSessionStatus(context.Background(), sessionID, "completed")
		if err != nil {
			log.Printf("[modes] Failed to complete session %s: %v", sessionID, err)
		}
		execs, _ := s.getModeAppExecs(modeID)
		UnblockApps(execs)
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
		existing, err := s.queries.GetFocusModeByID(context.Background(), modeID)
		if err == nil && existing != nil {
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
		log.Printf("[modes] Timer auto-deactivated mode %s", modeID)
	})
}
