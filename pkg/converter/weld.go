package converter

import (
	"github.com/akavel/polyclip-go"
)

// GlyphContourGroup represents all contours belonging to a single glyph at its positioned layout.
type GlyphContourGroup struct {
	Contours []Contour
}

// WeldGlyphs performs polygon boolean union across all overlapping glyph contours,
// merging intersecting outer boundaries while preserving all interior counter holes.
// It returns a topologically clean, oriented, and manifold set of contours.
func WeldGlyphs(groups []GlyphContourGroup, tol float64) ([]Contour, []ClassifiedContour) {
	if len(groups) == 0 {
		return nil, nil
	}
	if tol <= 0 {
		tol = 1e-4
	}

	var polys []polyclip.Polygon

	for _, g := range groups {
		if len(g.Contours) == 0 {
			continue
		}

		// Classify within the glyph so holes are correctly CCW and outer is CW
		classified := ClassifyAndOrientContours(g.Contours, tol)
		var p polyclip.Polygon

		for _, cc := range classified {
			if len(cc.Points) < 3 {
				continue
			}
			var pc polyclip.Contour
			for _, pt := range cc.Points {
				pc.Add(polyclip.Point{X: pt.X, Y: pt.Y})
			}
			p.Add(pc)
		}

		if len(p) > 0 {
			polys = append(polys, p)
		}
	}

	if len(polys) == 0 {
		return nil, nil
	}

	// Pairwise divide-and-conquer union
	for len(polys) > 1 {
		var nextRound []polyclip.Polygon
		for i := 0; i < len(polys); i += 2 {
			if i+1 < len(polys) {
				// Union the pair
				merged := polys[i].Construct(polyclip.UNION, polys[i+1])
				nextRound = append(nextRound, merged)
			} else {
				nextRound = append(nextRound, polys[i])
			}
		}
		polys = nextRound
	}

	resultPoly := polys[0]
	var rawContours []Contour

	for _, pc := range resultPoly {
		if len(pc) < 3 {
			continue
		}
		pts := make([]Point, len(pc))
		for i, pt := range pc {
			pts[i] = Point{X: pt.X, Y: pt.Y}
		}
		rawContours = append(rawContours, Contour{
			Points: pts,
			Closed: true,
		})
	}

	// Final manifold classification and orientation tagging (CW outer, CCW hole)
	classifiedResult := ClassifyAndOrientContours(rawContours, tol)

	finalContours := make([]Contour, len(classifiedResult))
	for i, c := range classifiedResult {
		finalContours[i] = Contour{
			Points: c.Points,
			Closed: true,
		}
	}

	return finalContours, classifiedResult
}
