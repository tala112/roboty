package modes

import "time"

type FocusMode struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	DurationMinutes   int            `json:"duration_minutes"`
	MuteNotifications bool           `json:"mute_notifications"`
	Enabled           bool           `json:"enabled"`
	Icon              string         `json:"icon"`
	Color             string         `json:"color"`
	Apps              []FocusModeApp `json:"apps,omitempty"`
	AllowedURLs      []string       `json:"allowed_urls,omitempty"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

type FocusModeApp struct {
	ID               string `json:"id"`
	ModeID           string `json:"mode_id"`
	AppName          string `json:"app_name"`
	AppExec          string `json:"app_exec"`
	CloseOnActivate  bool   `json:"close_on_activate"`
	IsAllowed        bool   `json:"is_allowed"`
}

type FocusSession struct {
	ID         string  `json:"id"`
	ModeID     string  `json:"mode_id"`
	StartedAt  string  `json:"started_at"`
	EndsAt    *string `json:"ends_at,omitempty"`
	FinishedAt *string `json:"finished_at,omitempty"`
	Status    string  `json:"status"`
}

type InstalledApp struct {
	Name string `json:"name"`
	Exec string `json:"exec"`
	Icon string `json:"icon,omitempty"`
}

type CreateModeRequest struct {
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	DurationMinutes  int            `json:"duration_minutes"`
	MuteNotifications bool          `json:"mute_notifications"`
	Enabled          bool           `json:"enabled"`
	Icon             string         `json:"icon"`
	Color            string         `json:"color"`
	Apps             []FocusModeApp `json:"apps"`
	AllowedURLs     []string       `json:"allowed_urls"`
}

type UpdateModeRequest struct {
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	DurationMinutes  int            `json:"duration_minutes"`
	MuteNotifications bool          `json:"mute_notifications"`
	Enabled          bool           `json:"enabled"`
	Icon             string         `json:"icon"`
	Color            string         `json:"color"`
	Apps             []FocusModeApp `json:"apps"`
	AllowedURLs     []string       `json:"allowed_urls"`
}

func nowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
