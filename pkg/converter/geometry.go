package converter

import (
	"math"
)

// SignedArea returns the signed area of a 2D contour using the shoelace formula.
// In Cartesian coordinates (Y points UP):
//   - Positive (> 0): Counter-Clockwise (CCW)
//   - Negative (< 0): Clockwise (CW)
func SignedArea(pts []Point) float64 {
	n := len(pts)
	if n < 3 {
		return 0
	}
	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return area * 0.5
}

// IsClockwise returns true if the vertices of pts wind clockwise in Cartesian coordinates (Y up).
func IsClockwise(pts []Point) bool {
	return SignedArea(pts) < 0
}

// EnsureOrientation reverses the points of the contour if necessary to match desired CW/CCW.
// If makeCW is true, ensures points wind Clockwise.
// If makeCW is false, ensures points wind Counter-Clockwise.
func EnsureOrientation(pts []Point, makeCW bool) []Point {
	cw := IsClockwise(pts)
	if cw == makeCW {
		return pts
	}
	// Reverse points
	n := len(pts)
	rev := make([]Point, n)
	for i := 0; i < n; i++ {
		rev[i] = pts[n-1-i]
	}
	return rev
}

// PointInPolygon tests if testPt is inside a closed polygon using ray casting.
func PointInPolygon(testPt Point, poly []Point) bool {
	n := len(poly)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := poly[i].X, poly[i].Y
		xj, yj := poly[j].X, poly[j].Y

		intersect := ((yi > testPt.Y) != (yj > testPt.Y)) &&
			(testPt.X < (xj-xi)*(testPt.Y-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

// ContourBoundingBox returns the bounding box for a set of points.
func ContourBoundingBox(pts []Point) BoundingBox {
	if len(pts) == 0 {
		return BoundingBox{}
	}
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := pts[0].X, pts[0].Y
	for _, p := range pts[1:] {
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
	return BoundingBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

// IsContourInside tests if child is geometrically inside parent.
func IsContourInside(child, parent []Point) bool {
	if len(child) < 3 || len(parent) < 3 {
		return false
	}
	bChild := ContourBoundingBox(child)
	bParent := ContourBoundingBox(parent)

	// Quick bounding box rejection
	if bChild.MinX < bParent.MinX || bChild.MaxX > bParent.MaxX ||
		bChild.MinY < bParent.MinY || bChild.MaxY > bParent.MaxY {
		return false
	}

	// Test multiple points from child to ensure confidence
	insideCount := 0
	tests := 0
	step := len(child) / 5
	if step < 1 {
		step = 1
	}
	for i := 0; i < len(child); i += step {
		// Pick point slightly offset towards centroid or next vertex
		p := child[i]
		tests++
		if PointInPolygon(p, parent) {
			insideCount++
		}
	}

	return insideCount > (tests / 2)
}

// CleanContour applies strict vertex deduplication, zero-length segment removal,
// and collinear edge simplification.
func CleanContour(pts []Point, tol float64) []Point {
	if len(pts) < 3 {
		return pts
	}
	if tol <= 0 {
		tol = 1e-5
	}

	// 1. Remove consecutive duplicate points
	var dedup []Point
	for i := 0; i < len(pts); i++ {
		p := pts[i]
		if len(dedup) == 0 {
			dedup = append(dedup, p)
			continue
		}
		prev := dedup[len(dedup)-1]
		distSq := (p.X-prev.X)*(p.X-prev.X) + (p.Y-prev.Y)*(p.Y-prev.Y)
		if distSq > tol*tol {
			dedup = append(dedup, p)
		}
	}

	// Also check if last point is duplicate of first point
	if len(dedup) > 2 {
		first := dedup[0]
		last := dedup[len(dedup)-1]
		distSq := (last.X-first.X)*(last.X-first.X) + (last.Y-first.Y)*(last.Y-first.Y)
		if distSq <= tol*tol {
			dedup = dedup[:len(dedup)-1]
		}
	}

	if len(dedup) < 3 {
		return dedup
	}

	// 2. Remove collinear vertices
	var simplified []Point
	n := len(dedup)
	for i := 0; i < n; i++ {
		pPrev := dedup[(i+n-1)%n]
		pCurr := dedup[i]
		pNext := dedup[(i+1)%n]

		// Perpendicular distance from pCurr to line (pPrev -> pNext)
		dx := pNext.X - pPrev.X
		dy := pNext.Y - pPrev.Y
		lenSq := dx*dx + dy*dy
		if lenSq <= tol*tol {
			// pPrev and pNext are identical; pCurr is kept or dropped
			continue
		}

		// Cross product (twice the triangle area)
		cross := math.Abs((pCurr.X-pPrev.X)*dy - (pCurr.Y-pPrev.Y)*dx)
		dist := cross / math.Sqrt(lenSq)

		if dist > tol {
			simplified = append(simplified, pCurr)
		}
	}

	if len(simplified) < 3 {
		return dedup
	}
	return simplified
}
