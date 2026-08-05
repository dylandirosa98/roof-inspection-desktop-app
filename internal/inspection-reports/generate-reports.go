package inspection_reports

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"roof-inspection-desktop-app/internal/analysis"
	"roof-inspection-desktop-app/internal/database"

	"github.com/disintegration/imaging"
	"github.com/phpdave11/gofpdf"
)

const (
	reportMargin       = 10.0
	reportPageWidth    = 215.9
	reportContentWidth = reportPageWidth - reportMargin*2
	reportCropSize     = 1600
)

type Result struct {
	Path       string
	PageCount  int
	ImageCount int
}

type editedAnnotations struct {
	Detections []analysis.ImageDetection `json:"detections"`
}

type reportImage struct {
	Path string
}

func Generate(report database.InspectionReport, rows []database.RetrieveReviewedReportImagesRow, outputPath string, logo []byte) (Result, error) {
	if outputPath == "" {
		return Result{}, errors.New("output path is required")
	}
	if len(logo) == 0 {
		return Result{}, errors.New("report logo is unavailable")
	}

	tempDir, err := os.MkdirTemp("", "hailscan-report-")
	if err != nil {
		return Result{}, fmt.Errorf("create report image directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	images := make([]reportImage, 0, len(rows))
	for _, row := range rows {
		annotations, err := parseEditedAnnotations(row.EditedAnnotationsJson.String)
		if err != nil {
			return Result{}, fmt.Errorf("read annotations for image %d: %w", row.ImageID, err)
		}
		if len(annotations.Detections) == 0 {
			continue
		}

		annotatedPath := filepath.Join(tempDir, fmt.Sprintf("photo-%d.jpg", len(images)+1))
		if err := createAnnotatedPreview(row.ImagePath, annotations.Detections, annotatedPath); err != nil {
			return Result{}, fmt.Errorf("prepare image %d: %w", row.ImageID, err)
		}
		images = append(images, reportImage{Path: annotatedPath})
	}
	if len(images) == 0 {
		return Result{}, errors.New("no reviewed damage photos are available for this report")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return Result{}, fmt.Errorf("create report output directory: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(reportMargin, reportMargin, reportMargin)
	pdf.SetAutoPageBreak(false, reportMargin)
	pdf.RegisterImageOptionsReader("spartan-report-logo", gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logo))

	renderSummaryPage(pdf, report)
	renderPhotoPages(pdf, report, images)

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return Result{}, err
	}

	return Result{Path: outputPath, PageCount: pdf.PageCount(), ImageCount: len(images)}, nil
}

func parseEditedAnnotations(raw string) (editedAnnotations, error) {
	var annotations editedAnnotations
	if err := json.Unmarshal([]byte(raw), &annotations); err != nil {
		return editedAnnotations{}, err
	}
	return annotations, nil
}

func createAnnotatedPreview(sourcePath string, detections []analysis.ImageDetection, outputPath string) error {
	source, err := imaging.Open(sourcePath)
	if err != nil {
		return err
	}

	bounds := source.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	if width == 0 || height == 0 {
		return errors.New("image has no dimensions")
	}

	preview := imaging.Fill(source, reportCropSize, reportCropSize, imaging.Center, imaging.Lanczos)
	canvas := image.NewRGBA(preview.Bounds())
	draw.Draw(canvas, canvas.Bounds(), preview, preview.Bounds().Min, draw.Src)

	scale := math.Max(float64(reportCropSize)/width, float64(reportCropSize)/height)
	offsetX := (float64(reportCropSize) - width*scale) / 2
	offsetY := (float64(reportCropSize) - height*scale) / 2
	for _, detection := range detections {
		box := detection.BoundingBox
		drawBox(
			canvas,
			box.Left*float32(scale)+float32(offsetX),
			box.Top*float32(scale)+float32(offsetY),
			box.Right*float32(scale)+float32(offsetX),
			box.Bottom*float32(scale)+float32(offsetY),
		)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return jpeg.Encode(file, canvas, &jpeg.Options{Quality: 92})
}

func drawBox(canvas *image.RGBA, left, top, right, bottom float32) {
	minX := clamp(int(math.Round(float64(left))), 0, reportCropSize-1)
	minY := clamp(int(math.Round(float64(top))), 0, reportCropSize-1)
	maxX := clamp(int(math.Round(float64(right))), 0, reportCropSize-1)
	maxY := clamp(int(math.Round(float64(bottom))), 0, reportCropSize-1)
	if maxX <= minX || maxY <= minY {
		return
	}

	red := image.NewUniform(color.RGBA{R: 200, G: 25, B: 34, A: 255})
	const lineWidth = 12
	draw.Draw(canvas, image.Rect(minX, minY, maxX, minY+lineWidth), red, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(minX, maxY-lineWidth, maxX, maxY), red, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(minX, minY, minX+lineWidth, maxY), red, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(maxX-lineWidth, minY, maxX, maxY), red, image.Point{}, draw.Src)
}

func clamp(value, minimum, maximum int) int {
	return min(max(value, minimum), maximum)
}

func renderSummaryPage(pdf *gofpdf.Fpdf, report database.InspectionReport) {
	pdf.AddPage()
	renderHeader(pdf, report, report.ReportTitle)

	metadata := [][2]string{
		{"Report Number", report.ReportNumber},
		{"Property", joinAddress(report.PropertyAddress, report.CityStateZip)},
		{"Inspection Date", report.InspectionDate},
		{"Date of Loss", report.DateOfLoss},
		{"Customer / Contact", report.CustomerName},
		{"Inspector", report.InspectorName},
		{"Insurance Carrier", report.InsuranceCarrier},
		{"Claim Number", report.ClaimNumber},
	}
	renderMetadata(pdf, metadata)
	renderSectionTitle(pdf, "Inspection Summary")
	renderTextBox(pdf, report.Summary, 31)
	renderSectionTitle(pdf, "Documented Findings and Review Notes")
	renderTextBox(pdf, report.Notes, 38)

	pdf.SetY(244)
	pdf.SetFillColor(253, 247, 247)
	pdf.SetDrawColor(225, 196, 198)
	pdf.Rect(reportMargin, pdf.GetY(), reportContentWidth, 12, "DF")
	pdf.SetXY(reportMargin+3, pdf.GetY()+3)
	pdf.SetTextColor(92, 62, 66)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(reportContentWidth-6, 3.7, "Prepared by Spartan Exteriors to document observed roof conditions and inspection photographs.", "", "L", false)
	renderFooter(pdf)
}

func renderPhotoPages(pdf *gofpdf.Fpdf, report database.InspectionReport, images []reportImage) {
	for index := 0; index < len(images); index += 4 {
		pdf.AddPage()
		renderHeader(pdf, report, "Photo Documentation")
		renderSectionTitle(pdf, "Selected Inspection Images")

		for photoIndex := 0; photoIndex < 4 && index+photoIndex < len(images); photoIndex++ {
			column := photoIndex % 2
			row := photoIndex / 2
			x := reportMargin + float64(column)*99
			y := 56 + float64(row)*96
			pdf.SetDrawColor(209, 216, 223)
			pdf.Rect(x, y, 93, 90, "D")
			pdf.ImageOptions(images[index+photoIndex].Path, x+5, y, 83, 83, false, gofpdf.ImageOptions{ImageType: "JPG"}, 0, "")
			pdf.SetXY(x+3, y+84)
			pdf.SetTextColor(32, 36, 42)
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(87, 5, fmt.Sprintf("Photo %d", index+photoIndex+1), "", 0, "L", false, 0, "")
		}
		renderFooter(pdf)
	}
}

func renderHeader(pdf *gofpdf.Fpdf, report database.InspectionReport, title string) {
	pdf.SetFillColor(200, 25, 34)
	pdf.Rect(reportMargin, reportMargin, reportContentWidth, 2, "F")
	pdf.ImageOptions("spartan-report-logo", reportMargin, 15, 30, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.SetTextColor(32, 36, 42)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetXY(65, 17)
	pdf.MultiCell(140, 7, title, "", "R", false)
	pdf.SetTextColor(89, 99, 110)
	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(65, 34)
	pdf.CellFormat(140, 4, fmt.Sprintf("%s | %s", report.ReportNumber, joinAddress(report.PropertyAddress, report.CityStateZip)), "", 0, "R", false, 0, "")
	pdf.SetDrawColor(207, 213, 219)
	pdf.Line(reportMargin, 46, reportPageWidth-reportMargin, 46)
}

func renderMetadata(pdf *gofpdf.Fpdf, entries [][2]string) {
	for index, entry := range entries {
		column := index % 2
		row := index / 2
		x := reportMargin + float64(column)*98
		y := 51 + float64(row)*14
		pdf.SetFillColor(244, 246, 248)
		pdf.Rect(x, y, 94, 11, "F")
		pdf.SetFillColor(200, 25, 34)
		pdf.Rect(x, y, 1.5, 11, "F")
		pdf.SetTextColor(104, 115, 126)
		pdf.SetFont("Arial", "B", 6.5)
		pdf.SetXY(x+3, y+2)
		pdf.CellFormat(88, 3, entry[0], "", 0, "L", false, 0, "")
		pdf.SetTextColor(31, 41, 51)
		pdf.SetFont("Arial", "B", 8.5)
		pdf.SetXY(x+3, y+5.5)
		pdf.CellFormat(88, 4, entry[1], "", 0, "L", false, 0, "")
	}
	pdf.SetY(111)
}

func renderSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetTextColor(32, 36, 42)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(reportContentWidth, 7, title, "", 1, "L", false, 0, "")
	pdf.SetDrawColor(200, 25, 34)
	pdf.SetLineWidth(0.7)
	pdf.Line(reportMargin, pdf.GetY(), reportPageWidth-reportMargin, pdf.GetY())
	pdf.SetLineWidth(0.2)
	pdf.Ln(3)
}

func renderTextBox(pdf *gofpdf.Fpdf, value string, height float64) {
	if value == "" {
		value = "No notes provided."
	}
	y := pdf.GetY()
	pdf.SetFillColor(251, 252, 253)
	pdf.SetDrawColor(216, 222, 228)
	pdf.Rect(reportMargin, y, reportContentWidth, height, "DF")
	pdf.SetXY(reportMargin+3, y+3)
	pdf.SetTextColor(31, 41, 51)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(reportContentWidth-6, 4.5, value, "", "L", false)
	pdf.SetY(y + height + 5)
}

func renderFooter(pdf *gofpdf.Fpdf) {
	pdf.SetDrawColor(207, 213, 219)
	pdf.Line(reportMargin, 266, reportPageWidth-reportMargin, 266)
	pdf.SetTextColor(104, 115, 126)
	pdf.SetFont("Arial", "", 7.5)
	pdf.SetXY(reportMargin, 269)
	pdf.CellFormat(120, 4, "Spartan Exteriors | Roof Inspection and Photo Documentation", "", 0, "L", false, 0, "")
	pdf.CellFormat(75, 4, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "R", false, 0, "")
}

func joinAddress(address, cityStateZIP string) string {
	if address == "" {
		return cityStateZIP
	}
	if cityStateZIP == "" {
		return address
	}
	return address + ", " + cityStateZIP
}
