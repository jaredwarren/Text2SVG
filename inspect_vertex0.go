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
	cross := nPrev.X*nCurr.Y - nPrev.Y*nCurr.X
	dot := nPrev.X*nCurr.X + nPrev.Y*nCurr.Y

	fmt.Printf("p: (%.3f, %.3f)\n", p.X, p.Y)
	fmt.Printf("pPrev: (%.3f, %.3f)\n", c0[n-1].X, c0[n-1].Y)
	fmt.Printf("pNext: (%.3f, %.3f)\n", c0[1].X, c0[1].Y)
	fmt.Printf("nPrev: (%.3f, %.3f)\n", nPrev.X, nPrev.Y)
	fmt.Printf("nCurr: (%.3f, %.3f)\n", nCurr.X, nCurr.Y)
	fmt.Printf("cross: %.5f, dot: %.5f\n", cross, dot)
	fmt.Printf("isOpeningCorner: %v\n", (cross*d) <= 0)

	pPrevOffset := converter.Point{X: p.X + d*nPrev.X, Y: p.Y + d*nPrev.Y}
	pCurrOffset := converter.Point{X: p.X + d*nCurr.X, Y: p.Y + d*nCurr.Y}
	fmt.Printf("pPrevOffset: (%.3f, %.3f)\n", pPrevOffset.X, pPrevOffset.Y)
	fmt.Printf("pCurrOffset: (%.3f, %.3f)\n", pCurrOffset.X, pCurrOffset.Y)

	d1x, d1y := -nPrev.Y, nPrev.X
	d2x, d2y := -nCurr.Y, nCurr.X
	denom := d1x*d2y - d1y*d2x
	fmt.Printf("denom: %.5f\n", denom)
	u := ((pCurrOffset.X-pPrevOffset.X)*d2y - (pCurrOffset.Y-pPrevOffset.Y)*d2x) / denom
	inter := converter.Point{X: pPrevOffset.X + u*d1x, Y: pPrevOffset.Y + u*d1y}
	dist := math.Hypot(inter.X-p.X, inter.Y-p.Y)
	fmt.Printf("inter: (%.3f, %.3f), dist: %.3f (limit: %.3f)\n", inter.X, inter.Y, dist, 3.0*d)
}
