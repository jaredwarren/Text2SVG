package converter

import (
	"os"
	"strings"
	"testing"
)

func TestFullConversionPipeline(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("System Arial font not found, skipping test:", err)
		return
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont failed: %v", err)
	}

	params := TextParams{
		Text:         "DXF Cut 123",
		Size:         30.0, // 30 mm
		Units:        UnitsMM,
		Kerning:      1.5, // 1.5 mm
		LineHeight:   1.2,
		Datum:        DatumBottomLeft,
		CurveSamples: 16,
		LayerName:    "CUT",
	}

	result, err := lf.LayoutText(params)
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	if result.GlyphCount == 0 {
		t.Errorf("Expected glyphs, got 0")
	}
	if len(result.Contours) == 0 {
		t.Errorf("Expected contours, got 0")
	}

	t.Logf("Bounds: Width=%.2f mm, Height=%.2f mm (MinX=%.2f, MinY=%.2f, MaxX=%.2f, MaxY=%.2f)",
		result.Bounds.Width(), result.Bounds.Height(),
		result.Bounds.MinX, result.Bounds.MinY, result.Bounds.MaxX, result.Bounds.MaxY)

	// Since datum is BottomLeft, MinX and MinY should be 0.0 (or very close)
	if result.Bounds.MinX != 0.0 || result.Bounds.MinY != 0.0 {
		t.Errorf("Expected MinX and MinY to be 0.0 with DatumBottomLeft, got MinX=%f, MinY=%f",
			result.Bounds.MinX, result.Bounds.MinY)
	}

	// Verify DXF structure
	if !strings.Contains(result.DXF, "AC1009") {
		t.Errorf("Expected DXF to contain AC1009 header")
	}
	if !strings.Contains(result.DXF, "POLYLINE") {
		t.Errorf("Expected DXF to contain POLYLINE entity")
	}
	if !strings.Contains(result.DXF, "EOF") {
		t.Errorf("Expected DXF to contain EOF")
	}

	// Verify SVG structure
	if !strings.Contains(result.SVG, "<svg") || !strings.Contains(result.SVG, "</svg>") {
		t.Errorf("Expected valid SVG tags")
	}
	if !strings.Contains(result.SVG, "viewBox=") {
		t.Errorf("Expected SVG viewBox attribute")
	}

	t.Logf("DXF length: %d bytes, SVG length: %d bytes", len(result.DXF), len(result.SVG))
}
