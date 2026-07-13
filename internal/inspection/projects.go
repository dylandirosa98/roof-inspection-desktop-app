package inspection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"roof-inspection-desktop-app/internal/database"
)

func CreateProject(path string, name string, q *database.Queries, ctx context.Context) (database.Project, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error opening directory: %s\n", err)
		return database.Project{}, err
	}
	projectParams := database.CreateProjectParams{
		Directory: path,
		Name:      name,
	}
	project, err := q.CreateProject(ctx, projectParams)
	if err != nil {
		return database.Project{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		image, err := GetImage(filepath.Join(path, entry.Name()), project.ID)
		if err != nil {
			continue
		}
		_, err = q.CreateImage(ctx, image)
		if err != nil {
			fmt.Printf("Error creating image: %s\n", err)
		}
	}
	return project, nil
}
