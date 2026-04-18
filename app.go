package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandPreview represents a preview of what a command will do
type CommandPreview struct {
	Command     string
	IsDangerous bool
	Message     string
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
		Command:     cmd,
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

	out, err := exec.Command("cmd", "/C", cmd).CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// RunCommand is kept for backward compatibility
func (a *App) RunCommand(cmd string) string {
	return a.ExecuteCommand(cmd, true)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
