package converter

import (
	"math"

	"github.com/lestrrat-go/polyclip"
	"github.com/lestrrat-go/polyclip/geom"
)

// GenerateBoxBoundary creates a rectangular or rounded-corner enclosure (box, rounded-box, pill/capsule)
// padded around the given bounding box.
func GenerateBoxBoundary(bounds BoundingBox, padX, padY, radius float64, samples int) Contour {
	if padX < 0 {
		padX = 0
	}
	if padY < 0 {
		padY = padX
	}
	if samples < 4 {
		samples = 6
	}

	minX := bounds.MinX - padX
	maxX := bounds.MaxX + padX
	minY := bounds.MinY - padY
	maxY := bounds.MaxY + padY

	w := maxX - minX
	h := maxY - minY

	maxR := math.Min(w/2.0, h/2.0)
	if radius < 0 {
		radius = 0
	}
	if radius > maxR {
		radius = maxR
	}

	// Sharp rectangle
	if radius < 1e-4 {
		return Contour{
			Points: []Point{
				{X: minX, Y: minY},
				{X: minX, Y: maxY},
				{X: maxX, Y: maxY},
				{X: maxX, Y: minY},
			},
			Closed: true,
		}
	}

	// Rounded rectangle / pill with arc sampling
	var pts []Point

	// Helper to sample an arc from startAngle to endAngle (clockwise in Cartesian)
	sampleArc := func(cx, cy, r, startAngle, endAngle float64) {
		sweep := endAngle - startAngle
		for i := 0; i <= samples; i++ {
			t := float64(i) / float64(samples)
			a := startAngle + t*sweep
			pts = append(pts, Point{
				X: cx + r*math.Cos(a),
				Y: cy + r*math.Sin(a),
			})
		}
	}

	// 1. Top-left arc: center (minX+radius, maxY-radius), angle pi to pi/2
	sampleArc(minX+radius, maxY-radius, radius, math.Pi, math.Pi/2.0)

	// 2. Top-right arc: center (maxX-radius, maxY-radius), angle pi/2 to 0
	sampleArc(maxX-radius, maxY-radius, radius, math.Pi/2.0, 0)

	// 3. Bottom-right arc: center (maxX-radius, minY+radius), angle 0 to -pi/2
	sampleArc(maxX-radius, minY+radius, radius, 0, -math.Pi/2.0)

	// 4. Bottom-left arc: center (minX+radius, minY+radius), angle -pi/2 to -pi
	sampleArc(minX+radius, minY+radius, radius, -math.Pi/2.0, -math.Pi)

	cleaned := CleanContour(pts, 1e-4)
	return Contour{
		Points: cleaned,
		Closed: true,
	}
}

// GenerateConformalBoundary generates a contour bubble around the filled text.
// The bubble is the boundary of the text region expanded by distance d (a round,
// miter, or bevel morphological offset). Nearby letters merge where their
// expansions meet, and concave notches are trimmed instead of growing spikes.
// When fillHoles is set, interior counters are treated as solid so the badge
// is one continuous plate; otherwise counters remain as shrunk cutouts.
func GenerateConformalBoundary(contours []Contour, d float64, fillHoles bool, join CornerJoin, miterLimit float64) []Contour {
	if len(contours) == 0 {
		return nil
	}
	if d <= 0 {
		d = 3.0
	}
	if join == "" {
		join = JoinRound
	}
	if miterLimit <= 1.0 {
		miterLimit = 3.0
	}

	region := materialMultiPolygon(contours, fillHoles)
	if len(region) == 0 {
		region = materialMultiPolygon(contours, false)
	}
	// Soften micro-hairpins on engraved/script faces before offsetting.
	// Keep epsilon modest so flat stems stay flat (large values sawtooth).
	preEps := math.Max(0.08, d*0.02)
	if simplified := polyclip.SimplifyPaths(region, preEps); len(simplified) > 0 {
		region = simplified
	}
	expanded := offsetSolid(region, d, join, miterLimit)
	if len(expanded) == 0 {
		return nil
	}

	if fillHoles {
		expanded = dropHolesAndNested(expanded)
	}

	raw := multiToContours(expanded)
	if len(raw) == 0 {
		return nil
	}

	classified := ClassifyAndOrientContours(raw, 1e-4)
	minArea := math.Max(1e-3, math.Pi*d*d*0.05)

	var result []Contour
	for _, cc := range classified {
		if fillHoles && cc.Role != RoleOuter {
			continue
		}
		if cc.Area < minArea {
			continue
		}
		result = append(result, Contour{
			Points: cc.Points,
			Closed: true,
		})
	}
	return result
}

// dropHolesAndNested strips holes and discards any piece whose representative
// point lies inside a larger piece — leftover islands from offset self-union.
func dropHolesAndNested(m geom.MultiPolygon) geom.MultiPolygon {
	if len(m) == 0 {
		return m
	}
	type piece struct {
		outer geom.Polygon
		area  float64
		pts   []Point
	}
	pieces := make([]piece, 0, len(m))
	for _, ex := range m {
		if len(ex.Outer) < 3 {
			continue
		}
		pts := fromRing(ex.Outer)
		if len(pts) < 3 {
			continue
		}
		a := math.Abs(SignedArea(pts))
		if a < 1e-8 {
			continue
		}
		pieces = append(pieces, piece{outer: ex.Outer, area: a, pts: pts})
	}
	bestA := 0.0
	for _, p := range pieces {
		if p.area > bestA {
			bestA = p.area
		}
	}
	var kept geom.MultiPolygon
	for i, p := range pieces {
		if p.area < bestA*0.15 {
			continue
		}
		hits := 0
		samples := p.pts
		if len(samples) > 8 {
			samples = samples[:8]
		}
		for _, q := range samples {
			for j, o := range pieces {
				if i == j || o.area <= p.area {
					continue
				}
				if PointInPolygon(q, o.pts) {
					hits++
					break
				}
			}
		}
		if len(samples) > 0 && hits*2 >= len(samples) {
			continue
		}
		kept = append(kept, geom.ExPolygon{Outer: p.outer})
	}
	return kept
}
