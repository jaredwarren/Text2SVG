//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
)

func main() {
	data, err := os.ReadFile("/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf")
	if err != nil {
		panic(err)
	}
	lf, err := converter.ParseFont(data, "Academy")
	if err != nil {
		panic(err)
	}

	res, err := lf.LayoutText(converter.TextParams{
		Text:              "LASER CUT 2026",
		Size:              30.0,
		Units:             converter.UnitsMM,
		Weld:              true,
		BoundaryMode:      converter.BoundaryConformal,
		BoundaryOffset:    8.0,
		BoundaryFillHoles: true,
		Datum:             converter.DatumBottomLeft,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Number of contours: %d\n", len(res.Contours))
	fmt.Printf("Number of boundary contours: %d\n", len(res.BoundaryContours))

	for i, c := range res.Contours {
		bb := converter.ContourBoundingBox(c.Points)
		if bb.MinX > 180 { // Numbers 2026
			fmt.Printf("Contour %d (2026): bounds [%.2f, %.2f] to [%.2f, %.2f], pts=%d, area=%.2f\n",
				i, bb.MinX, bb.MinY, bb.MaxX, bb.MaxY, len(c.Points), converter.SignedArea(c.Points))
		}
	}

	for i, bc := range res.BoundaryContours {
		bb := converter.ContourBoundingBox(bc.Points)
		fmt.Printf("Boundary %d: bounds [%.2f, %.2f] to [%.2f, %.2f], pts=%d\n", i, bb.MinX, bb.MinY, bb.MaxX, bb.MaxY, len(bc.Points))
	}
	_ = os.WriteFile("test_investigate.svg", []byte(res.SVG), 0644)
}
