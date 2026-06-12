package main

import (
	"context"
	"fmt"
	"roof-inspection-desktop-app/internal/inspection"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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

func (a *App) GetProject(path string) (inspection.Project, error) {
	project, err := inspection.GetProject(path)
	if err != nil {
		fmt.Printf("Failed to get project from %s: %s\n", path, err)
		return project, err
	}
	return project, nil
}
