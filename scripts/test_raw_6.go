//go:build ignore

package main

import (
	"fmt"
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
	offsetPts := converter.OffsetContour(cwPts, 8.0, converter.JoinRound, 3.0)

	// Save original contour and offset contour as separate SVG paths
	var sb = fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-20 -20 60 60" width="800" height="800">
<path d="%s" fill="none" stroke="red" stroke-width="0.2"/>
<path d="%s" fill="none" stroke="blue" stroke-width="0.2"/>
</svg>`, pathToD(cwPts), pathToD(offsetPts))
	_ = os.WriteFile("raw_offset_6.svg", []byte(sb), 0644)

	// Let's check self-intersections in offsetPts
	n := len(offsetPts)
	for i := 0; i < n; i++ {
		p1 := offsetPts[i]
		p2 := offsetPts[(i+1)%n]
		for j := i + 2; j < n; j++ {
			if (j+1)%n == i {
				continue
			}
			p3 := offsetPts[j]
			p4 := offsetPts[(j+1)%n]
			inter, ok := converter.SegmentsIntersect(p1, p2, p3, p4)
			if ok {
				fmt.Printf("Self-intersection between seg %d and seg %d at (%.2f, %.2f)\n", i, j, inter.X, inter.Y)
			}
		}
	}
}

func pathToD(pts []converter.Point) string {
	var s string
	for i, p := range pts {
		// Flip Y for SVG: svgY = -p.Y
		if i == 0 {
			s += fmt.Sprintf("M %.2f,%.2f ", p.X, -p.Y)
		} else {
			s += fmt.Sprintf("L %.2f,%.2f ", p.X, -p.Y)
		}
	}
	s += "Z"
	return s
}
