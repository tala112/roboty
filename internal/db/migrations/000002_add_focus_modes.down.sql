-- +migrate Down
DROP TRIGGER IF EXISTS update_focus_modes_updated_at;
DROP INDEX IF EXISTS idx_focus_mode_sessions_status;
DROP INDEX IF EXISTS idx_focus_mode_sessions_mode_id;
DROP INDEX IF EXISTS idx_focus_mode_apps_mode_id;
DROP TABLE IF EXISTS focus_mode_sessions;
DROP TABLE IF EXISTS focus_mode_apps;
DROP TABLE IF EXISTS focus_modes;
