package analysis

import (
	"fmt"
	"roof-inspection-desktop-app/internal/database"
	"sort"

	ort "github.com/yalue/onnxruntime_go"
)

func RunModel(inputData []float32) ([]float32, error) {
	if len(inputData) != 3*640*640 {
		return nil, fmt.Errorf("Input data length must be 3*640*640")
	}
	ort.SetSharedLibraryPath("runtime/linux-amd64/libonnxruntime.so")
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, err
	}
	defer ort.DestroyEnvironment()

	inputTensor, err := ort.NewTensor(
		ort.NewShape(1, 3, 640, 640),
		inputData,
	)
	if err != nil {
		return nil, err
	}
	defer inputTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](
		ort.NewShape(1, 5, 8400),
	)
	if err != nil {
		return nil, err
	}
	defer outputTensor.Destroy()

	session, err := ort.NewAdvancedSession(
		"models/roof-hail-v1.onnx",
		[]string{"images"},
		[]string{"output0"},
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer session.Destroy()

	if err := session.Run(); err != nil {
		return nil, err
	}

	result := append([]float32(nil), outputTensor.GetData()...)
	return result, nil
}

type Detection struct {
	Class      string
	Confidence float32
	X          float32
	Y          float32
	Width      float32
	Height     float32
}

func DetectionToBox(detection Detection) (left, top, right, bottom float32) {
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

func ModelBoxToOriginal(detection Detection, preprocessed PreprocessedImage) (originalLeft, originalTop, originalRight, originalBottom float32) {
	left, top, right, bottom := DetectionToBox(detection)
	originalLeft = (left - preprocessed.XPadding) / preprocessed.Scale
	originalTop = (top - preprocessed.YPadding) / preprocessed.Scale
	originalRight = (right - preprocessed.XPadding) / preprocessed.Scale
	originalBottom = (bottom - preprocessed.YPadding) / preprocessed.Scale
	return originalLeft, originalTop, originalRight, originalBottom
}

func NonMaxSuppression(detections []Detection, iouThreshold float32) []Detection {
	sortedDetections := append([]Detection(nil), detections...)
	sort.Slice(sortedDetections, func(left, right int) bool {
		return sortedDetections[left].Confidence > sortedDetections[right].Confidence
	})

	iou := func(first, second Detection) float32 {
		firstLeft, firstTop, firstRight, firstBottom := DetectionToBox(first)
		secondLeft, secondTop, secondRight, secondBottom := DetectionToBox(second)

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

func Postprocess(output []float32, threshold float32) ([]Detection, error) {
	if len(output) != 42000 {
		return nil, fmt.Errorf("Output length must be 4200")
	}
	detections := make([]Detection, 0)
	for i := 0; i < 8400; i++ {
		x := output[i]
		y := output[i+8400]
		width := output[i+16800]
		height := output[i+25200]
		confidence := output[i+33600]

		if confidence > threshold {
			detections = append(detections, Detection{
				"hail-damage", confidence, x, y, width, height})
		}
	}
	return detections, nil
}

func AnalyzeImage(image database.Image) (ImageDetections, error) {
	resizedImage, err := ResizeImage(image)
	if err != nil {
		return ImageDetections{}, err
	}
	pixels := ImageToPixel(resizedImage.Resized)
	output, err := RunModel(pixels)
	if err != nil {
		return ImageDetections{}, err
	}
	detections, err := Postprocess(output, .25)
	if err != nil {
		return ImageDetections{}, err
	}
	detections = NonMaxSuppression(detections, .45)
	originalImageDetections := make([]Coordinates, len(detections))
	resizedImageDetections := make([]Coordinates, len(detections))
	for i, detection := range detections {
		left, top, right, bottom := ModelBoxToOriginal(detection, resizedImage)
		originalImageDetections[i] = Coordinates{
			top:    top,
			left:   left,
			right:  right,
			bottom: bottom,
		}
		rleft, rtop, rright, rbottom := DetectionToBox(detection)
		resizedImageDetections[i] = Coordinates{
			top:    rtop,
			left:   rleft,
			right:  rright,
			bottom: rbottom,
		}
	}
	result := ImageDetections{
		originalImageDetections: originalImageDetections,
		resizedImageDetections:  resizedImageDetections,
	}
	return result, nil
}

type Coordinates struct {
	top    float32
	left   float32
	right  float32
	bottom float32
}

type ImageDetections struct {
	originalImageDetections []Coordinates
	resizedImageDetections  []Coordinates
}
