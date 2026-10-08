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

	res, _ := lf.LayoutText(converter.TextParams{
		Text:              "6",
		Size:              30.0,
		Units:             converter.UnitsMM,
		Weld:              true,
		BoundaryMode:      converter.BoundaryConformal,
		BoundaryOffset:    8.0,
		BoundaryFillHoles: true,
		Datum:             converter.DatumCenter,
	})

	_ = os.WriteFile("test_digit_6.svg", []byte(res.SVG), 0644)
	fmt.Printf("Boundary contours count for '6': %d\n", len(res.BoundaryContours))
	for i, bc := range res.BoundaryContours {
		bb := converter.ContourBoundingBox(bc.Points)
		fmt.Printf("BC %d: %d pts, bounds: [%.2f, %.2f] to [%.2f, %.2f], area=%.2f\n",
			i, len(bc.Points), bb.MinX, bb.MinY, bb.MaxX, bb.MaxY, converter.SignedArea(bc.Points))
	}
}
