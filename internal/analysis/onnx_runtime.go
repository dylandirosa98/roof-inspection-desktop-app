package analysis

import (
	"fmt"

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
