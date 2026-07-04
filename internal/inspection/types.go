package inspection

type Image struct {
	Width      int
	Height     int
	Format     string
	FileSize   int64
	Path       string
	DataURL    string
	PreviewURL string
}

type ProjectImage struct {
	ID        int
	ProjectID int
	Image     *Image
}

type Project struct {
	ID        int
	Directory string
	Name      string
	Images    []ProjectImage
}
