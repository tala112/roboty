package main

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	app := NewAppManager()

	err := wails.Run(&options.App{
		Title:  "Focus Mode",
		Width:  800,
		Height: 600,
		Bind: []interface{}{
			app,
		},
		OnStartup: app.startup,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}