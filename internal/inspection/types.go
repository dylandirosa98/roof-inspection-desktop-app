package inspection

import (
	"github.com/google/uuid"
)

type Image struct {
	Width    int
	Height   int
	Format   string
	FileSize int64
	Path     string
	DataURL  string
}

type ProjectImage struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Image     *Image
}

type Project struct {
	ID        uuid.UUID
	Directory string
	Images    []ProjectImage
}
