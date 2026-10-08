package converter

import (
	"math"
	"os"
	"testing"
)

func TestGenerateBoxBoundary_Sharp(t *testing.T) {
	bounds := BoundingBox{MinX: 10, MinY: 20, MaxX: 50, MaxY: 40}
	box := GenerateBoxBoundary(bounds, 5, 5, 0, 8)

	if len(box.Points) != 4 {
		t.Fatalf("expected 4 vertices for sharp box, got %d", len(box.Points))
	}

	// Verify extents
	minX, maxX := 1e9, -1e9
	minY, maxY := 1e9, -1e9
	for _, p := range box.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	if math.Abs(minX-5) > 1e-3 || math.Abs(maxX-55) > 1e-3 {
		t.Errorf("expected X bounds [5, 55], got [%.2f, %.2f]", minX, maxX)
	}
	if math.Abs(minY-15) > 1e-3 || math.Abs(maxY-45) > 1e-3 {
		t.Errorf("expected Y bounds [15, 45], got [%.2f, %.2f]", minY, maxY)
	}
}

func TestGenerateBoxBoundary_RoundedAndPill(t *testing.T) {
	bounds := BoundingBox{MinX: 0, MinY: 0, MaxX: 100, MaxY: 20}
	// Rounded rectangle with R=5
	box := GenerateBoxBoundary(bounds, 10, 5, 5, 6)
	if len(box.Points) < 16 {
		t.Errorf("expected >= 16 points for rounded box, got %d", len(box.Points))
	}

	// Pill / Capsule mode: R large (e.g. 50, capped to height/2 = (20+10)/2 = 15)
	pill := GenerateBoxBoundary(bounds, 10, 5, 50, 8)
	if len(pill.Points) < 16 {
		t.Errorf("expected >= 16 points for pill box, got %d", len(pill.Points))
	}
}

func TestGenerateBoxBoundary_AsymmetricPadding(t *testing.T) {
	bounds := BoundingBox{MinX: 0, MinY: 0, MaxX: 40, MaxY: 20}
	padX := 15.0
	padY := 5.0
	box := GenerateBoxBoundary(bounds, padX, padY, 0, 6)

	minX, maxX := 1e9, -1e9
	minY, maxY := 1e9, -1e9
	for _, p := range box.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	expectedW := 40.0 + 2*padX
	expectedH := 20.0 + 2*padY

	actualW := maxX - minX
	actualH := maxY - minY

	if math.Abs(actualW-expectedW) > 1e-3 {
		t.Errorf("expected width %.2f, got %.2f", expectedW, actualW)
	}
	if math.Abs(actualH-expectedH) > 1e-3 {
		t.Errorf("expected height %.2f, got %.2f", expectedH, actualH)
	}
}

func TestGenerateConformalBoundary(t *testing.T) {
	// Two square glyphs close to each other
	c1 := Contour{
		Points: []Point{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}},
		Closed: true,
	}
	c2 := Contour{
		Points: []Point{{X: 12, Y: 0}, {X: 12, Y: 10}, {X: 22, Y: 10}, {X: 22, Y: 0}},
		Closed: true,
	}

	// With offset = 3.0, the two squares (gap of 2.0) should overlap and union into 1 continuous bubble
	res := GenerateConformalBoundary([]Contour{c1, c2}, 3.0, true, JoinRound, 3.0)
	if len(res) == 0 {
		t.Fatalf("expected non-empty conformal boundary")
	}

	// Should union into a single outer contour
	if len(res) != 1 {
		t.Errorf("expected 1 merged bubble contour, got %d", len(res))
	}

	// Check that the merged boundary encloses both squares
	minX, maxX := 1e9, -1e9
	for _, p := range res[0].Points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
	}

	if minX >= 0 {
		t.Errorf("expected boundary minX < 0, got %.2f", minX)
	}
	if maxX <= 22 {
		t.Errorf("expected boundary maxX > 22, got %.2f", maxX)
	}
}

func TestConformalSilhouette_Word(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Arial.ttf")
	if err != nil {
		t.Skip("System Arial font not found, skipping:", err)
	}
	lf, err := ParseFont(data, "Arial")
	if err != nil {
		t.Fatalf("ParseFont failed: %v", err)
	}

	res, err := lf.LayoutText(TextParams{
		Text:              "SILHOUETTE",
		Size:              30.0,
		Units:             UnitsMM,
		BoundaryMode:      BoundaryConformal,
		BoundaryOffset:    3.0,
		BoundaryFillHoles: true,
	})
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	t.Logf("Boundary contours count: %d", len(res.BoundaryContours))
	for i, bc := range res.BoundaryContours {
		bb := ContourBoundingBox(bc.Points)
		t.Logf("Contour %d: %d pts, area: %.2f, bounds: [%.2f, %.2f] to [%.2f, %.2f]", i, len(bc.Points), math.Abs(SignedArea(bc.Points)), bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
	}
	if len(res.BoundaryContours) == 0 {
		t.Fatalf("expected non-empty boundary contours")
	}
}

func TestConformalSilhouette_AcademyLA(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf")
	if err != nil {
		t.Skip("Academy Engraved font not found, skipping:", err)
	}
	lf, err := ParseFont(data, "Academy Engraved LET Fonts")
	if err != nil {
		t.Fatalf("ParseFont failed: %v", err)
	}

	res, err := lf.LayoutText(TextParams{
		Text:              "LA",
		Size:              30.0,
		Units:             UnitsMM,
		BoundaryMode:      BoundaryConformal,
		BoundaryOffset:    4.0,
		BoundaryFillHoles: true,
	})
	if err != nil {
		t.Fatalf("LayoutText failed: %v", err)
	}

	t.Logf("Academy LA boundary count: %d", len(res.BoundaryContours))
	for i, bc := range res.BoundaryContours {
		bb := ContourBoundingBox(bc.Points)
		t.Logf("Contour %d: %d pts, area: %.2f, bounds: [%.2f, %.2f] to [%.2f, %.2f]",
			i, len(bc.Points), math.Abs(SignedArea(bc.Points)), bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
	}

	// Verify text bounds vs boundary bounds
	t.Logf("Overall Bounds: W=%.2f, H=%.2f, min=(%.2f, %.2f) max=(%.2f, %.2f)",
		res.Bounds.Width(), res.Bounds.Height(), res.Bounds.MinX, res.Bounds.MinY, res.Bounds.MaxX, res.Bounds.MaxY)

	textB := ContourBoundingBox(res.Contours[0].Points)
	for _, c := range res.Contours[1:] {
		cb := ContourBoundingBox(c.Points)
		if cb.MinX < textB.MinX {
			textB.MinX = cb.MinX
		}
		if cb.MaxX > textB.MaxX {
			textB.MaxX = cb.MaxX
		}
		if cb.MinY < textB.MinY {
			textB.MinY = cb.MinY
		}
		if cb.MaxY > textB.MaxY {
			textB.MaxY = cb.MaxY
		}
	}
	t.Logf("Text geometry bounds: [%.2f, %.2f] to [%.2f, %.2f]", textB.MinX, textB.MinY, textB.MaxX, textB.MaxY)
	boundaryB := ContourBoundingBox(res.BoundaryContours[0].Points)
	t.Logf("Boundary bounds:      [%.2f, %.2f] to [%.2f, %.2f]", boundaryB.MinX, boundaryB.MinY, boundaryB.MaxX, boundaryB.MaxY)

	// Boundary MUST enclose text: boundary min <= text min, boundary max >= text max
	if boundaryB.MinX > textB.MinX || boundaryB.MaxX < textB.MaxX ||
		boundaryB.MinY > textB.MinY || boundaryB.MaxY < textB.MaxY {
		t.Fatalf("Boundary does NOT enclose text! Boundary: %+v, Text: %+v", boundaryB, textB)
	}
	t.Logf("SUCCESS: Boundary cleanly encloses text on all sides!")
}

func TestConformalBubbleTracksText(t *testing.T) {
	fonts := []string{
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
		"/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf",
		"/System/Library/Fonts/Supplemental/SnellRoundhand.ttf",
		"/System/Library/Fonts/Supplemental/Brush Script.ttf",
	}
	words := []string{"6", "A", "gy", "LA", "B8", "LASER"}
	distances := []float64{0.5, 1, 2, 4, 8, 12}

	for _, fp := range fonts {
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		lf, err := ParseFont(data, fp)
		if err != nil {
			t.Errorf("parse %s: %v", fp, err)
			continue
		}
		for _, word := range words {
			for _, d := range distances {
				res, err := lf.LayoutText(TextParams{
					Text:              word,
					Size:              30.0,
					Units:             UnitsMM,
					BoundaryMode:      BoundaryConformal,
					BoundaryOffset:    d,
					BoundaryFillHoles: true,
					CornerJoin:        JoinRound,
					Datum:             DatumCenter,
				})
				if err != nil {
					t.Errorf("%s %q d=%.1f: %v", fp, word, d, err)
					continue
				}
				if len(res.BoundaryContours) == 0 {
					t.Errorf("%s %q d=%.1f: empty bubble", fp, word, d)
					continue
				}

				var source [][]Point
				for _, c := range res.Contours {
					if len(c.Points) >= 2 {
						source = append(source, c.Points)
					}
				}

				for ci, bc := range res.BoundaryContours {
					n := len(bc.Points)
					hits := 0
					for i := 0; i < n; i++ {
						p1 := bc.Points[i]
						p2 := bc.Points[(i+1)%n]
						for j := i + 2; j < n; j++ {
							if (j+1)%n == i {
								continue
							}
							if _, ok := SegmentsIntersect(p1, p2, bc.Points[j], bc.Points[(j+1)%n]); ok {
								hits++
							}
						}
					}
					if hits > 0 {
						t.Errorf("%s %q d=%.1f contour %d self-intersects (%d)", fp, word, d, ci, hits)
					}

					minDist := math.Inf(1)
					maxDist := 0.0
					for _, p := range bc.Points {
						dist := distToPolylines(p, source)
						if dist < minDist {
							minDist = dist
						}
						if dist > maxDist {
							maxDist = dist
						}
					}
					// Round offset vertices sit on the d-circle. A chord or spike
					// pulls that range well away from d.
					if minDist < d*0.72 {
						t.Errorf("%s %q d=%.1f contour %d cuts inward: min dist %.2f", fp, word, d, ci, minDist)
					}
					if maxDist > d*1.35+0.4 {
						t.Errorf("%s %q d=%.1f contour %d spikes outward: max dist %.2f", fp, word, d, ci, maxDist)
					}
				}

				for _, c := range res.Contours {
					for _, p := range c.Points {
						if !pointInAnyOuter(p, res.BoundaryContours) {
							t.Errorf("%s %q d=%.1f: bubble does not enclose text point (%.2f, %.2f)", fp, word, d, p.X, p.Y)
							break
						}
					}
				}
			}
		}
	}
}

func distToPolylines(p Point, lines [][]Point) float64 {
	best := math.Inf(1)
	for _, ln := range lines {
		n := len(ln)
		if n < 2 {
			continue
		}
		for i := 0; i < n; i++ {
			if e := distPointSeg(p, ln[i], ln[(i+1)%n]); e < best {
				best = e
			}
		}
	}
	return best
}

func distPointSeg(p, a, b Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	l2 := dx*dx + dy*dy
	if l2 < 1e-18 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return math.Hypot(p.X-(a.X+t*dx), p.Y-(a.Y+t*dy))
}

func pointInAnyOuter(p Point, contours []Contour) bool {
	for _, c := range contours {
		if len(c.Points) >= 3 && PointInPolygon(p, c.Points) {
			return true
		}
	}
	return false
}

func TestContourOffsetDistances(t *testing.T) {
	fontPaths := []string{
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
	}

	testWords := []string{"LASER CUT 2026", "A", "TEXT", "MINIMAL"}

	for _, fp := range fontPaths {
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		lf, err := ParseFont(data, fp)
		if err != nil {
			continue
		}

		for _, word := range testWords {
			for _, fillHoles := range []bool{true, false} {
				for _, join := range []CornerJoin{JoinRound, JoinMiter, JoinBevel} {
					for d := 1.0; d <= 8.0; d += 2.0 {
						res, err := lf.LayoutText(TextParams{
							Text:              word,
							Size:              30.0,
							Units:             UnitsMM,
							BoundaryMode:      BoundaryConformal,
							BoundaryOffset:    d,
							BoundaryFillHoles: fillHoles,
							CornerJoin:        join,
							Datum:             DatumCenter,
						})
						if err != nil {
							t.Errorf("Font %s, word %s, d=%.1f, fill=%v, join=%s failed: %v", fp, word, d, fillHoles, join, err)
							continue
						}

						for cIdx, c := range res.BoundaryContours {
							n := len(c.Points)
							selfIntersects := 0
							for i := 0; i < n; i++ {
								p1 := c.Points[i]
								p2 := c.Points[(i+1)%n]
								for j := i + 2; j < n; j++ {
									if (j+1)%n == i {
										continue
									}
									p3 := c.Points[j]
									p4 := c.Points[(j+1)%n]
									if _, ok := SegmentsIntersect(p1, p2, p3, p4); ok {
										selfIntersects++
									}
								}
							}
							if selfIntersects > 0 {
								t.Errorf("FAIL Boundary: Font %s, word %s, d=%.1f, fill=%v, join=%s, contour %d has %d self-intersections! Points=%d",
									fp, word, d, fillHoles, join, cIdx, selfIntersects, n)
							}
						}
					}
				}
			}
		}
	}
}
