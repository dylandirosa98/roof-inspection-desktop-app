package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"roof-inspection-desktop-app/internal/database"
	"roof-inspection-desktop-app/internal/inspection"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// App struct
type App struct {
	ctx     context.Context
	db      *sql.DB
	queries *database.Queries
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	db, err := sql.Open("sqlite3", "./data/app.db")
	if err != nil {
		log.Fatal(err)
	}
	a.db = db
	err = runMigrations(db)
	if err != nil {
		log.Fatal(err)
	}
	queries := database.New(db)
	a.queries = queries
}

// migrations function
func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Up(db, "sql/schema"); err != nil {
		return err
	}
	return nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) CreateProject(path string) (inspection.Project, error) {
	project, err := inspection.CreateProject(path)
	if err != nil {
		fmt.Printf("Failed to get project from %s: %s\n", path, err)
		return project, err
	}
	return project, nil
}
