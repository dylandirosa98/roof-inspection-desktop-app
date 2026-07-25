package analysis

import (
	"image"
	"image/color"
	"roof-inspection-desktop-app/internal/database"

	"github.com/disintegration/imaging"
)

func ResizeImage(imag database.Image) (image.Image, error) {
	img, err := imaging.Open(imag.Path)
	if err != nil {
		return nil, err
	}
	newImg := imaging.Fit(img, 640, 640, imaging.Lanczos)
	xOffset := (640 - newImg.Bounds().Dx()) / 2
	yOffset := (640 - newImg.Bounds().Dy()) / 2
	finalImage := imaging.New(640, 640, color.NRGBA{A: 255})
	point := image.Point{X: xOffset, Y: yOffset}
	return imaging.Paste(finalImage, newImg, point), nil
}

func ImageToPixel(img image.Image) []float32 {
	finalList := make([]float32, 3*640*640)
	for i := 0; i < 640; i++ {
		for j := 0; j < 640; j++ {
			r, g, b, _ := img.At(j, i).RGBA()
			index := i*640 + j
			finalList[index] = float32(r) / 65535
			finalList[index+(640*640)] = float32(g) / 65535
			finalList[index+(2*640*640)] = float32(b) / 65535
		}
	}
	return finalList
}
