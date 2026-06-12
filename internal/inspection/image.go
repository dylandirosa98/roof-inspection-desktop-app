package inspection

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type Image struct {
	Width    int
	Height   int
	Format   string
	FileSize int64
	Path     string
}

func GetImage(path string) (Image, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return Image{}, err
	}
	defer file.Close()
	img, format, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding file: %s\n", err)
		return Image{}, err
	}
	switch format {
	case "jpeg":
		break
	case "png":
		break
	case "jpg":
		break
	default:
		err = errors.New("unsupported image format")
		return Image{}, err
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	info, err := file.Stat()
	if err != nil {
		fmt.Printf("Error checking stats: %s\n", err)
		return Image{}, err
	}
	size := info.Size()
	imageStruct := Image{
		Width:    width,
		Height:   height,
		Format:   format,
		FileSize: size,
		Path:     path,
	}
	return imageStruct, nil
}

type Project struct {
	Directory string
	Images    []Image
}

func GetProject(path string) (Project, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error opening directory: %s\n", err)
		return Project{}, err
	}
	project := Project{
		Directory: path,
		Images:    make([]Image, 0),
	}
	for _, entry := range entries {
		imageStruct, err := GetImage(filepath.Join(path, entry.Name()))
		if err != nil {
			continue
		}
		project.Images = append(project.Images, imageStruct)
	}
	return project, nil
}
