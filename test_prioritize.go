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
		pCurr := cwPts[i]
		pNext := cwPts[(i+1)%n]
		dx := pNext.X - pCurr.X
		dy := pNext.Y - pCurr.Y
		lenSeg := math.Hypot(dx, dy)
		if lenSeg < 1e-9 {
			normals = append(normals, converter.Point{X: 0, Y: 0})
		} else {
			normals = append(normals, converter.Point{X: dy / lenSeg, Y: -dx / lenSeg})
		}
	}

	var rawOffset []converter.Point
	for i := 0; i < n; i++ {
		p := cwPts[i]
		nPrev := normals[(i-1+n)%n]
		nCurr := normals[i]
		cross := nPrev.X*nCurr.Y - nPrev.Y*nCurr.X
		dot := nPrev.X*nCurr.X + nPrev.Y*nCurr.Y

		if math.Abs(cross) < 1e-6 && dot > 0 {
			rawOffset = append(rawOffset, converter.Point{X: p.X + nCurr.X*d, Y: p.Y + nCurr.Y*d})
			continue
		}

		isConvex := cross < 0
		if !isConvex {
			dPrev := converter.Point{X: -nPrev.Y, Y: nPrev.X}
			dCurr := converter.Point{X: -nCurr.Y, Y: nCurr.X}
			denom := dPrev.X*dCurr.Y - dPrev.Y*dCurr.X
			if math.Abs(denom) > 1e-6 {
				pPrevOffset := converter.Point{X: p.X + nPrev.X*d, Y: p.Y + nPrev.Y*d}
				pCurrOffset := converter.Point{X: p.X + nCurr.X*d, Y: p.Y + nCurr.Y*d}
				u := ((pCurrOffset.X-pPrevOffset.X)*dCurr.Y - (pCurrOffset.Y-pPrevOffset.Y)*dCurr.X) / denom
				inter := converter.Point{X: pPrevOffset.X + u*dPrev.X, Y: pPrevOffset.Y + u*dPrev.Y}
				distSq := (inter.X-p.X)*(inter.X-p.X) + (inter.Y-p.Y)*(inter.Y-p.Y)
				if distSq <= (miterLimit*d)*(miterLimit*d) {
					rawOffset = append(rawOffset, inter)
				} else {
					rawOffset = append(rawOffset, pPrevOffset, pCurrOffset)
				}
			} else {
				rawOffset = append(rawOffset, converter.Point{X: p.X + nCurr.X*d, Y: p.Y + nCurr.Y*d})
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
	fmt.Printf("Initial cleanedRaw points: %d\n", len(cleanedRaw))

	// Untangle prioritizing the intersection with the largest pruned loop
	cur := cleanedRaw
	for iter := 0; iter < 100; iter++ {
		nCur := len(cur)
		if nCur < 4 {
			break
		}

		bestI, bestJ := -1, -1
		var bestInter converter.Point
		var bestRemaining []converter.Point
		maxPrunedPoints := 0

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
					// Loop A: cur[0..i] + inter + cur[j+1..n-1]
					loopA := append([]converter.Point{}, cur[:i+1]...)
					loopA = append(loopA, inter)
					if j+1 < nCur {
						loopA = append(loopA, cur[j+1:]...)
					}

					// Loop B: inter + cur[i+1..j]
					loopB := []converter.Point{inter}
					loopB = append(loopB, cur[i+1:j+1]...)

					areaA := math.Abs(converter.SignedArea(loopA))
					areaB := math.Abs(converter.SignedArea(loopB))

					var remaining []converter.Point
					prunedPoints := 0
					if areaA >= areaB {
						remaining = loopA
						prunedPoints = len(loopB)
					} else {
						remaining = loopB
						prunedPoints = len(loopA)
					}

					if prunedPoints > maxPrunedPoints {
						maxPrunedPoints = prunedPoints
						bestI, bestJ = i, j
						bestInter = inter
						bestRemaining = remaining
					}
				}
			}
		}

		if maxPrunedPoints == 0 {
			fmt.Printf("Successfully untangled after %d iterations! Final pts=%d\n", iter, len(cur))
			break
		}

		fmt.Printf("Iter %d: pruned %d points at segs %d & %d (inter: %.2f, %.2f)\n",
			iter, maxPrunedPoints, bestI, bestJ, bestInter.X, bestInter.Y)
		cur = converter.CleanContour(bestRemaining, 1e-4)
	}

	// Render the result
	_ = os.WriteFile("test_best_untangle_6.svg", []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-20 -20 60 60" width="800" height="800">
<path d="%s" fill="none" stroke="red" stroke-width="0.3"/>
<path d="%s" fill="none" stroke="cyan" stroke-width="0.5"/>
</svg>`, pathToD(cwPts), pathToD(cur))), 0644)
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
