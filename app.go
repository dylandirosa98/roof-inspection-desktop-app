package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"roof-inspection-desktop-app/internal/analysis"
	"roof-inspection-desktop-app/internal/database"
	"roof-inspection-desktop-app/internal/inspection"
	inspectionreports "roof-inspection-desktop-app/internal/inspection-reports"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/disintegration/imaging"
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
	databasePath, err := appDatabasePath()
	if err != nil {
		log.Fatal(err)
	}
	db, err := openDatabase(databasePath)
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

func openDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// migrations function
func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations)
	defer goose.SetBaseFS(nil)
	if err := goose.Up(db, "sql/schema"); err != nil {
		return err
	}
	return nil
}

func appDatabasePath() (string, error) {
	if goruntime.GOOS != "windows" {
		if err := os.MkdirAll("data", 0o755); err != nil {
			return "", err
		}
		return filepath.Join("data", "app.db"), nil
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", errors.New("LOCALAPPDATA is unavailable")
	}
	directory := filepath.Join(localAppData, "Spartan Roof Inspection")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(directory, "app.db"), nil
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
	if err != nil || projects == nil {
		return []database.RetrieveProjectsRow{}
	}
	return projects
}

func (a *App) DeleteProject(id int64) error {
	if id <= 0 {
		return errors.New("invalid project ID")
	}
	return a.queries.DeleteProject(a.ctx, id)
}

func (a *App) CreateInspectionReport(projectID int64, reportNumber string) (database.InspectionReport, error) {
	if strings.TrimSpace(reportNumber) == "" {
		return database.InspectionReport{}, errors.New("report number is required")
	}

	return a.queries.CreateInspectionReport(a.ctx, database.CreateInspectionReportParams{
		ProjectID:    projectID,
		ReportNumber: reportNumber,
	})
}

func (a *App) GetInspectionReportByProjectID(projectID int64) (database.InspectionReport, error) {
	return a.queries.GetInspectionReportByProjectID(a.ctx, projectID)
}

func (a *App) UpdateInspectionReport(report database.UpdateInspectionReportParams) (database.InspectionReport, error) {
	if strings.TrimSpace(report.ReportNumber) == "" {
		return database.InspectionReport{}, errors.New("report number is required")
	}

	return a.queries.UpdateInspectionReport(a.ctx, report)
}

func (a *App) GenerateInspectionReport(reportID int64, outputPath string) (inspectionreports.Result, error) {
	report, err := a.queries.GetInspectionReport(a.ctx, reportID)
	if err != nil {
		return inspectionreports.Result{}, err
	}

	images, err := a.queries.RetrieveReviewedReportImages(a.ctx, reportID)
	if err != nil {
		return inspectionreports.Result{}, err
	}

	result, err := inspectionreports.Generate(report, images, outputPath, reportLogo)
	if err != nil {
		return inspectionreports.Result{}, err
	}
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		return inspectionreports.Result{}, err
	}
	if err := a.queries.UpdateInspectionReportOutput(a.ctx, database.UpdateInspectionReportOutputParams{
		LastGeneratedPdfPath: outputPath,
		LastGeneratedAt:      time.Now().In(eastern).Format("2006-01-02 03:04 PM MST"),
		ID:                   reportID,
	}); err != nil {
		return inspectionreports.Result{}, err
	}

	return result, nil
}

func (a *App) OpenLastGeneratedInspectionReport(reportID int64) error {
	report, err := a.queries.GetInspectionReport(a.ctx, reportID)
	if err != nil {
		return err
	}
	if report.LastGeneratedPdfPath == "" {
		return errors.New("this report has not been generated yet")
	}
	if _, err := os.Stat(report.LastGeneratedPdfPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("the last generated PDF was moved or deleted")
		}
		return err
	}

	var command *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		command = exec.Command("open", report.LastGeneratedPdfPath)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", report.LastGeneratedPdfPath)
	default:
		command = exec.Command("xdg-open", report.LastGeneratedPdfPath)
	}
	return command.Start()
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

func (a *App) AnalyzeImage(imageRow database.RetrieveImagesRow) (analysis.AnalysisResult, error) {
	image := database.Image{
		ID:         imageRow.ID,
		Width:      imageRow.Width,
		Height:     imageRow.Height,
		FileSize:   imageRow.FileSize,
		Format:     imageRow.Format,
		Path:       imageRow.Path,
		DataUrl:    imageRow.DataUrl,
		PreviewUrl: imageRow.PreviewUrl,
		ProjectID:  imageRow.ProjectID,
	}
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

func (a *App) AnalyzeProject(project database.Project) error {
	images, err := a.queries.RetrieveImages(a.ctx, project.ID)
	if err != nil {
		return err
	}

	var analysisErrors []error
	for _, image := range images {
		_, err := a.AnalyzeImage(image)
		if err != nil {
			analysisErrors = append(analysisErrors, fmt.Errorf("image %d: %w", image.ID, err))
		}
	}
	return errors.Join(analysisErrors...)
}

func (a *App) RetrieveAiImages(projectID int64) ([]database.RetrieveAiImagesRow, error) {
	return a.queries.RetrieveAiImages(a.ctx, projectID)
}

func (a *App) SaveEditedAnnotations(imageID int64, annotationsJSON string) error {
	return a.queries.UpdateAiImageEditedAnnotations(a.ctx, database.UpdateAiImageEditedAnnotationsParams{
		ImageID:               imageID,
		EditedAnnotationsJson: sql.NullString{String: annotationsJSON, Valid: true},
	})
}

func (a *App) ApproveImageReview(imageID int64, annotationsJSON string) error {
	return a.queries.ApproveImageReview(a.ctx, database.ApproveImageReviewParams{
		ImageID:               imageID,
		EditedAnnotationsJson: sql.NullString{String: annotationsJSON, Valid: true},
	})
}

func (a *App) GetOriginalImageDataURL(imageID int64) (string, error) {
	image, err := a.queries.RetrieveImage(a.ctx, imageID)
	if err != nil {
		return "", err
	}

	original, err := imaging.Open(image.Path)
	if err != nil {
		return "", err
	}

	var encoded bytes.Buffer
	if err := imaging.Encode(&encoded, original, imaging.JPEG); err != nil {
		return "", err
	}

	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(encoded.Bytes())), nil
}
