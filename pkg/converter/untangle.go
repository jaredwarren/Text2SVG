package converter

import (
	"math"
)

// SegmentsIntersect checks if line segment p1->p2 intersects segment p3->p4.
// If they intersect strictly in their interiors, it returns the intersection point and true.
func SegmentsIntersect(p1, p2, p3, p4 Point) (Point, bool) {
	d := (p2.X-p1.X)*(p4.Y-p3.Y) - (p2.Y-p1.Y)*(p4.X-p3.X)
	if math.Abs(d) < 1e-9 {
		return Point{}, false
	}
	u := ((p3.X-p1.X)*(p4.Y-p3.Y) - (p3.Y-p1.Y)*(p4.X-p3.X)) / d
	v := ((p3.X-p1.X)*(p2.Y-p1.Y) - (p3.Y-p1.Y)*(p2.X-p1.X)) / d
	if u > 1e-5 && u < 1.0-1e-5 && v > 1e-5 && v < 1.0-1e-5 {
		return Point{
			X: p1.X + u*(p2.X-p1.X),
			Y: p1.Y + u*(p2.Y-p1.Y),
		}, true
	}
	return Point{}, false
}

// UntangleContour removes self-intersecting swallowtail loops from an offset contour,
// keeping the primary outer boundary loop.
func UntangleContour(pts []Point) []Point {
	cur := pts
	maxIterations := 100
	changed := true

	for changed && maxIterations > 0 {
		changed = false
		maxIterations--
		n := len(cur)
		if n < 4 {
			break
		}

		for i := 0; i < n; i++ {
			p1 := cur[i]
			p2 := cur[(i+1)%n]

			for j := i + 2; j < n; j++ {
				// Don't check adjacent wrap-around segment
				if (j+1)%n == i {
					continue
				}
				p3 := cur[j]
				p4 := cur[(j+1)%n]

				inter, ok := SegmentsIntersect(p1, p2, p3, p4)
				if ok {
					// Split into loopA and loopB
					// loopA: cur[0..i] + inter + cur[j+1..n-1]
					loopA := make([]Point, 0, i+1+(n-j))
					loopA = append(loopA, cur[:i+1]...)
					loopA = append(loopA, inter)
					if j+1 < n {
						loopA = append(loopA, cur[j+1:]...)
					}

					// loopB: inter + cur[i+1..j]
					loopB := make([]Point, 0, (j-i)+1)
					loopB = append(loopB, inter)
					loopB = append(loopB, cur[i+1:j+1]...)

					areaA := math.Abs(SignedArea(loopA))
					areaB := math.Abs(SignedArea(loopB))

					if areaA >= areaB {
						cur = loopA
					} else {
						cur = loopB
					}
					changed = true
					break
				}
			}
			if changed {
				break
			}
		}
	}

	return CleanContour(cur, 1e-4)
}
