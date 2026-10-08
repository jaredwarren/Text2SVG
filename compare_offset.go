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

	// Call OffsetContour
	out1 := converter.OffsetContour(cwPts, 8.0, converter.JoinRound, 3.0)
	fmt.Printf("OffsetContour result: %d pts, area=%.2f\n", len(out1), converter.SignedArea(out1))
}
