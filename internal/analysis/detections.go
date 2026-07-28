package analysis

import (
	"fmt"
	"sort"
)

type Detection struct {
	Class      string
	Confidence float32
	X          float32
	Y          float32
	Width      float32
	Height     float32
}

type BoundingBox struct {
	Top    float32
	Left   float32
	Right  float32
	Bottom float32
}

type ImageDetection struct {
	BoundingBox BoundingBox
	Detection   Detection
}
type AnalysisResult struct {
	// OriginalImageBoxes are suitable for storing and rendering on the source image.
	OriginalImageBoxes []ImageDetection
	ModelImageBoxes    []ImageDetection
}

func outputToDetections(output []float32, threshold float32) ([]Detection, error) {
	if len(output) != 42000 {
		return nil, fmt.Errorf("Output length must be 4200")
	}

	detections := make([]Detection, 0)
	for candidateIndex := 0; candidateIndex < 8400; candidateIndex++ {
		x := output[candidateIndex]
		y := output[candidateIndex+8400]
		width := output[candidateIndex+16800]
		height := output[candidateIndex+25200]
		confidence := output[candidateIndex+33600]

		if confidence > threshold {
			detections = append(detections, Detection{
				"hail-damage", confidence, x, y, width, height})
		}
	}
	return detections, nil
}

func detectionBounds(detection Detection) (left, top, right, bottom float32) {
	left = detection.X - detection.Width/2
	top = detection.Y - detection.Height/2
	right = detection.X + detection.Width/2
	bottom = detection.Y + detection.Height/2

	left = max(0, left)
	top = max(0, top)
	right = min(640, right)
	bottom = min(640, bottom)
	return left, top, right, bottom
}

func modelBoundsToOriginal(detection Detection, prepared PreparedImage) (left, top, right, bottom float32) {
	left, top, right, bottom = detectionBounds(detection)
	left = (left - prepared.PaddingX) / prepared.Scale
	top = (top - prepared.PaddingY) / prepared.Scale
	right = (right - prepared.PaddingX) / prepared.Scale
	bottom = (bottom - prepared.PaddingY) / prepared.Scale
	return left, top, right, bottom
}

func suppressOverlappingDetections(detections []Detection, iouThreshold float32) []Detection {
	sortedDetections := append([]Detection(nil), detections...)
	sort.Slice(sortedDetections, func(left, right int) bool {
		return sortedDetections[left].Confidence > sortedDetections[right].Confidence
	})

	iou := func(first, second Detection) float32 {
		firstLeft, firstTop, firstRight, firstBottom := detectionBounds(first)
		secondLeft, secondTop, secondRight, secondBottom := detectionBounds(second)

		intersectionWidth := max(0, min(firstRight, secondRight)-max(firstLeft, secondLeft))
		intersectionHeight := max(0, min(firstBottom, secondBottom)-max(firstTop, secondTop))
		intersection := intersectionWidth * intersectionHeight
		firstArea := (firstRight - firstLeft) * (firstBottom - firstTop)
		secondArea := (secondRight - secondLeft) * (secondBottom - secondTop)
		return intersection / (firstArea + secondArea - intersection)
	}

	kept := make([]Detection, 0, len(sortedDetections))
	for len(sortedDetections) > 0 {
		best := sortedDetections[0]
		kept = append(kept, best)

		remaining := sortedDetections[:0]
		for _, candidate := range sortedDetections[1:] {
			if iou(best, candidate) <= iouThreshold {
				remaining = append(remaining, candidate)
			}
		}
		sortedDetections = remaining
	}

	return kept
}

func detectionsToResult(detections []Detection, prepared PreparedImage) AnalysisResult {
	originalImageBoxes := make([]ImageDetection, len(detections))
	modelImageBoxes := make([]ImageDetection, len(detections))

	for index, detection := range detections {
		left, top, right, bottom := modelBoundsToOriginal(detection, prepared)
		originalImageBoxes[index] = ImageDetection{
			BoundingBox: BoundingBox{Top: top, Left: left, Right: right, Bottom: bottom},
			Detection:   detection,
		}

		left, top, right, bottom = detectionBounds(detection)
		modelImageBoxes[index] = ImageDetection{
			BoundingBox: BoundingBox{Top: top, Left: left, Right: right, Bottom: bottom},
			Detection:   detection,
		}
	}

	return AnalysisResult{
		OriginalImageBoxes: originalImageBoxes,
		ModelImageBoxes:    modelImageBoxes,
	}
}
