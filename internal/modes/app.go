package main

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
)

type App struct {
	Name string `json:"name"`
	Exec string `json:"exec"`
}

type AppManager struct {
	ctx context.Context
}

func NewAppManager() *AppManager {
	return &AppManager{}
}

func (a *AppManager) startup(ctx context.Context) {
	a.ctx = ctx
}

//
// 🔹 Get installed apps (.desktop)
//
func (a *AppManager) GetApps() ([]App, error) {
	paths := []string{
		"/usr/share/applications/",
		filepath.Join(os.Getenv("HOME"), ".local/share/applications/"),
	}

	var apps []App

	for _, path := range paths {
		files, _ := filepath.Glob(filepath.Join(path, "*.desktop"))

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				continue
			}
			defer f.Close()

			var name, execCmd string
			noDisplay := false

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()

				if strings.HasPrefix(line, "Name=") && name == "" {
					name = strings.TrimPrefix(line, "Name=")
				}

				if strings.HasPrefix(line, "Exec=") && execCmd == "" {
					execCmd = strings.TrimPrefix(line, "Exec=")
					execCmd = strings.Split(execCmd, " ")[0]
				}

				if strings.Contains(line, "NoDisplay=true") {
					noDisplay = true
				}
			}

			if name != "" && execCmd != "" && !noDisplay {
				apps = append(apps, App{
					Name: name,
					Exec: execCmd,
				})
			}
		}
	}

	return apps, nil
}

//
// 🔹 Block apps (STOP)
//
func (a *AppManager) BlockApps(apps []string) {
	for _, app := range apps {
		exec.Command("pkill", "-STOP", app).Run()
	}
}

//
// 🔹 Unblock apps (CONT)
//
func (a *AppManager) UnblockApps(apps []string) {
	for _, app := range apps {
		exec.Command("pkill", "-CONT", app).Run()
	}
}