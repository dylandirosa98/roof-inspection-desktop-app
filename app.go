package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"roof-inspection-desktop-app/internal/analysis"
	"roof-inspection-desktop-app/internal/database"
	"roof-inspection-desktop-app/internal/inspection"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

func (a *App) CreateProject(path string, name string) (database.Project, error) {
	project, err := inspection.CreateProject(path, name, a.queries, a.ctx)
	if err != nil {
		fmt.Printf("Failed to get project from %s: %s\n", path, err)
		return project, err
	}
	return project, nil
}

func (a *App) RetrieveProject(id int64) ([]database.RetrieveImagesRow, error) {
	images, err := a.queries.RetrieveImages(a.ctx, id)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (a *App) GetProjects() []database.RetrieveProjectsRow {
	projects, err := a.queries.RetrieveProjects(a.ctx)
	if err != nil {
		return nil
	}
	return projects
}

func (a *App) PickDirectory() (string, error) {
	selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return "", err
	}
	return selected, nil
}

func (a *App) AnalyzeImage(image database.Image) (analysis.AnalysisResult, error) {
	analysisResult, err := analysis.AnalyzeImage(image)
	if err != nil {
		return analysis.AnalysisResult{}, err
	}
	annotationsJSON, err := json.Marshal(analysisResult)
	if err != nil {
		return analysis.AnalysisResult{}, err
	}
	aiImage := database.CreateAiImageParams{
		ImageID: image.ID,
		AnnotationsJson: sql.NullString{
			String: string(annotationsJSON),
			Valid:  true,
		},
	}
	_, err = a.queries.CreateAiImage(a.ctx, aiImage)
	if err != nil {
		return analysis.AnalysisResult{}, err
	}
	return analysisResult, nil
}
