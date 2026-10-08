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

	fmt.Printf("Glyph '6' has %d contours:\n", len(g.Contours))
	for i, c := range g.Contours {
		sa := converter.SignedArea(c.Points)
		bb := converter.ContourBoundingBox(c.Points)
		fmt.Printf("Contour %d: %d pts, area=%.2f, CW=%v, bounds: [%.2f, %.2f] to [%.2f, %.2f]\n",
			i, len(c.Points), sa, sa < 0, bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
	}

	scale := 30.0 / float64(lf.UnitsPerEm)
	var scaled []converter.Contour
	for _, c := range g.Contours {
		var pts []converter.Point
		for _, p := range c.Points {
			pts = append(pts, converter.Point{X: p.X * scale, Y: p.Y * scale})
		}
		scaled = append(scaled, converter.Contour{Points: pts, Closed: true})
	}

	classified := converter.ClassifyAndOrientContours(scaled, 1e-4)
	fmt.Printf("\nClassified:\n")
	for i, cc := range classified {
		fmt.Printf("CC %d: role=%s, depth=%d, area=%.2f\n", i, cc.Role, cc.Depth, cc.Area)
	}

	// Now check OffsetContour on CC 0 (outer)
	c0 := classified[0]
	cwPts := converter.EnsureOrientation(c0.Points, true)
	offsetPts := converter.OffsetContour(cwPts, 8.0, converter.JoinRound, 3.0)
	fmt.Printf("\nOffsetPts: %d pts\n", len(offsetPts))

	// Check untangling
	cleaned := converter.CleanContour(offsetPts, 1e-4)
	fmt.Printf("Cleaned: %d pts\n", len(cleaned))
	untangled := converter.UntangleContour(cleaned)
	fmt.Printf("Untangled: %d pts, bounds: [%.2f, %.2f] to [%.2f, %.2f], area=%.2f\n",
		len(untangled),
		converter.ContourBoundingBox(untangled).MinX, converter.ContourBoundingBox(untangled).MinY,
		converter.ContourBoundingBox(untangled).MaxX, converter.ContourBoundingBox(untangled).MaxY,
		converter.SignedArea(untangled))
}
