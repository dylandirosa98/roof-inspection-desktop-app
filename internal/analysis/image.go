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
	finalImage := imaging.New(640, 640, color.NRGBA{0, 0, 0, 255})
	point := image.Point{xOffset, yOffset}
	return imaging.Paste(finalImage, newImg, point), nil
}
