package converter

import (
	"os"
	"testing"
)

func TestWeldBrushScript(t *testing.T) {
	paths := []string{
		"/System/Library/Fonts/Supplemental/Brush Script.ttf",
		"/Library/Fonts/Brush Script.ttf",
	}
	var data []byte
	for _, p := range paths {
		if d, err := os.ReadFile(p); err == nil {
			data = d
			break
		}
	}
	if data == nil {
		t.Skip("Brush Script not found")
	}

	lf, err := ParseFont(data, "Brush Script")
	if err != nil {
		t.Fatalf("ParseFont error: %v", err)
	}

	// 1. Unwelded layout with negative kerning: letters overlap with separate crossing loops
	unweldedRes, err := lf.LayoutText(TextParams{
		Text:       "California",
		Size:       30.0,
		Units:      UnitsMM,
		Kerning:    -0.8,
		Datum:      DatumBottomLeft,
		Weld:       false,
	})
	if err != nil {
		t.Fatalf("Unwelded LayoutText error: %v", err)
	}

	// 2. Welded layout: overlapping strokes merged into unified continuous outer loops
	weldedRes, err := lf.LayoutText(TextParams{
		Text:       "California",
		Size:       30.0,
		Units:      UnitsMM,
		Kerning:    -0.8,
		Datum:      DatumBottomLeft,
		Weld:       true,
	})
	if err != nil {
		t.Fatalf("Welded LayoutText error: %v", err)
	}

	t.Logf("California unwelded path count: %d, welded path count: %d", unweldedRes.PathCount, weldedRes.PathCount)
	if !weldedRes.Welded {
		t.Errorf("Expected weldedRes.Welded to be true")
	}
	// In cursive script, welding multiple overlapping glyphs merges multiple loops into fewer merged boundaries
	if weldedRes.PathCount >= unweldedRes.PathCount {
		t.Errorf("Expected welded path count (%d) to be less than unwelded (%d)", weldedRes.PathCount, unweldedRes.PathCount)
	}

	// Verify all contours in welded result are topologically classified and oriented
	if len(weldedRes.ClassifiedPaths) == 0 {
		t.Errorf("Expected classified paths in welded result")
	}
	for i, cp := range weldedRes.ClassifiedPaths {
		cw := IsClockwise(cp.Points)
		if cp.Role == RoleOuter && !cw {
			t.Errorf("Contour %d role is Outer but is not Clockwise", i)
		}
		if cp.Role == RoleHole && cw {
			t.Errorf("Contour %d role is Hole but is not Counter-Clockwise", i)
		}
	}
}

func TestManifoldOrientationAndHoles(t *testing.T) {
	// Outer box (0,0) to (100,100)
	outer := Contour{
		Points: []Point{{X: 0, Y: 0}, {X: 0, Y: 100}, {X: 100, Y: 100}, {X: 100, Y: 0}},
		Closed: true,
	}
	// Hole 1 (10,10) to (40,40)
	hole1 := Contour{
		Points: []Point{{X: 10, Y: 10}, {X: 40, Y: 10}, {X: 40, Y: 40}, {X: 10, Y: 40}},
		Closed: true,
	}
	// Hole 2 (60,60) to (90,90)
	hole2 := Contour{
		Points: []Point{{X: 60, Y: 60}, {X: 90, Y: 60}, {X: 90, Y: 90}, {X: 60, Y: 90}},
		Closed: true,
	}

	classified := ClassifyAndOrientContours([]Contour{outer, hole1, hole2}, 1e-4)
	if len(classified) != 3 {
		t.Fatalf("Expected 3 classified contours, got %d", len(classified))
	}

	outerCount := 0
	holeCount := 0
	for _, c := range classified {
		if c.Role == RoleOuter {
			outerCount++
			if !IsClockwise(c.Points) {
				t.Errorf("Outer contour is not clockwise")
			}
			if c.Depth != 0 {
				t.Errorf("Expected depth 0 for outer, got %d", c.Depth)
			}
		} else if c.Role == RoleHole {
			holeCount++
			if IsClockwise(c.Points) {
				t.Errorf("Hole contour is not counter-clockwise")
			}
			if c.Depth != 1 {
				t.Errorf("Expected depth 1 for hole, got %d", c.Depth)
			}
		}
	}

	if outerCount != 1 || holeCount != 2 {
		t.Errorf("Expected 1 outer and 2 holes, got %d outer and %d holes", outerCount, holeCount)
	}
}

func TestCleanContour(t *testing.T) {
	// Contour with consecutive duplicates, zero-length micro-segments, and collinear points
	raw := []Point{
		{X: 0, Y: 0},
		{X: 0, Y: 0}, // duplicate
		{X: 5, Y: 0}, // collinear midpoint on line (0,0) -> (10,0)
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
		{X: 0, Y: 0}, // closing duplicate
	}

	cleaned := CleanContour(raw, 1e-4)
	// After removing collinear midpoint and duplicate points, should be a clean rectangle of 4 vertices
	if len(cleaned) != 4 {
		t.Errorf("Expected 4 points after CleanContour, got %d: %v", len(cleaned), cleaned)
	}
}
