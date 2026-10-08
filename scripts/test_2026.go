//go:build ignore

package main

import (
	"os"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
)

func main() {
	data, _ := os.ReadFile("/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf")
	lf, _ := converter.ParseFont(data, "Academy")

	res, _ := lf.LayoutText(converter.TextParams{
		Text:              "2026",
		Size:              30.0,
		Units:             converter.UnitsMM,
		Weld:              true,
		BoundaryMode:      converter.BoundaryConformal,
		BoundaryOffset:    8.0,
		BoundaryFillHoles: true,
		Datum:             converter.DatumCenter,
	})

	_ = os.WriteFile("test_2026.svg", []byte(res.SVG), 0644)
}
