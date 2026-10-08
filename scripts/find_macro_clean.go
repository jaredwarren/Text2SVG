//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
)

func main() {
	data, _ := os.ReadFile("/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf")
	lf, _ := converter.ParseFont(data, "Academy")

	res, _ := lf.LayoutText(converter.TextParams{
		Text:              "6",
		Size:              30.0,
		Units:             converter.UnitsMM,
		Weld:              true,
		BoundaryMode:      converter.BoundaryNone,
		Datum:             converter.DatumBottomLeft,
	})

	c0 := res.Contours[0].Points
	d := 8.0
	n := len(c0)

	var normals []converter.Point
	for i := 0; i < n; i++ {
		p0 := c0[i]
		p1 := c0[(i+1)%n]
		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		lenSeg := math.Hypot(dx, dy)
		if lenSeg < 1e-9 {
			lenSeg = 1e-9
		}
		normals = append(normals, converter.Point{X: -dy / lenSeg, Y: dx / lenSeg})
	}

	var rawOffset []converter.Point
	for i := 0; i < n; i++ {
		p := c0[i]
		nPrev := normals[(i-1+n)%n]
		nCurr := normals[i]
		cross := nPrev.X*nCurr.Y - nPrev.Y*nCurr.X

		pPrevOffset := converter.Point{X: p.X + d*nPrev.X, Y: p.Y + d*nPrev.Y}
		pCurrOffset := converter.Point{X: p.X + d*nCurr.X, Y: p.Y + d*nCurr.Y}

		isOpeningCorner := (cross * d) <= 0
		if math.Abs(cross) < 1e-5 {
			rawOffset = append(rawOffset, pCurrOffset)
			continue
		}

		if !isOpeningCorner {
			// On concave corner:
			rawOffset = append(rawOffset, pPrevOffset, pCurrOffset)
		} else {
			// Convex corner: round arc
			anglePrev := math.Atan2(nPrev.Y, nPrev.X)
			angleCurr := math.Atan2(nCurr.Y, nCurr.X)
			sweep := angleCurr - anglePrev
			for sweep > 0 {
				sweep -= 2 * math.Pi
			}
			for sweep < -2*math.Pi {
				sweep += 2 * math.Pi
			}
			if sweep < -math.Pi {
				sweep += 2 * math.Pi
			}
			steps := int(math.Ceil(math.Abs(sweep) / (math.Pi / 6.0)))
			if steps < 1 {
				steps = 1
			}
			for s := 0; s <= steps; s++ {
				t := float64(s) / float64(steps)
				a := anglePrev + t*sweep
				rawOffset = append(rawOffset, converter.Point{
					X: p.X + d*math.Cos(a),
					Y: p.Y + d*math.Sin(a),
				})
			}
		}
	}

	cleanedRaw := converter.CleanContour(rawOffset, 1e-4)
	fmt.Printf("Cleaned raw points: %d\n", len(cleanedRaw))

	// Find all intersections
	nCur := len(cleanedRaw)
	for i := 0; i < nCur; i++ {
		p1 := cleanedRaw[i]
		p2 := cleanedRaw[(i+1)%nCur]
		for j := i + 5; j < nCur-2; j++ {
			p3 := cleanedRaw[j]
			p4 := cleanedRaw[(j+1)%nCur]
			inter, ok := converter.SegmentsIntersect(p1, p2, p3, p4)
			if ok {
				loopA := append([]converter.Point{}, cleanedRaw[:i+1]...)
				loopA = append(loopA, inter)
				if j+1 < nCur {
					loopA = append(loopA, cleanedRaw[j+1:]...)
				}
				loopB := []converter.Point{inter}
				loopB = append(loopB, cleanedRaw[i+1:j+1]...)

				saA := converter.SignedArea(loopA)
				saB := converter.SignedArea(loopB)
				fmt.Printf("Macro intersection seg %d x %d: loopA (pts=%d, sa=%.2f, CW=%v), loopB (pts=%d, sa=%.2f, CW=%v)\n",
					i, j, len(loopA), saA, saA < 0, len(loopB), saB, saB < 0)
			}
		}
	}
}
