package inspection

import (
	"image"

	"github.com/disintegration/imaging"
)

func mimeType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func makeSquarePreview(img image.Image) image.Image {
	const size = 250

	preview := imaging.Fill(
		img,
		size,
		size,
		imaging.Center,
		imaging.Lanczos,
	)

	return preview
}
