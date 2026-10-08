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
	g, _ := lf.LoadGlyph('6', 20)

	scale := 30.0 / float64(lf.UnitsPerEm)
	var pts []converter.Point
	for _, p := range g.Contours[0].Points {
		pts = append(pts, converter.Point{X: p.X * scale, Y: p.Y * scale})
	}

	cwPts := converter.EnsureOrientation(pts, true)
	d := 8.0
	miterLimit := 3.0
	n := len(cwPts)

	var normals []converter.Point
	for i := 0; i < n; i++ {
		p0 := cwPts[i]
		p1 := cwPts[(i+1)%n]
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
		p := cwPts[i]
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
	nRaw := len(cleanedRaw)

	type InterInfo struct {
		i, j       int
		pt         converter.Point
		loopASize  int
		loopBSize  int
		loopAArea  float64
		loopBArea  float64
	}

	var inters []InterInfo
	for i := 0; i < nRaw; i++ {
		p1 := cleanedRaw[i]
		p2 := cleanedRaw[(i+1)%nRaw]
		for j := i + 5; j < nRaw-2; j++ {
			p3 := cleanedRaw[j]
			p4 := cleanedRaw[(j+1)%nRaw]
			inter, ok := converter.SegmentsIntersect(p1, p2, p3, p4)
			if ok {
				loopA := append([]converter.Point{}, cleanedRaw[:i+1]...)
				loopA = append(loopA, inter)
				if j+1 < nRaw {
					loopA = append(loopA, cleanedRaw[j+1:]...)
				}

				loopB := []converter.Point{inter}
				loopB = append(loopB, cleanedRaw[i+1:j+1]...)

				saA := converter.SignedArea(loopA)
				saB := converter.SignedArea(loopB)

				inters = append(inters, InterInfo{
					i: i, j: j, pt: inter,
					loopASize: len(loopA), loopBSize: len(loopB),
					loopAArea: saA, loopBArea: saB,
				})
			}
		}
	}

	fmt.Printf("Total macro intersections (|j-i| >= 5): %d\n", len(inters))
	for idx, ii := range inters {
		if idx < 30 || idx >= len(inters)-10 {
			fmt.Printf("#%d: seg %d x %d at (%.2f, %.2f) | loopA: pts=%d, area=%.2f | loopB: pts=%d, area=%.2f\n",
				idx, ii.i, ii.j, ii.pt.X, ii.pt.Y, ii.loopASize, ii.loopAArea, ii.loopBSize, ii.loopBArea)
		}
	}
}
