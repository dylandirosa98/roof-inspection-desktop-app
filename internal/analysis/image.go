package analysis

import (
	"image"
	"image/color"
	"roof-inspection-desktop-app/internal/database"

	"github.com/disintegration/imaging"
)

type PreparedImage struct {
	Original image.Image
	Model    image.Image
	Scale    float32
	PaddingX float32
	PaddingY float32
}

func PrepareImage(imageRecord database.Image) (PreparedImage, error) {
	original, err := imaging.Open(imageRecord.Path)
	if err != nil {
		return PreparedImage{}, err
	}
	fitted := imaging.Fit(original, 640, 640, imaging.Lanczos)
	xOffset := (640 - fitted.Bounds().Dx()) / 2
	yOffset := (640 - fitted.Bounds().Dy()) / 2
	canvas := imaging.New(640, 640, color.NRGBA{A: 255})

	// The model always receives a 640x640 letterboxed image.
	return PreparedImage{
		Original: original,
		Model:    imaging.Paste(canvas, fitted, image.Point{X: xOffset, Y: yOffset}),
		Scale:    float32(fitted.Bounds().Dx()) / float32(original.Bounds().Dx()),
		PaddingX: float32(xOffset),
		PaddingY: float32(yOffset),
	}, nil
}

func imageToModelInput(image image.Image) []float32 {
	pixels := make([]float32, 3*640*640)
	for y := 0; y < 640; y++ {
		for x := 0; x < 640; x++ {
			r, g, b, _ := image.At(x, y).RGBA()
			pixelIndex := y*640 + x
			pixels[pixelIndex] = float32(r) / 65535
			pixels[pixelIndex+(640*640)] = float32(g) / 65535
			pixels[pixelIndex+(2*640*640)] = float32(b) / 65535
		}
	}
	return pixels
}
