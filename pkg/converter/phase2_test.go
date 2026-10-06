package converter

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestPhase2Slant(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("Arial not found")
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont error: %v", err)
	}

	// Unslanted text
	resStraight, err := lf.LayoutText(TextParams{
		Text: "CAD",
		Size: 20.0,
	})
	if err != nil {
		t.Fatalf("Straight text error: %v", err)
	}

	// Slanted text (+15° drafting italic)
	resSlanted, err := lf.LayoutText(TextParams{
		Text:       "CAD",
		Size:       20.0,
		SlantAngle: 15.0,
	})
	if err != nil {
		t.Fatalf("Slanted text error: %v", err)
	}

	t.Logf("Straight bounds: W=%.2f, H=%.2f", resStraight.Bounds.Width(), resStraight.Bounds.Height())
	t.Logf("Slanted bounds: W=%.2f, H=%.2f", resSlanted.Bounds.Width(), resSlanted.Bounds.Height())

	// Slanting +15° forward increases the width of the bounding box by height * tan(15°)
	expectedExtraWidth := 20.0 * math.Tan(15.0*math.Pi/180.0)
	actualExtraWidth := resSlanted.Bounds.Width() - resStraight.Bounds.Width()
	t.Logf("Expected extra width: ~%.2f, actual extra width: %.2f", expectedExtraWidth, actualExtraWidth)

	if actualExtraWidth <= 0 {
		t.Errorf("Expected slanted text to have wider bounding box")
	}
}

func TestPhase2ArcText(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("Arial not found")
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont error: %v", err)
	}

	// Circular text on a 50mm radius bezel with 120° sweep
	resArc, err := lf.LayoutText(TextParams{
		Text:       "ROTARY DIAL 2026",
		Size:       10.0,
		Units:      UnitsMM,
		ArcEnabled: true,
		ArcRadius:  50.0,
		ArcSweep:   120.0,
		ArcAlign:   ArcAlignCenter,
		ArcInward:  false,
		DXFFormat:  DXFFormatSpline,
	})
	if err != nil {
		t.Fatalf("Arc text layout error: %v", err)
	}

	if resArc.PathCount == 0 {
		t.Errorf("Expected contours in curved text")
	}

	t.Logf("Arc bounds: W=%.2f mm, H=%.2f mm, paths=%d",
		resArc.Bounds.Width(), resArc.Bounds.Height(), resArc.PathCount)

	// Verify DXF output contains valid AC1015 entities
	if !strings.Contains(resArc.DXF, "AC1015") {
		t.Errorf("Expected AC1015 in curved DXF export")
	}
	if !strings.Contains(resArc.SVG, "<svg") {
		t.Errorf("Expected valid SVG in curved SVG export")
	}
}

func TestPhase2FullCombinedPipeline(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("Arial not found")
	}

	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont error: %v", err)
	}

	// Slanted, curved along arc, welded, with a tolerance offset for 3D printed press-fit inlay!
	resCombined, err := lf.LayoutText(TextParams{
		Text:       "INLAY 3D",
		Size:       15.0,
		Units:      UnitsMM,
		SlantAngle: 10.0,
		ArcEnabled: true,
		ArcRadius:  60.0,
		ArcSweep:   90.0,
		Weld:       true,
		Offset:     -0.2, // -0.2mm inset for push-fit tolerance
		CornerJoin: JoinRound,
		DXFFormat:  DXFFormatSpline,
	})
	if err != nil {
		t.Fatalf("Combined pipeline error: %v", err)
	}

	t.Logf("Combined pipeline: W=%.2f mm, H=%.2f mm, loops=%d, welded=%v",
		resCombined.Bounds.Width(), resCombined.Bounds.Height(), resCombined.PathCount, resCombined.Welded)

	if resCombined.PathCount == 0 {
		t.Errorf("Expected valid contours in combined pipeline output")
	}
}
