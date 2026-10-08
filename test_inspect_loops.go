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

	cur := converter.CleanContour(rawOffset, 1e-4)

	// Split at seg 188 & 861
	i, j := 188, 861
	inter, ok := converter.SegmentsIntersect(cur[i], cur[i+1], cur[j], cur[j+1])
	if !ok {
		panic("expected intersection")
	}

	loopA := append([]converter.Point{}, cur[:i+1]...)
	loopA = append(loopA, inter)
	loopA = append(loopA, cur[j+1:]...)

	loopB := []converter.Point{inter}
	loopB = append(loopB, cur[i+1:j+1]...)

	loopA = converter.CleanContour(loopA, 1e-4)
	loopB = converter.CleanContour(loopB, 1e-4)

	fmt.Printf("loopA: pts=%d, area=%.2f, CW=%v, bb=[%.2f, %.2f] to [%.2f, %.2f]\n",
		len(loopA), converter.SignedArea(loopA), converter.IsClockwise(loopA),
		converter.ContourBoundingBox(loopA).MinX, converter.ContourBoundingBox(loopA).MinY,
		converter.ContourBoundingBox(loopA).MaxX, converter.ContourBoundingBox(loopA).MaxY)

	fmt.Printf("loopB: pts=%d, area=%.2f, CW=%v, bb=[%.2f, %.2f] to [%.2f, %.2f]\n",
		len(loopB), converter.SignedArea(loopB), converter.IsClockwise(loopB),
		converter.ContourBoundingBox(loopB).MinX, converter.ContourBoundingBox(loopB).MinY,
		converter.ContourBoundingBox(loopB).MaxX, converter.ContourBoundingBox(loopB).MaxY)

	sb := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-20 -40 80 80" width="800" height="800">
<path d="%s" fill="none" stroke="red" stroke-width="0.3"/>
<path d="%s" fill="none" stroke="orange" stroke-width="0.4"/>
<path d="%s" fill="none" stroke="cyan" stroke-width="0.4"/>
</svg>`, pathToD(cwPts), pathToD(loopA), pathToD(loopB))
	_ = os.WriteFile("test_inspect_loops.svg", []byte(sb), 0644)
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
