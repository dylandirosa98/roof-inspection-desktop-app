package inspection

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"roof-inspection-desktop-app/internal/database"
)

func GetImage(path string, projectID int64) (database.CreateImageParams, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return database.CreateImageParams{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	img, format, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding file: %s\n", err)
		return database.CreateImageParams{}, err
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	info, err := file.Stat()
	if err != nil {
		fmt.Printf("Error checking stats: %s\n", err)
		return database.CreateImageParams{}, err
	}
	size := info.Size()
	dst := makeSquarePreview(img)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 70})
	if err != nil {
		fmt.Printf("Error encoding image: %s\n", err)
		return database.CreateImageParams{}, err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	previewURL := fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
	imageStruct := database.CreateImageParams{
		Width:      sql.NullInt64{Int64: int64(width), Valid: true},
		Height:     sql.NullInt64{Int64: int64(height), Valid: true},
		Format:     sql.NullString{String: format, Valid: true},
		FileSize:   sql.NullInt64{Int64: size, Valid: true},
		Path:       path,
		PreviewUrl: sql.NullString{String: previewURL, Valid: true},
		DataUrl:    sql.NullString{Valid: false},
		ProjectID:  projectID,
	}
	return imageStruct, nil
}
