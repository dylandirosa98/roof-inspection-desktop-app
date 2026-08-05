package inspection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"roof-inspection-desktop-app/internal/database"
)

func CreateProject(path string, name string, q *database.Queries, ctx context.Context) (database.Project, error) {
	if name == "" {
		return database.Project{}, fmt.Errorf("project name is required")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error opening directory: %s\n", err)
		return database.Project{}, err
	}
	images := make([]database.CreateImageParams, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		image, err := GetImage(filepath.Join(path, entry.Name()), 0)
		if err != nil {
			continue
		}
		images = append(images, image)
	}
	if len(images) == 0 {
		return database.Project{}, fmt.Errorf("no supported image files were found in the selected folder")
	}

	project, err := q.CreateProject(ctx, database.CreateProjectParams{
		Directory: path,
		Name:      name,
	})
	if err != nil {
		return database.Project{}, err
	}
	for _, image := range images {
		image.ProjectID = project.ID
		_, err = q.CreateImage(ctx, image)
		if err != nil {
			return database.Project{}, fmt.Errorf("save image: %w", err)
		}
	}
	return project, nil
}
