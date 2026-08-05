package inspection_reports

import (
	"database/sql"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"roof-inspection-desktop-app/internal/analysis"
	"roof-inspection-desktop-app/internal/database"
	"testing"
)

func TestGenerateCreatesLetterPDFForReviewedImage(t *testing.T) {
	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "source.jpg")
	outputPath := filepath.Join(tempDir, "report.pdf")

	source := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			source.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 120, A: 255})
		}
	}
	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(file, source, nil); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	annotations, err := json.Marshal(editedAnnotations{Detections: []analysis.ImageDetection{{
		BoundingBox: analysis.BoundingBox{Left: 120, Top: 40, Right: 280, Bottom: 160},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	logo, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "assets", "images", "spartan-exteriors-logo.png"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := Generate(
		database.InspectionReport{
			ID:              1,
			ProjectID:       1,
			ReportNumber:    "SAMPLE-001",
			ReportTitle:     "Roof Inspection and Photo Documentation Report",
			PropertyAddress: "123 Example Lane",
			CityStateZip:    "Exampletown, NJ 00000",
			InspectionDate:  "2026-08-04",
			CreatedAt:       "2026-08-04",
		},
		[]database.RetrieveReviewedReportImagesRow{{
			ImageID:               1,
			ImagePath:             imagePath,
			EditedAnnotationsJson: sql.NullString{String: string(annotations), Valid: true},
		}},
		outputPath,
		logo,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != outputPath || result.ImageCount != 1 || result.PageCount != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("generated PDF is empty")
	}
}
