package analysis

import "testing"

func TestOutputToDetectionsReadsCandidatePlanes(t *testing.T) {
	output := make([]float32, 42000)
	candidate := 7
	output[candidate] = 320
	output[8400+candidate] = 240
	output[16800+candidate] = 100
	output[25200+candidate] = 80
	output[33600+candidate] = 0.9
	output[33600+8] = 0.2

	detections, err := outputToDetections(output, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if len(detections) != 1 {
		t.Fatalf("detections = %d, want 1", len(detections))
	}

	detection := detections[0]
	if detection.Class != "hail-damage" || detection.Confidence != 0.9 || detection.X != 320 || detection.Y != 240 || detection.Width != 100 || detection.Height != 80 {
		t.Fatalf("unexpected detection: %#v", detection)
	}
}

func TestDetectionBoundsConvertsAndClamps(t *testing.T) {
	left, top, right, bottom := detectionBounds(Detection{X: 20, Y: 20, Width: 100, Height: 100})
	if left != 0 || top != 0 || right != 70 || bottom != 70 {
		t.Fatalf("box = (%v, %v, %v, %v), want (0, 0, 70, 70)", left, top, right, bottom)
	}
}

func TestModelBoundsToOriginalReversesLetterboxing(t *testing.T) {
	prepared := PreparedImage{Scale: 0.5, PaddingX: 0, PaddingY: 80}
	detection := Detection{X: 320, Y: 320, Width: 100, Height: 200}

	left, top, right, bottom := modelBoundsToOriginal(detection, prepared)
	if left != 540 || top != 280 || right != 740 || bottom != 680 {
		t.Fatalf("box = (%v, %v, %v, %v), want (540, 280, 740, 680)", left, top, right, bottom)
	}
}

func TestSuppressOverlappingDetectionsRemovesLowerConfidenceBoxes(t *testing.T) {
	detections := []Detection{
		{Class: "hail-damage", Confidence: 0.7, X: 105, Y: 105, Width: 100, Height: 100},
		{Class: "hail-damage", Confidence: 0.9, X: 100, Y: 100, Width: 100, Height: 100},
		{Class: "hail-damage", Confidence: 0.8, X: 400, Y: 400, Width: 80, Height: 80},
	}

	kept := suppressOverlappingDetections(detections, 0.45)
	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2", len(kept))
	}
	if kept[0].Confidence != 0.9 || kept[1].Confidence != 0.8 {
		t.Fatalf("unexpected confidences: %v, %v", kept[0].Confidence, kept[1].Confidence)
	}
}
