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

	// Step 1: Raw offset vertices without UntangleContour
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
			rawOffset = append(rawOffset, converter.Point{
				X: p.X + nCurr.X*d,
				Y: p.Y + nCurr.Y*d,
			})
			continue
		}

		isConvex := cross < 0
		if !isConvex {
			// Concave corner: line intersection
			dPrev := converter.Point{X: -nPrev.Y, Y: nPrev.X}
			dCurr := converter.Point{X: -nCurr.Y, Y: nCurr.X}
			denom := dPrev.X*dCurr.Y - dPrev.Y*dCurr.X
			if math.Abs(denom) > 1e-6 {
				pPrevOffset := converter.Point{X: p.X + nPrev.X*d, Y: p.Y + nPrev.Y*d}
				pCurrOffset := converter.Point{X: p.X + nCurr.X*d, Y: p.Y + nCurr.Y*d}
				u := ((pCurrOffset.X-pPrevOffset.X)*dCurr.Y - (pCurrOffset.Y-pPrevOffset.Y)*dCurr.X) / denom
				inter := converter.Point{
					X: pPrevOffset.X + u*dPrev.X,
					Y: pPrevOffset.Y + u*dPrev.Y,
				}
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
	fmt.Printf("cleanedRaw points: %d\n", len(cleanedRaw))

	// Find self-intersections in cleanedRaw
	nRaw := len(cleanedRaw)
	interCount := 0
	for i := 0; i < nRaw; i++ {
		p1 := cleanedRaw[i]
		p2 := cleanedRaw[(i+1)%nRaw]
		for j := i + 2; j < nRaw; j++ {
			if (j+1)%nRaw == i {
				continue
			}
			p3 := cleanedRaw[j]
			p4 := cleanedRaw[(j+1)%nRaw]
			inter, ok := converter.SegmentsIntersect(p1, p2, p3, p4)
			if ok {
				interCount++
				fmt.Printf("Intersection %d: between seg %d and seg %d at (%.2f, %.2f)\n", interCount, i, j, inter.X, inter.Y)
			}
		}
	}

	// Now trace UntangleContour step-by-step
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

					areaA := math.Abs(converter.SignedArea(loopA))
					areaB := math.Abs(converter.SignedArea(loopB))

					fmt.Printf("Iter %d: split at seg %d & %d. AreaA=%.2f (pts=%d), AreaB=%.2f (pts=%d)\n",
						iter, i, j, areaA, len(loopA), areaB, len(loopB))

					if areaA >= areaB {
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
			fmt.Printf("No more intersections after iter %d!\n", iter)
			break
		}
	}
}
