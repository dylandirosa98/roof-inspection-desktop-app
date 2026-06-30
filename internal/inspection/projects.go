package inspection

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func CreateProject(path string) (Project, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error opening directory: %s\n", err)
		return Project{}, err
	}
	projectID := uuid.New()
	project := Project{
		ID:        projectID,
		Directory: path,
		Images:    make([]ProjectImage, 0),
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		imageStruct, err := GetImage(filepath.Join(path, entry.Name()))
		if err != nil {
			continue
		}
		projectImage := ProjectImage{
			Image:     &imageStruct,
			ID:        uuid.New(),
			ProjectID: projectID,
		}
		project.Images = append(project.Images, projectImage)
	}
	return project, nil
}
