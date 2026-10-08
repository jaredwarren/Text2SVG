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
		BoundaryMode:      converter.BoundaryNone,
		Datum:             converter.DatumBottomLeft,
	})

	fmt.Printf("LayoutText('6') produced %d contours:\n", len(res.Contours))
	for i, c := range res.Contours {
		sa := converter.SignedArea(c.Points)
		bb := converter.ContourBoundingBox(c.Points)
		fmt.Printf("Contour %d: %d pts, area=%.2f, CW=%v, bb=[%.2f, %.2f] to [%.2f, %.2f]\n",
			i, len(c.Points), sa, sa < 0, bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
	}
}
