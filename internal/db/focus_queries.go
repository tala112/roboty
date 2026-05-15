package db

import (
	"context"
	"database/sql"
)

// Focus mode queries
const (
	createFocusMode = `
INSERT INTO focus_modes (id, name, description, duration_minutes, mute_notifications, enabled, icon, color, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, datetime('now'), datetime('now'))
RETURNING *`

	getFocusModeByID = `
SELECT id, name, description, duration_minutes, mute_notifications, enabled, icon, color, created_at, updated_at
FROM focus_modes WHERE id = $1`

	getAllFocusModes = `
SELECT id, name, description, duration_minutes, mute_notifications, enabled, icon, color, created_at, updated_at
FROM focus_modes ORDER BY created_at ASC`

	updateFocusMode = `
UPDATE focus_modes SET
    name = $2, description = $3, duration_minutes = $4,
    mute_notifications = $5, enabled = $6, icon = $7, color = $8,
    updated_at = datetime('now')
WHERE id = $1
RETURNING *`

	deleteFocusMode = `DELETE FROM focus_modes WHERE id = $1`
)

// Focus mode app queries
const (
	createFocusModeApp = `
INSERT INTO focus_mode_apps (id, mode_id, app_name, app_exec, close_on_activate, is_allowed, created_at)
VALUES ($1, $2, $3, $4, $5, $6, datetime('now'))
RETURNING *`

	getFocusModeAppsByModeID = `
SELECT id, mode_id, app_name, app_exec, close_on_activate, created_at, is_allowed
FROM focus_mode_apps WHERE mode_id = $1 ORDER BY app_name ASC`

	getFocusModeAllowedAppsByModeID = `
SELECT id, mode_id, app_name, app_exec, close_on_activate, created_at, is_allowed
FROM focus_mode_apps WHERE mode_id = $1 AND is_allowed = 1 ORDER BY app_name ASC`

	deleteFocusModeApp = `DELETE FROM focus_mode_apps WHERE id = $1`

	deleteFocusModeAppsByModeID = `DELETE FROM focus_mode_apps WHERE mode_id = $1`
)

// Focus mode URL queries
const (
	createFocusModeURL = `
INSERT INTO focus_mode_urls (id, mode_id, url, created_at)
VALUES ($1, $2, $3, datetime('now'))
RETURNING *`

	getFocusModeURLsByModeID = `
SELECT id, mode_id, url, created_at
FROM focus_mode_urls WHERE mode_id = $1 ORDER BY url ASC`

	deleteFocusModeURLsByModeID = `DELETE FROM focus_mode_urls WHERE mode_id = $1`
)

// Focus mode session queries
const (
	createFocusSession = `
INSERT INTO focus_mode_sessions (id, mode_id, started_at, ends_at, finished_at, status, created_at)
VALUES ($1, $2, datetime('now'), $3, NULL, $4, datetime('now'))
RETURNING *`

	getFocusSessionByID = `
SELECT id, mode_id, started_at, ends_at, finished_at, status, created_at
FROM focus_mode_sessions WHERE id = $1`

	getActiveFocusSession = `
SELECT id, mode_id, started_at, ends_at, finished_at, status, created_at
FROM focus_mode_sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1`

	getLatestSessionByModeID = `
SELECT id, mode_id, started_at, ends_at, finished_at, status, created_at
FROM focus_mode_sessions WHERE mode_id = $1 ORDER BY started_at DESC LIMIT 1`

	updateFocusSessionStatus = `
UPDATE focus_mode_sessions SET status = $2, finished_at = CASE WHEN $2 = 'completed' OR $2 = 'cancelled' THEN datetime('now') ELSE NULL END WHERE id = $1
RETURNING *`
)

func (q *Queries) CreateFocusMode(ctx context.Context, arg CreateFocusModeParams) (*FocusMode, error) {
	row := q.db.QueryRowContext(ctx, createFocusMode,
		arg.ID, arg.Name, arg.Description, arg.DurationMinutes,
		arg.MuteNotifications, arg.Enabled, arg.Icon, arg.Color,
	)
	var m FocusMode
	err := row.Scan(
		&m.ID, &m.Name, &m.Description, &m.DurationMinutes,
		&m.MuteNotifications, &m.Enabled, &m.Icon, &m.Color,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return &m, err
}

func (q *Queries) GetFocusModeByID(ctx context.Context, id string) (*FocusMode, error) {
	row := q.db.QueryRowContext(ctx, getFocusModeByID, id)
	var m FocusMode
	err := row.Scan(
		&m.ID, &m.Name, &m.Description, &m.DurationMinutes,
		&m.MuteNotifications, &m.Enabled, &m.Icon, &m.Color,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (q *Queries) GetAllFocusModes(ctx context.Context) ([]FocusMode, error) {
	rows, err := q.db.QueryContext(ctx, getAllFocusModes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var modes []FocusMode
	for rows.Next() {
		var m FocusMode
		err := rows.Scan(
			&m.ID, &m.Name, &m.Description, &m.DurationMinutes,
			&m.MuteNotifications, &m.Enabled, &m.Icon, &m.Color,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		modes = append(modes, m)
	}
	return modes, rows.Err()
}

func (q *Queries) UpdateFocusMode(ctx context.Context, arg UpdateFocusModeParams) (*FocusMode, error) {
	row := q.db.QueryRowContext(ctx, updateFocusMode,
		arg.ID, arg.Name, arg.Description, arg.DurationMinutes,
		arg.MuteNotifications, arg.Enabled, arg.Icon, arg.Color,
	)
	var m FocusMode
	err := row.Scan(
		&m.ID, &m.Name, &m.Description, &m.DurationMinutes,
		&m.MuteNotifications, &m.Enabled, &m.Icon, &m.Color,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return &m, err
}

func (q *Queries) DeleteFocusMode(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteFocusMode, id)
	return err
}

func (q *Queries) CreateFocusModeApp(ctx context.Context, arg CreateFocusModeAppParams) (*FocusModeApp, error) {
	isAllowed := 0
	if arg.IsAllowed {
		isAllowed = 1
	}
	row := q.db.QueryRowContext(ctx, createFocusModeApp,
		arg.ID, arg.ModeID, arg.AppName, arg.AppExec, arg.CloseOnActivate, isAllowed,
	)
	var a FocusModeApp
	err := row.Scan(
		&a.ID, &a.ModeID, &a.AppName, &a.AppExec, &a.CloseOnActivate, &a.CreatedAt, &a.IsAllowed,
	)
	return &a, err
}

func (q *Queries) GetFocusModeAppsByModeID(ctx context.Context, modeID string) ([]FocusModeApp, error) {
	rows, err := q.db.QueryContext(ctx, getFocusModeAppsByModeID, modeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []FocusModeApp
	for rows.Next() {
		var a FocusModeApp
		err := rows.Scan(
			&a.ID, &a.ModeID, &a.AppName, &a.AppExec, &a.CloseOnActivate, &a.CreatedAt, &a.IsAllowed,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (q *Queries) GetFocusModeAllowedAppsByModeID(ctx context.Context, modeID string) ([]FocusModeApp, error) {
	rows, err := q.db.QueryContext(ctx, getFocusModeAllowedAppsByModeID, modeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []FocusModeApp
	for rows.Next() {
		var a FocusModeApp
		err := rows.Scan(
			&a.ID, &a.ModeID, &a.AppName, &a.AppExec, &a.CloseOnActivate, &a.CreatedAt, &a.IsAllowed,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (q *Queries) DeleteFocusModeApp(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteFocusModeApp, id)
	return err
}

func (q *Queries) DeleteFocusModeAppsByModeID(ctx context.Context, modeID string) error {
	_, err := q.db.ExecContext(ctx, deleteFocusModeAppsByModeID, modeID)
	return err
}

// Focus mode URL functions
func (q *Queries) CreateFocusModeURL(ctx context.Context, arg CreateFocusModeURLParams) (*FocusModeURL, error) {
	row := q.db.QueryRowContext(ctx, createFocusModeURL,
		arg.ID, arg.ModeID, arg.URL,
	)
	var u FocusModeURL
	err := row.Scan(
		&u.ID, &u.ModeID, &u.URL, &u.CreatedAt,
	)
	return &u, err
}

func (q *Queries) GetFocusModeURLsByModeID(ctx context.Context, modeID string) ([]FocusModeURL, error) {
	rows, err := q.db.QueryContext(ctx, getFocusModeURLsByModeID, modeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []FocusModeURL
	for rows.Next() {
		var u FocusModeURL
		err := rows.Scan(
			&u.ID, &u.ModeID, &u.URL, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, rows.Err()
}

func (q *Queries) DeleteFocusModeURLsByModeID(ctx context.Context, modeID string) error {
	_, err := q.db.ExecContext(ctx, deleteFocusModeURLsByModeID, modeID)
	return err
}

func (q *Queries) CreateFocusSession(ctx context.Context, arg CreateFocusSessionParams) (*FocusModeSession, error) {
	row := q.db.QueryRowContext(ctx, createFocusSession,
		arg.ID, arg.ModeID, arg.EndsAt, arg.Status,
	)
	var s FocusModeSession
	err := row.Scan(
		&s.ID, &s.ModeID, &s.StartedAt, &s.EndsAt, &s.FinishedAt, &s.Status, &s.CreatedAt,
	)
	return &s, err
}

func (q *Queries) GetFocusSessionByID(ctx context.Context, id string) (*FocusModeSession, error) {
	row := q.db.QueryRowContext(ctx, getFocusSessionByID, id)
	var s FocusModeSession
	err := row.Scan(
		&s.ID, &s.ModeID, &s.StartedAt, &s.EndsAt, &s.FinishedAt, &s.Status, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (q *Queries) GetActiveFocusSession(ctx context.Context) (*FocusModeSession, error) {
	row := q.db.QueryRowContext(ctx, getActiveFocusSession)
	var s FocusModeSession
	err := row.Scan(
		&s.ID, &s.ModeID, &s.StartedAt, &s.EndsAt, &s.FinishedAt, &s.Status, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (q *Queries) GetLatestSessionByModeID(ctx context.Context, modeID string) (*FocusModeSession, error) {
	row := q.db.QueryRowContext(ctx, getLatestSessionByModeID, modeID)
	var s FocusModeSession
	err := row.Scan(
		&s.ID, &s.ModeID, &s.StartedAt, &s.EndsAt, &s.FinishedAt, &s.Status, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (q *Queries) UpdateFocusSessionStatus(ctx context.Context, id, status string) (*FocusModeSession, error) {
	row := q.db.QueryRowContext(ctx, updateFocusSessionStatus, id, status)
	var s FocusModeSession
	err := row.Scan(
		&s.ID, &s.ModeID, &s.StartedAt, &s.EndsAt, &s.FinishedAt, &s.Status, &s.CreatedAt,
	)
	return &s, err
}
