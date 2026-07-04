package inspection

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"
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
	previewWidth := 250
	previewHeight := 250
	dst := image.NewRGBA(image.Rect(0, 0, previewWidth, previewHeight))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 70})
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	previewurl := fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
	imageStruct := Image{
		Width:      width,
		Height:     height,
		Format:     format,
		FileSize:   size,
		Path:       path,
		PreviewURL: previewurl,
	}
	return imageStruct, nil
}

func GetImagePreview(path string) (Image, error) {
	imageStruct := Image{
		Path: path,
	}
	return imageStruct, nil
}
