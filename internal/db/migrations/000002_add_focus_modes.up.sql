-- +migrate Up
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
END;
