package inspection

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

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
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	info, err := file.Stat()
	if err != nil {
		fmt.Printf("Error checking stats: %s\n", err)
		return Image{}, err
	}
	size := info.Size()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return Image{}, err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if encoded == "" {
		fmt.Printf("Error decoding file: %s\n", err)
	}
	dataURL := "data:" + mimeType(format) + ";base64" + encoded
	imageStruct := Image{
		Width:    width,
		Height:   height,
		Format:   format,
		FileSize: size,
		Path:     path,
		DataURL:  dataURL,
	}
	return imageStruct, nil
}
