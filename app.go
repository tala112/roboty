package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"Roboty/internal/db"
	"Roboty/internal/modes"
)

// CommandPreview represents a preview of what a command will do
type CommandPreview struct {
	Command     string `json:"command"`
	IsDangerous bool   `json:"is_dangerous"`
	Message    string `json:"message"`
}

// ChatInfo represents a chat for frontend
type ChatInfo struct {
	ID           string  `json:"id"`
	ParentChatID *string `json:"parent_chat_id,omitempty"`
	Title       string  `json:"title"`
	MessageCount int    `json:"message_count"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// MessageInfo represents a message for frontend
type MessageInfo struct {
	ID        string  `json:"id"`
	ChatID   string  `json:"chat_id"`
	Role     string  `json:"role"`
	Content  string  `json:"content"`
	Model    *string `json:"model,omitempty"`
	Provider *string `json:"provider,omitempty"`
	IsSummary bool   `json:"is_summary"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	FinishedAt *string `json:"finished_at,omitempty"`
}

// App struct
type App struct {
	ctx         context.Context
	database    *db.DB
	queries     *db.Queries
	modeService *modes.ModeService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// InitDatabase initializes the database connection
func (a *App) InitDatabase() error {
	log.Println("[INFO] Initializing database...")

	// Determine the best location for the database
	// Priority: 
	// 1. If roboty.db exists in CWD, use CWD (for development)
	// 2. If not, use directory next to executable
	var dbDir string
	
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	
	// Check if database already exists in CWD
	dbInCwd := filepath.Join(cwd, "roboty.db")
	if _, err := os.Stat(dbInCwd); err == nil {
		log.Printf("[DEBUG] Found existing database in CWD: %s", dbInCwd)
		dbDir = cwd
	} else {
		// Database not in CWD - use exe directory
		execPath, err := os.Executable()
		if err != nil {
			log.Printf("[WARN] Could not get executable path: %v", err)
			dbDir = cwd
		} else {
			dbDir = filepath.Dir(execPath)
		}
	}
	
	log.Printf("[DEBUG] Using database directory: %s", dbDir)
	
	database, err := db.NewDB(dbDir)
	if err != nil {
		log.Printf("[ERROR] InitDatabase: failed to open database: %v", err)
		return fmt.Errorf("failed to open database: %w", err)
	}
	a.database = database
	a.queries = db.NewQueries(database.DB())

	a.modeService = modes.NewModeService(a.database, a.queries)

	// Wire global emergency callback (used by safeGo panic recovery and signal handler)
	modes.SetGlobalEmergencyCallback(a.modeService.EmergencyStop)
	modes.SetupSignalHandler()

	if err := a.modeService.InitFocusSchema(); err != nil {
		log.Printf("[WARN] Failed to init focus schema: %v", err)
	}

	log.Println("[INFO] Database opened successfully")

	// Check if schema exists - count all tables
	var tableCount int
	err = database.DB().QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
	if err != nil {
		log.Printf("[ERROR] InitDatabase: error checking tables: %v", err)
		return fmt.Errorf("error checking tables: %w", err)
	}
	
	log.Printf("[DEBUG] Found %d tables in database", tableCount)
	
	if tableCount < 5 {
		// Need to create schema
		log.Println("[INFO] Schema doesn't exist, creating...")
		
		schemaPath := "internal/db/queries/schema_raw.sql"
		absPath, _ := filepath.Abs(schemaPath)
		log.Printf("[DEBUG] Loading schema from: %s", absPath)
		
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			log.Printf("[ERROR] InitDatabase: failed to read schema: %v", err)
			return fmt.Errorf("failed to read schema: %w", err)
		}
		
		log.Printf("[DEBUG] Schema file size: %d bytes", len(schema))
		
		// Execute schema in parts for better error handling
		schemaStr := string(schema)
		_, err = database.DB().Exec(schemaStr)
		if err != nil {
			log.Printf("[ERROR] InitDatabase: failed to create schema: %v", err)
			return fmt.Errorf("failed to create schema: %w", err)
		}
		
		// VERIFY tables were actually created
		err = database.DB().QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
		if err != nil {
			log.Printf("[FATAL] Schema execution returned no error but tables not found!")
			return fmt.Errorf("schema creation failed - cannot verify tables: %w", err)
		}
		
		log.Printf("[INFO] Schema created and verified: %d tables", tableCount)
	} else {
		log.Printf("[INFO] Schema already exists (%d tables)", tableCount)
	}

	// Create welcome message if no chats exist
	chats, err := a.queries.GetAllChats(context.Background())
	if err != nil || len(chats) == 0 {
		log.Println("[INFO] No chats found, creating welcome chat...")
		// Create first chat with welcome message
		chatID := fmt.Sprintf("chat-%d", time.Now().UnixMilli())
		now := time.Now().Format("2006-01-02 15:04:05")
		_, err = database.DB().Exec(`
			INSERT INTO chats (id, title, message_count, prompt_tokens, completion_tokens, cost, metadata, created_at, updated_at)
			VALUES (?, ?, 0, 0, 0, 0, '{}', ?, ?)
		`, chatID, "Welcome Chat", now, now)
		if err != nil {
			return fmt.Errorf("failed to create welcome chat: %w", err)
		}

		// Add welcome message
		msgID := fmt.Sprintf("msg-%d", time.Now().UnixMilli())
		welcomeContent := `{"text": "Hello! I'm Roboty, your AI assistant. I'm here to help you with anything you need."}`
		_, err = database.DB().Exec(`
			INSERT INTO messages (id, chat_id, role, content, model, provider, is_summary, created_at, updated_at)
			VALUES (?, ?, 'assistant', ?, NULL, NULL, 0, ?, ?)
		`, msgID, chatID, welcomeContent, now, now)
		if err != nil {
			return fmt.Errorf("failed to create welcome message: %w", err)
		}
		log.Println("[INFO] Welcome chat created successfully")
	}

	a.modeService.CheckResumeSessions()

	log.Println("[INFO] Database initialization complete")
	return nil
}

// CreateChat creates a new chat
func (a *App) CreateChat(title string) (ChatInfo, error) {
	if title == "" {
		title = fmt.Sprintf("Chat %s", time.Now().Format("2006-01-02 15:04"))
	}

	chatID := fmt.Sprintf("chat-%d", time.Now().UnixMilli())
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := a.database.DB().Exec(`
		INSERT INTO chats (id, title, message_count, prompt_tokens, completion_tokens, cost, metadata, created_at, updated_at)
		VALUES (?, ?, 0, 0, 0, 0, '{}', ?, ?)
	`, chatID, title, now, now)
	if err != nil {
		log.Printf("[ERROR] CreateChat: failed to insert chat: %v", err)
		return ChatInfo{}, fmt.Errorf("failed to create chat: %w", err)
	}

	log.Printf("[DEBUG] CreateChat: created chat ID=%s, title=%s", chatID, title)

	// Add welcome message
	msgID := fmt.Sprintf("msg-%d", time.Now().UnixMilli())
	welcomeContent := `{"text": "Hello! I'm Roboty, your AI assistant."}`
	_, err = a.database.DB().Exec(`
		INSERT INTO messages (id, chat_id, role, content, model, provider, is_summary, created_at, updated_at)
		VALUES (?, ?, 'assistant', ?, NULL, NULL, 0, ?, ?)
	`, msgID, chatID, welcomeContent, now, now)
	if err != nil {
		log.Printf("[ERROR] CreateChat: failed to insert welcome message: %v", err)
		return ChatInfo{}, fmt.Errorf("failed to create welcome message: %w", err)
	}

	log.Printf("[DEBUG] CreateChat: created welcome message ID=%s", msgID)

	return ChatInfo{
		ID:           chatID,
		Title:        title,
		MessageCount: 0,
		CreatedAt:    now,
		UpdatedAt:   now,
	}, nil
}

// GetChats returns all chats
func (a *App) GetChats() ([]ChatInfo, error) {
	chats, err := a.queries.GetAllChats(context.Background())
	if err != nil {
		log.Printf("[ERROR] GetChats: %v", err)
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	log.Printf("[DEBUG] GetChats: found %d chats", len(chats))

	result := make([]ChatInfo, len(chats))
	for i, c := range chats {
		result[i] = ChatInfo{
			ID:            c.ID,
			ParentChatID:  c.ParentChatID,
			Title:         c.Title,
			MessageCount:  c.MessageCount,
			CreatedAt:     c.CreatedAt.String(),
			UpdatedAt:    c.UpdatedAt.String(),
		}
	}
	return result, nil
}

// GetActiveChat returns the most recent chat
func (a *App) GetActiveChat() (*ChatInfo, error) {
	chats, err := a.queries.GetAllChats(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}
	if len(chats) == 0 {
		return nil, nil
	}

	c := chats[0]
	result := &ChatInfo{
		ID:            c.ID,
		ParentChatID:  c.ParentChatID,
		Title:         c.Title,
		MessageCount:  c.MessageCount,
		CreatedAt:     c.CreatedAt.String(),
		UpdatedAt:    c.UpdatedAt.String(),
	}
	return result, nil
}

// GetChatMessages returns messages for a chat
func (a *App) GetChatMessages(chatID string) ([]MessageInfo, error) {
	messages, err := a.queries.GetMessagesByChatID(context.Background(), chatID)
	if err != nil {
		log.Printf("[ERROR] GetChatMessages(%s): %v", chatID, err)
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	log.Printf("[DEBUG] GetChatMessages(%s): found %d messages", chatID, len(messages))

	result := make([]MessageInfo, len(messages))
	for i, m := range messages {
		result[i] = MessageInfo{
			ID:         m.ID,
			ChatID:    m.ChatID,
			Role:      string(m.Role),
			Content:   m.Content,
			Model:     m.Model,
			Provider:  m.Provider,
			IsSummary: m.IsSummary,
			CreatedAt: m.CreatedAt.String(),
			UpdatedAt: m.UpdatedAt.String(),
		}
	}
	return result, nil
}

// SaveMessage saves a message to the database
func (a *App) SaveMessage(chatID, role, content string) (MessageInfo, error) {
	if chatID == "" {
		// Get active chat
		chat, err := a.queries.GetAllChats(context.Background())
		if err != nil || len(chat) == 0 {
			log.Printf("[ERROR] SaveMessage: no active chat")
			return MessageInfo{}, fmt.Errorf("no active chat")
		}
		chatID = chat[0].ID
	}

	msgID := fmt.Sprintf("msg-%d", time.Now().UnixMilli())
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := a.database.DB().Exec(`
		INSERT INTO messages (id, chat_id, role, content, model, provider, is_summary, created_at, updated_at)
		VALUES (?, ?, ?, ?, NULL, NULL, 0, ?, ?)
	`, msgID, chatID, role, content, now, now)
	if err != nil {
		log.Printf("[ERROR] SaveMessage(%s, %s): %v", chatID, role, err)
		return MessageInfo{}, fmt.Errorf("failed to save message: %w", err)
	}

	log.Printf("[DEBUG] SaveMessage: created message ID=%s, chat=%s, role=%s", msgID, chatID, role)

	// Auto-update chat title if this is the first user message
	if role == "user" {
		var count int
		err = a.database.DB().QueryRow("SELECT message_count FROM chats WHERE id = ?", chatID).Scan(&count)
		if err == nil && count == 1 {
			// Extract text from JSON content
			title := content
			if strings.HasPrefix(content, "{\"") {
				// Try to parse JSON
				var raw map[string]interface{}
				if json.Unmarshal([]byte(content), &raw) == nil {
					if text, ok := raw["text"].(string); ok {
						title = text
					}
				}
			}
			// Set title to first 25 chars
			if len(title) > 25 {
				title = title[:25] + "..."
			}
			a.UpdateChatTitle(chatID, title)
		}
	}

	return MessageInfo{
		ID:        msgID,
		ChatID:    chatID,
		Role:      role,
		Content:  content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// DeleteChat deletes a chat
func (a *App) DeleteChat(chatID string) error {
	_, err := a.database.DB().Exec("DELETE FROM chats WHERE id = ?", chatID)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	return nil
}

// UpdateChatTitle updates a chat's title
func (a *App) UpdateChatTitle(chatID, title string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := a.database.DB().Exec(`
		UPDATE chats SET title = ?, updated_at = ? WHERE id = ?
	`, title, now, chatID)
	if err != nil {
		log.Printf("[ERROR] UpdateChatTitle(%s, %s): %v", chatID, title, err)
		return fmt.Errorf("failed to update chat title: %w", err)
	}
	log.Printf("[DEBUG] UpdateChatTitle: updated chat %s to title %s", chatID, title)
	return nil
}

// PreviewCommand analyzes a command without executing it
func (a *App) PreviewCommand(cmd string) CommandPreview {
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))

	dangerousPatterns := []string{
		"del /f /s /q",
		"rmdir /s /q",
		"rm -rf",
		"del /s /q",
		"format",
		"diskpart",
		"reg delete",
		"bcdedit",
		"shutdown",
		"taskkill /f",
		"netsh",
		"icacls",
		"takeown",
		"del ",
		"erase ",
		"rd ",
		"dir",
		"rmdir",
		"move ",
		"replace",
		"attrib -r",
		"cacls",
		"cipher",
	}

	isBlocked := false
	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdLower, pattern) {
			isBlocked = true
			break
		}
	}

	preview := CommandPreview{
		Command:    cmd,
		IsDangerous: isBlocked,
	}

	if isBlocked {
		preview.Message = "⛔ This command is BLOCKED and cannot be executed for security reasons.\n\nCommand: " + cmd
	} else {
		preview.Message = "⚠️ This command requires your confirmation to execute.\n\nCommand: " + cmd + "\n\nClick Confirm to execute or Cancel to abort."
	}

	return preview
}

// ExecuteCommand runs a command only if explicitly approved
func (a *App) ExecuteCommand(cmd string, approved bool) string {
	if !approved {
		return "Error: Command not approved. Please confirm execution first."
	}

	/*
	out, err := exec.Command("cmd", "/C", cmd).CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)*/



	out, err := exec.Command("wsl", "bash", "-lc", cmd).CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// RunCommand is kept for backward compatibility
func (a *App) RunCommand(cmd string) string {
	return a.ExecuteCommand(cmd, true)
}

// GetSessionInfo returns session/chat info as JSON
func (a *App) GetSessionInfo(chatID string) string {
	chat, err := a.queries.GetChatByID(context.Background(), chatID)
	if err != nil {
		log.Printf("[ERROR] GetSessionInfo(%s): %v", chatID, err)
		return "{}"
	}
	info := map[string]interface{}{
		"title":            chat.Title,
		"message_count":    chat.MessageCount,
		"prompt_tokens":    chat.PromptTokens,
		"completion_tokens": chat.CompletionTokens,
		"cost":             chat.Cost,
		"created_at":       chat.CreatedAt.String(),
		"updated_at":       chat.UpdatedAt.String(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		log.Printf("[ERROR] GetSessionInfo(%s): marshal error: %v", chatID, err)
		return "{}"
	}
	return string(data)
}

// GetSessionFiles returns file paths for a session as JSON array
func (a *App) GetSessionFiles(chatID string) string {
	files, err := a.queries.GetFilesByChatID(context.Background(), chatID)
	if err != nil {
		log.Printf("[ERROR] GetSessionFiles(%s): %v", chatID, err)
		return "[]"
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	data, err := json.Marshal(paths)
	if err != nil {
		log.Printf("[ERROR] GetSessionFiles(%s): marshal error: %v", chatID, err)
		return "[]"
	}
	return string(data)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.modeService != nil {
		a.modeService.SetContext(ctx)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ============================================================
// Focus Modes API
// ============================================================

// ListModes returns all focus modes
func (a *App) ListModes() (string, error) {
	modes, err := a.modeService.ListModes()
	if err != nil {
		return "[]", err
	}
	data, _ := json.Marshal(modes)
	return string(data), nil
}

// CreateMode creates a new focus mode (accepts appsJSON + urlsJSON)
func (a *App) CreateMode(name, description string, durationMinutes int, muteNotifications bool, icon, color string, appsJSON, urlsJSON string) (string, error) {
	var apps []modes.FocusModeApp
	if appsJSON != "" {
		if err := json.Unmarshal([]byte(appsJSON), &apps); err != nil {
			return "", fmt.Errorf("invalid apps JSON: %w", err)
		}
	}
	var urls []string
	if urlsJSON != "" {
		if err := json.Unmarshal([]byte(urlsJSON), &urls); err != nil {
			return "", fmt.Errorf("invalid urls JSON: %w", err)
		}
	}
	mode, err := a.modeService.CreateMode(modes.CreateModeRequest{
		Name:             name,
		Description:      description,
		DurationMinutes:  durationMinutes,
		MuteNotifications: muteNotifications,
		Icon:             icon,
		Color:            color,
		Apps:             apps,
		AllowedURLs:     urls,
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(mode)
	return string(data), nil
}

// UpdateMode updates an existing focus mode (accepts appsJSON + urlsJSON)
func (a *App) UpdateMode(id, name, description string, durationMinutes int, muteNotifications, enabled bool, icon, color string, appsJSON, urlsJSON string) (string, error) {
	var apps []modes.FocusModeApp
	if appsJSON != "" {
		if err := json.Unmarshal([]byte(appsJSON), &apps); err != nil {
			return "", fmt.Errorf("invalid apps JSON: %w", err)
		}
	}
	var urls []string
	if urlsJSON != "" {
		if err := json.Unmarshal([]byte(urlsJSON), &urls); err != nil {
			return "", fmt.Errorf("invalid urls JSON: %w", err)
		}
	}
	mode, err := a.modeService.UpdateMode(id, modes.UpdateModeRequest{
		Name:             name,
		Description:      description,
		DurationMinutes:  durationMinutes,
		MuteNotifications: muteNotifications,
		Enabled:          enabled,
		Icon:             icon,
		Color:            color,
		Apps:             apps,
		AllowedURLs:     urls,
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(mode)
	return string(data), nil
}

// DeleteMode deletes a focus mode
func (a *App) DeleteMode(id string) error {
	return a.modeService.DeleteMode(id)
}

// ToggleMode enables or disables a focus mode
func (a *App) ToggleMode(id string, enabled bool) error {
	return a.modeService.ToggleMode(id, enabled)
}

// GetInstalledApps returns all installed applications
func (a *App) GetInstalledApps() (string, error) {
	apps, err := a.modeService.GetInstalledApps()
	if err != nil {
		return "[]", err
	}
	data, _ := json.Marshal(apps)
	return string(data), nil
}

// ActivateMode activates a focus mode immediately
func (a *App) ActivateMode(modeID string) (string, error) {
	session, err := a.modeService.ActivateMode(modeID)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(session)
	return string(data), nil
}

// DeactivateMode deactivates an active session
func (a *App) DeactivateMode(sessionID string) error {
	return a.modeService.DeactivateMode(sessionID)
}

// GetActiveSession returns the currently active session if any
func (a *App) GetActiveSession() (string, error) {
	session, err := a.modeService.GetActiveSession()
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", nil
	}
	data, _ := json.Marshal(session)
	return string(data), nil
}

// GetAllDetectableApps returns apps from mappings + installed + running (deduplicated)
func (a *App) GetAllDetectableApps() (string, error) {
	apps, err := a.modeService.GetInstalledApps()
	if err != nil {
		return "[]", err
	}
	data, _ := json.Marshal(apps)
	return string(data), nil
}

// CheckAppOnPC checks if an app exec name exists on the user's PC (installed or running)
func (a *App) CheckAppOnPC(appExec string) bool {
	return a.modeService.CheckAppOnPC(appExec)
}

// AddAllowedApp adds an app to the allowed list for a mode and persists to mappings
// Only adds to mappings if the app is confirmed on PC or force=true
func (a *App) AddAllowedApp(modeID, appName, appExec, category string, force bool) (string, error) {
	if category == "" {
		category = "productive"
	}

	// Only persist to app-mappings if app is on PC or user forced it
	onPC := a.modeService.CheckAppOnPC(appExec)
	if onPC || force {
		if err := a.modeService.AddToAppMappings(appName, appExec, category); err != nil {
			log.Printf("[app] Failed to persist to app-mappings: %v", err)
		}
	}

	return `{"status":"ok","on_pc":` + fmt.Sprintf("%v", onPC) + `}`, nil
}

// GetURLBlockerStatus returns whether the URL blocker is running
func (a *App) GetURLBlockerStatus() string {
	if a.modeService == nil {
		return `{"running":false}`
	}
	running := a.modeService.GetURLBlockerRunning()
	return fmt.Sprintf(`{"running":%v}`, running)
}