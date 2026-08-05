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
	"strings"

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
	Notes      string                    `json:"notes"`
}

type reportImage struct {
	Path  string
	Notes string
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
		images = append(images, reportImage{Path: annotatedPath, Notes: annotations.Notes})
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

	firstPageImageCount := renderSummaryPage(pdf, report, images)
	renderPhotoPages(pdf, report, images[firstPageImageCount:], firstPageImageCount)

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

func renderSummaryPage(pdf *gofpdf.Fpdf, report database.InspectionReport, images []reportImage) int {
	pdf.AddPage()
	renderHeader(pdf, report, report.ReportTitle)

	metadata := [][2]string{
		{"Report Number", report.ReportNumber},
		{"Property", formatPropertyAddress(report)},
		{"Inspection Date", report.InspectionDate},
		{"Date of Loss", report.DateOfLoss},
		{"Customer / Contact", report.CustomerName},
		{"Inspector", report.InspectorName},
		{"Insurance Carrier", report.InsuranceCarrier},
		{"Claim Number", report.ClaimNumber},
	}
	renderMetadata(pdf, metadata)
	renderSectionTitle(pdf, "Inspection Summary")
	summaryHeight := textBoxHeight(pdf, report.Summary, 18)
	renderTextBox(pdf, report.Summary, summaryHeight)

	firstPageImageCount := 0
	const firstPageCardHeight = 88.0
	photoStartY := pdf.GetY() + 6
	if len(images) > 0 && photoStartY+firstPageCardHeight <= 240 {
		pdf.SetXY(reportMargin, pdf.GetY())
		pdf.SetTextColor(32, 36, 42)
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(reportContentWidth, 5, "Approved Damage Photos", "", 0, "L", false, 0, "")
		for index := 0; index < min(2, len(images)); index++ {
			x := reportMargin + float64(index)*99
			renderPhotoCard(pdf, images[index], index+1, x, photoStartY, firstPageCardHeight, 66)
			firstPageImageCount++
		}
	}

	pdf.SetY(244)
	pdf.SetFillColor(253, 247, 247)
	pdf.SetDrawColor(225, 196, 198)
	pdf.Rect(reportMargin, pdf.GetY(), reportContentWidth, 12, "DF")
	pdf.SetXY(reportMargin+3, pdf.GetY()+3)
	pdf.SetTextColor(92, 62, 66)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(reportContentWidth-6, 3.7, "Prepared by Spartan Exteriors to document observed roof conditions and inspection photographs.", "", "L", false)
	renderFooter(pdf)
	return firstPageImageCount
}

func renderPhotoPages(pdf *gofpdf.Fpdf, report database.InspectionReport, images []reportImage, imageNumberOffset int) {
	for index := 0; index < len(images); index += 4 {
		pdf.AddPage()
		renderHeader(pdf, report, "Photo Documentation")
		renderSectionTitle(pdf, "Selected Inspection Images")

		for photoIndex := 0; photoIndex < 4 && index+photoIndex < len(images); photoIndex++ {
			column := photoIndex % 2
			row := photoIndex / 2
			x := reportMargin + float64(column)*99
			y := 60 + float64(row)*101
			renderPhotoCard(pdf, images[index+photoIndex], imageNumberOffset+index+photoIndex+1, x, y, 98, 76)
		}
		renderFooter(pdf)
	}
}

func renderPhotoCard(pdf *gofpdf.Fpdf, image reportImage, number int, x, y, height, imageSize float64) {
	pdf.SetDrawColor(209, 216, 223)
	pdf.Rect(x, y, 93, height, "D")
	pdf.ImageOptions(image.Path, x+(93-imageSize)/2, y, imageSize, imageSize, false, gofpdf.ImageOptions{ImageType: "JPG"}, 0, "")
	pdf.SetXY(x+3, y+imageSize+2)
	pdf.SetTextColor(32, 36, 42)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(87, 5, fmt.Sprintf("Photo %d", number), "", 0, "L", false, 0, "")
	if note := photoNote(image.Notes); note != "" {
		pdf.SetXY(x+3, y+imageSize+7)
		pdf.SetTextColor(89, 99, 110)
		pdf.SetFont("Arial", "", 7.5)
		pdf.MultiCell(87, 3.2, note, "", "L", false)
	}
}

func renderHeader(pdf *gofpdf.Fpdf, report database.InspectionReport, title string) {
	const titleBlockX = 80.0
	const titleBlockWidth = 110.0

	pdf.SetFillColor(200, 25, 34)
	pdf.Rect(reportMargin, reportMargin, reportContentWidth, 2, "F")
	pdf.ImageOptions("spartan-report-logo", reportMargin, 15, 30, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.SetTextColor(32, 36, 42)
	pdf.SetFont("Arial", "B", 18)
	if title == "Roof Inspection and Photo Documentation Report" {
		title = "Roof Inspection and\nPhoto Documentation Report"
	}
	pdf.SetXY(titleBlockX, 16)
	pdf.MultiCell(titleBlockWidth, 8, title, "", "C", false)
	pdf.SetTextColor(89, 99, 110)
	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(65, 35)
	pdf.CellFormat(140, 4, fmt.Sprintf("%s | %s", report.ReportNumber, formatPropertyAddress(report)), "", 0, "R", false, 0, "")
	pdf.SetDrawColor(207, 213, 219)
	pdf.Line(reportMargin, 46, reportPageWidth-reportMargin, 46)
	pdf.SetXY(reportMargin, 49)
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
	pdf.SetX(reportMargin)
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

func textBoxHeight(pdf *gofpdf.Fpdf, value string, minimum float64) float64 {
	if value == "" {
		value = "No summary provided."
	}
	pdf.SetFont("Arial", "", 9)
	lineCount := len(pdf.SplitText(value, reportContentWidth-6))
	return max(minimum, float64(lineCount)*4.5+6)
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

func formatPropertyAddress(report database.InspectionReport) string {
	locality := strings.TrimSpace(strings.Join(nonEmptyStrings(report.PropertyCity, report.PropertyState), " "))
	if report.PropertyZip != "" {
		locality = strings.TrimSpace(locality + " " + report.PropertyZip)
	}
	if locality == "" {
		locality = report.CityStateZip
	}
	if report.PropertyAddress == "" {
		return locality
	}
	if locality == "" {
		return report.PropertyAddress
	}
	return report.PropertyAddress + ", " + locality
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func photoNote(note string) string {
	note = strings.TrimSpace(note)
	runes := []rune(note)
	if len(runes) <= 180 {
		return note
	}
	return string(runes[:177]) + "..."
}
