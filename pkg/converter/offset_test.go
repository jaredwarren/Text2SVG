package converter

import (
	"math"
	"os"
	"testing"
)

func TestOffsetSquare(t *testing.T) {
	// Clockwise square from (0,0) to (10,10): (0,10) -> (10,10) -> (10,0) -> (0,0)
	sq := []Point{
		{X: 0, Y: 10},
		{X: 10, Y: 10},
		{X: 10, Y: 0},
		{X: 0, Y: 0},
	}

	// 1. Outward offset by +2.0 (miter)
	outward := OffsetContour(sq, 2.0, JoinMiter, 3.0)
	bOut := ContourBoundingBox(outward)
	t.Logf("Outward bounds: min=(%.2f, %.2f) max=(%.2f, %.2f) pts=%d", bOut.MinX, bOut.MinY, bOut.MaxX, bOut.MaxY, len(outward))
	if math.Abs(bOut.MinX-(-2.0)) > 0.1 || math.Abs(bOut.MaxX-12.0) > 0.1 {
		t.Errorf("Unexpected outward bounds: %v", bOut)
	}

	// 2. Inward offset by -2.0 (miter)
	inward := OffsetContour(sq, -2.0, JoinMiter, 3.0)
	bIn := ContourBoundingBox(inward)
	t.Logf("Inward bounds: min=(%.2f, %.2f) max=(%.2f, %.2f) pts=%d", bIn.MinX, bIn.MinY, bIn.MaxX, bIn.MaxY, len(inward))
	if math.Abs(bIn.MinX-2.0) > 0.1 || math.Abs(bIn.MaxX-8.0) > 0.1 {
		t.Errorf("Unexpected inward bounds: %v", bIn)
	}
}

func TestOffsetLetterO(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("Arial font not found")
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont error: %v", err)
	}

	rawO, err := lf.LoadGlyph('O', 20)
	if err != nil {
		t.Fatalf("LoadGlyph error: %v", err)
	}

	classified := ClassifyAndOrientContours(rawO.Contours, 1e-4)
	if len(classified) != 2 {
		t.Fatalf("Expected 2 contours for 'O', got %d", len(classified))
	}

	origOuterBox := ContourBoundingBox(classified[0].Points)
	origHoleBox := ContourBoundingBox(classified[1].Points)

	// Expand profile by +15 font units (scale ~ 1/100)
	d := 15.0
	offsetResult := OffsetProfile(classified, d, JoinRound, 3.0)
	if len(offsetResult) != 2 {
		t.Fatalf("Expected 2 contours after offset, got %d", len(offsetResult))
	}

	newClassified := ClassifyAndOrientContours(offsetResult, 1e-4)
	newOuterBox := ContourBoundingBox(newClassified[0].Points)
	newHoleBox := ContourBoundingBox(newClassified[1].Points)

	t.Logf("Original Outer: W=%.1f H=%.1f, Offset Outer: W=%.1f H=%.1f",
		origOuterBox.Width(), origOuterBox.Height(), newOuterBox.Width(), newOuterBox.Height())
	t.Logf("Original Hole: W=%.1f H=%.1f, Offset Hole: W=%.1f H=%.1f",
		origHoleBox.Width(), origHoleBox.Height(), newHoleBox.Width(), newHoleBox.Height())

	// Outer should be larger
	if newOuterBox.Width() <= origOuterBox.Width() {
		t.Errorf("Expected outer width to expand")
	}
	// Hole should be smaller (material expanded inward into hole)
	if newHoleBox.Width() >= origHoleBox.Width() {
		t.Errorf("Expected hole width to shrink")
	}
}
