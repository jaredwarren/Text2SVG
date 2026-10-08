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
	fmt.Printf("Contour 0: %d pts, area=%.2f, CW=%v\n", len(c0), converter.SignedArea(c0), converter.IsClockwise(c0))

	// Step 1: Raw offset of c0
	d := 8.0
	miterLimit := 3.0
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
			d1x, d1y := -nPrev.Y, nPrev.X
			d2x, d2y := -nCurr.Y, nCurr.X
			denom := d1x*d2y - d1y*d2x
			if math.Abs(denom) > 1e-7 {
				u := ((pCurrOffset.X-pPrevOffset.X)*d2y - (pCurrOffset.Y-pPrevOffset.Y)*d2x) / denom
				inter := converter.Point{X: pPrevOffset.X + u*d1x, Y: pPrevOffset.Y + u*d1y}
				distSq := (inter.X-p.X)*(inter.X-p.X) + (inter.Y-p.Y)*(inter.Y-p.Y)
				if distSq <= (miterLimit*d)*(miterLimit*d) {
					rawOffset = append(rawOffset, inter)
				} else {
					rawOffset = append(rawOffset, pPrevOffset, pCurrOffset)
				}
			} else {
				rawOffset = append(rawOffset, pPrevOffset, pCurrOffset)
			}
		} else {
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
	fmt.Printf("cleanedRaw: %d pts, area=%.2f, CW=%v\n", len(cleanedRaw), converter.SignedArea(cleanedRaw), converter.IsClockwise(cleanedRaw))

	// Save raw offset
	sb := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-20 -40 80 80" width="800" height="800">
<path d="%s" fill="none" stroke="red" stroke-width="0.3"/>
<path d="%s" fill="none" stroke="blue" stroke-width="0.3"/>
</svg>`, pathToD(c0), pathToD(cleanedRaw))
	_ = os.WriteFile("raw_from_c0.svg", []byte(sb), 0644)

	// Step 2: Now trace UntangleContour on cleanedRaw!
	cur := cleanedRaw
	for iter := 0; iter < 10; iter++ {
		nCur := len(cur)
		found := false
		for i := 0; i < nCur; i++ {
			p1 := cur[i]
			p2 := cur[(i+1)%nCur]
			for j := i + 2; j < nCur; j++ {
				if (j+1)%nCur == i {
					continue
				}
				p3 := cur[j]
				p4 := cur[(j+1)%nCur]
				inter, ok := converter.SegmentsIntersect(p1, p2, p3, p4)
				if ok {
					loopA := append([]converter.Point{}, cur[:i+1]...)
					loopA = append(loopA, inter)
					if j+1 < nCur {
						loopA = append(loopA, cur[j+1:]...)
					}

					loopB := []converter.Point{inter}
					loopB = append(loopB, cur[i+1:j+1]...)

					saA := converter.SignedArea(loopA)
					saB := converter.SignedArea(loopB)

					fmt.Printf("Iter %d: split seg %d & %d (%d pts vs %d pts). AreaA=%.2f (CW=%v), AreaB=%.2f (CW=%v)\n",
						iter, i, j, len(loopA), len(loopB), saA, saA < 0, saB, saB < 0)

					if math.Abs(saA) >= math.Abs(saB) {
						cur = loopA
					} else {
						cur = loopB
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			fmt.Printf("Untangle finished after %d iters. Final pts: %d\n", iter, len(cur))
			break
		}
	}
}

func pathToD(pts []converter.Point) string {
	var s string
	for i, p := range pts {
		if i == 0 {
			s += fmt.Sprintf("M %.2f,%.2f ", p.X, -p.Y)
		} else {
			s += fmt.Sprintf("L %.2f,%.2f ", p.X, -p.Y)
		}
	}
	s += "Z"
	return s
}
