package converter

import (
	"os"
	"strings"
	"testing"
)

func TestPhase3Datums(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("System Arial font not found, skipping:", err)
		return
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont failed: %v", err)
	}

	// 1. Test Baseline-Left datum:
	// For baseline-left, typographical baseline is at Y=0, and left edge is at X=0
	resBaselineLeft, err := lf.LayoutText(TextParams{
		Text:      "BASELINE",
		Size:      30.0,
		Units:     UnitsMM,
		Datum:     DatumBaselineLeft,
		DXFFormat: DXFFormatPolyline,
	})
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	if resBaselineLeft.Bounds.MinX != 0.0 {
		t.Errorf("Expected MinX=0.0 for DatumBaselineLeft, got %.4f", resBaselineLeft.Bounds.MinX)
	}
	// Arial letters like 'B', 'A', 'S' sit on baseline Y=0, with slight descenders if any, so MinY should be <= 0 and MaxY > 0
	if resBaselineLeft.Bounds.MaxY <= 0 {
		t.Errorf("Expected MaxY > 0 for baseline text, got %.4f", resBaselineLeft.Bounds.MaxY)
	}

	// 2. Test Baseline-Center datum:
	// For baseline-center, X is centered around 0 (MinX = -MaxX)
	resBaselineCenter, err := lf.LayoutText(TextParams{
		Text:      "BASELINE",
		Size:      30.0,
		Units:     UnitsMM,
		Datum:     DatumBaselineCenter,
		DXFFormat: DXFFormatPolyline,
	})
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	diffCenter := resBaselineCenter.Bounds.MinX + resBaselineCenter.Bounds.MaxX
	if diffCenter > 0.01 || diffCenter < -0.01 {
		t.Errorf("Expected X center around 0.0, got sum MinX+MaxX=%.4f (MinX=%.4f, MaxX=%.4f)",
			diffCenter, resBaselineCenter.Bounds.MinX, resBaselineCenter.Bounds.MaxX)
	}
}

func TestPhase3ConstructionBox(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("System Arial font not found, skipping:", err)
		return
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont failed: %v", err)
	}

	res, err := lf.LayoutText(TextParams{
		Text:            "CAD BOX",
		Size:            25.0,
		Units:           UnitsMM,
		Datum:           DatumBottomLeft,
		DXFFormat:       DXFFormatPolyline,
		ConstructionBox: true,
	})
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	// Check DXF for CONSTRUCTION layer
	if !strings.Contains(res.DXF, "CONSTRUCTION") {
		t.Errorf("Expected DXF to contain CONSTRUCTION layer")
	}

	// Check SVG for construction-frame rect
	if !strings.Contains(res.SVG, "construction-frame") {
		t.Errorf("Expected SVG to contain construction-frame rect")
	}

	// Check SVG for evenodd fill-rule
	if !strings.Contains(res.SVG, `fill-rule="evenodd"`) {
		t.Errorf("Expected SVG to contain fill-rule=\"evenodd\" attribute")
	}
}
