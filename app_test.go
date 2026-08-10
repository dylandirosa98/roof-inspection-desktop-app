package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"roof-inspection-desktop-app/internal/database"
)

func TestDeleteProjectCascadesWithoutDeletingSourcePhotos(t *testing.T) {
	directory := t.TempDir()
	photoPath := filepath.Join(directory, "roof.jpg")
	if err := os.WriteFile(photoPath, []byte("test photo"), 0o600); err != nil {
		t.Fatal(err)
	}

	db, err := openDatabase(filepath.Join(directory, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	queries := database.New(db)
	project, err := queries.CreateProject(ctx, database.CreateProjectParams{Name: "demo", Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.ExecContext(ctx, "INSERT INTO images (path, project_id) VALUES (?, ?)", photoPath, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	imageID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO ai_images (image_id, annotations_json) VALUES (?, ?)", imageID, "{}"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO inspection_reports (project_id, report_number) VALUES (?, ?)", project.ID, "TEST-1"); err != nil {
		t.Fatal(err)
	}

	if err := queries.DeleteProject(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"projects", "images", "ai_images", "inspection_reports"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("%s count = %d, want 0", table, count)
		}
	}
	if _, err := os.Stat(photoPath); err != nil {
		t.Fatalf("source photo was removed: %v", err)
	}

	projects := (&App{ctx: ctx, queries: queries}).GetProjects()
	if projects == nil {
		t.Fatal("GetProjects returned nil after deleting the final project")
	}
	if len(projects) != 0 {
		t.Fatalf("GetProjects returned %d projects, want 0", len(projects))
	}
}
