package analysis

import (
	"fmt"
	"roof-inspection-desktop-app/internal/database"

	ort "github.com/yalue/onnxruntime_go"
)

func runModel(inputData []float32) ([]float32, error) {
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

	return append([]float32(nil), outputTensor.GetData()...), nil
}

func AnalyzeImage(imageRecord database.Image) (AnalysisResult, error) {
	prepared, err := PrepareImage(imageRecord)
	if err != nil {
		return AnalysisResult{}, err
	}
	modelOutput, err := runModel(imageToModelInput(prepared.Model))
	if err != nil {
		return AnalysisResult{}, err
	}
	candidates, err := outputToDetections(modelOutput, .25)
	if err != nil {
		return AnalysisResult{}, err
	}
	detections := suppressOverlappingDetections(candidates, .45)
	return detectionsToResult(detections, prepared), nil
}
