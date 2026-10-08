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

	res, _ := lf.LayoutText(converter.TextParams{
		Text:              "6",
		Size:              30.0,
		Units:             converter.UnitsMM,
		Weld:              true,
		BoundaryMode:      converter.BoundaryNone,
		Datum:             converter.DatumBottomLeft,
	})

	c0 := res.Contours[0].Points
	d := 8.0
	n := len(c0)

	var normals []converter.Point
	for i := 0; i < n; i++ {
		p0 := c0[i]
		p1 := c0[(i+1)%n]
		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		lenSeg := math.Hypot(dx, dy)
		if lenSeg < 1e-9 {
			lenSeg = 1e-9
		}
		normals = append(normals, converter.Point{X: -dy / lenSeg, Y: dx / lenSeg})
	}

	i := 0
	p := c0[i]
	nPrev := normals[(i-1+n)%n]
	nCurr := normals[i]

	pPrevOffset := converter.Point{X: p.X + d*nPrev.X, Y: p.Y + d*nPrev.Y}
	pCurrOffset := converter.Point{X: p.X + d*nCurr.X, Y: p.Y + d*nCurr.Y}

	// vPrev is direction of prev segment: (p.X - c0[n-1].X, p.Y - c0[n-1].Y)
	vPrev := converter.Point{X: p.X - c0[n-1].X, Y: p.Y - c0[n-1].Y}
	lenPrev := math.Hypot(vPrev.X, vPrev.Y)
	vPrev.X /= lenPrev
	vPrev.Y /= lenPrev

	// vCurr is direction of curr segment: (c0[1].X - p.X, c0[1].Y - p.Y)
	vCurr := converter.Point{X: c0[1].X - p.X, Y: c0[1].Y - p.Y}
	lenCurr := math.Hypot(vCurr.X, vCurr.Y)
	vCurr.X /= lenCurr
	vCurr.Y /= lenCurr

	// In offset.go:
	d1x, d1y := -nPrev.Y, nPrev.X
	d2x, d2y := -nCurr.Y, nCurr.X
	fmt.Printf("pPrevOffset: (%.3f, %.3f), pCurrOffset: (%.3f, %.3f)\n", pPrevOffset.X, pPrevOffset.Y, pCurrOffset.X, pCurrOffset.Y)
	fmt.Printf("vPrev: (%.3f, %.3f), d1: (%.3f, %.3f)\n", vPrev.X, vPrev.Y, d1x, d1y)
	fmt.Printf("vCurr: (%.3f, %.3f), d2: (%.3f, %.3f)\n", vCurr.X, vCurr.Y, d2x, d2y)
}
