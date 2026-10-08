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

	c0 := res.Contours[0].Points
	fmt.Printf("c0 has %d points\n", len(c0))
	for i := 0; i < 20; i++ {
		fmt.Printf("pt %d: (%.3f, %.3f)\n", i, c0[i].X, c0[i].Y)
	}
	for i := len(c0) - 20; i < len(c0); i++ {
		fmt.Printf("pt %d: (%.3f, %.3f)\n", i, c0[i].X, c0[i].Y)
	}
}
