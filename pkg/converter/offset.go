package converter

import (
	"math"
)

// OffsetContour offsets a single closed contour by distance d.
// If makeCW is true, outward normal is (-dy/L, dx/L).
func OffsetContour(pts []Point, d float64, join CornerJoin, miterLimit float64) []Point {
	pts = CleanContour(pts, 1e-5)
	n := len(pts)
	if n < 3 || math.Abs(d) < 1e-6 {
		return pts
	}
	if miterLimit <= 1.0 {
		miterLimit = 3.0
	}
	if join == "" {
		join = JoinRound
	}

	wasCCW := !IsClockwise(pts)
	if wasCCW {
		pts = EnsureOrientation(pts, true)
	}

	// Compute edge directions and outward normals (strictly for CW contour)
	normals := make([]Point, n)
	for i := 0; i < n; i++ {
		p0 := pts[i]
		p1 := pts[(i+1)%n]
		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		len := math.Sqrt(dx*dx + dy*dy)
		if len < 1e-9 {
			len = 1e-9
		}
		// Normal pointing outward for CW contour
		normals[i] = Point{
			X: -dy / len,
			Y: dx / len,
		}
	}

	var out []Point

	for i := 0; i < n; i++ {
		prevIdx := (i + n - 1) % n
		nPrev := normals[prevIdx]
		nCurr := normals[i]
		p := pts[i]

		cross := nPrev.X*nCurr.Y - nPrev.Y*nCurr.X

		pPrevOffset := Point{X: p.X + d*nPrev.X, Y: p.Y + d*nPrev.Y}
		pCurrOffset := Point{X: p.X + d*nCurr.X, Y: p.Y + d*nCurr.Y}

		// When turn direction matches sign of d, corner opens outward (convex turn)
		isOpeningCorner := (cross * d) <= 0

		if math.Abs(cross) < 1e-5 {
			// Nearly parallel / collinear
			out = append(out, pCurrOffset)
			continue
		}

		if !isOpeningCorner {
			// Closing / concave corner: find intersection of the two offset lines
			// Line 1: through pPrevOffset with dir (-nPrev.Y, nPrev.X)
			// Line 2: through pCurrOffset with dir (-nCurr.Y, nCurr.X)
			d1x, d1y := -nPrev.Y, nPrev.X
			d2x, d2y := -nCurr.Y, nCurr.X
			denom := d1x*d2y - d1y*d2x
			if math.Abs(denom) > 1e-7 {
				u := ((pCurrOffset.X-pPrevOffset.X)*d2y - (pCurrOffset.Y-pPrevOffset.Y)*d2x) / denom
				inter := Point{X: pPrevOffset.X + u*d1x, Y: pPrevOffset.Y + u*d1y}
				out = append(out, inter)
			} else {
				out = append(out, pPrevOffset, pCurrOffset)
			}
		} else {
			// Opening / convex corner
			switch join {
			case JoinMiter:
				d1x, d1y := -nPrev.Y, nPrev.X
				d2x, d2y := -nCurr.Y, nCurr.X
				denom := d1x*d2y - d1y*d2x
				if math.Abs(denom) > 1e-7 {
					u := ((pCurrOffset.X-pPrevOffset.X)*d2y - (pCurrOffset.Y-pPrevOffset.Y)*d2x) / denom
					inter := Point{X: pPrevOffset.X + u*d1x, Y: pPrevOffset.Y + u*d1y}
					distSq := (inter.X-p.X)*(inter.X-p.X) + (inter.Y-p.Y)*(inter.Y-p.Y)
					if distSq <= (miterLimit*d)*(miterLimit*d) {
						out = append(out, inter)
					} else {
						// Bevel fallback if miter exceeds limit
						out = append(out, pPrevOffset, pCurrOffset)
					}
				} else {
					out = append(out, pPrevOffset, pCurrOffset)
				}

			case JoinBevel:
				out = append(out, pPrevOffset, pCurrOffset)

			case JoinRound:
				fallthrough
			default:
				// Arc interpolation between pPrevOffset and pCurrOffset
				anglePrev := math.Atan2(nPrev.Y, nPrev.X)
				angleCurr := math.Atan2(nCurr.Y, nCurr.X)
				sweep := angleCurr - anglePrev
				if d > 0 {
					for sweep < 0 {
						sweep += 2 * math.Pi
					}
				} else {
					for sweep > 0 {
						sweep -= 2 * math.Pi
					}
				}

				steps := int(math.Ceil(math.Abs(sweep) / (math.Pi / 6.0)))
				if steps < 1 {
					steps = 1
				}
				for s := 0; s <= steps; s++ {
					t := float64(s) / float64(steps)
					a := anglePrev + t*sweep
					out = append(out, Point{
						X: p.X + math.Abs(d)*math.Cos(a)*math.Copysign(1.0, d),
						Y: p.Y + math.Abs(d)*math.Sin(a)*math.Copysign(1.0, d),
					})
				}
			}
		}
	}

	cleaned := CleanContour(out, 1e-4)
	if wasCCW {
		cleaned = EnsureOrientation(cleaned, false)
	}
	return cleaned
}

// OffsetProfile applies contour offsetting to a list of classified manifold contours.
// - Outer boundaries are offset by distance d (positive expands outer size).
// - Inner holes are offset by -d (positive expands text solid, so hole contracts).
func OffsetProfile(classified []ClassifiedContour, d float64, join CornerJoin, miterLimit float64) []Contour {
	if math.Abs(d) < 1e-5 || len(classified) == 0 {
		res := make([]Contour, len(classified))
		for i, c := range classified {
			res[i] = Contour{Points: c.Points, Closed: true}
		}
		return res
	}

	var offsetContours []Contour

	for _, c := range classified {
		var effD float64
		if c.Role == RoleOuter {
			effD = d
		} else {
			// For holes, positive d shrinks the hole (material expands into hole)
			effD = -d
		}

		offsetPts := OffsetContour(c.Points, effD, join, miterLimit)
		if len(offsetPts) < 3 {
			continue
		}

		// Check if offset area collapsed to zero (e.g. over-inset)
		area := math.Abs(SignedArea(offsetPts))
		if area < 1e-4 {
			continue
		}

		offsetContours = append(offsetContours, Contour{
			Points: offsetPts,
			Closed: true,
		})
	}

	return offsetContours
}
