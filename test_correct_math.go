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
	var tangents []converter.Point
	for i := 0; i < n; i++ {
		p0 := c0[i]
		p1 := c0[(i+1)%n]
		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		lenSeg := math.Hypot(dx, dy)
		if lenSeg < 1e-9 {
			lenSeg = 1e-9
		}
		tangents = append(tangents, converter.Point{X: dx / lenSeg, Y: dy / lenSeg})
		normals = append(normals, converter.Point{X: -dy / lenSeg, Y: dx / lenSeg})
	}

	i := 0
	p := c0[i]
	nPrev := normals[(i-1+n)%n]
	nCurr := normals[i]
	vPrev := tangents[(i-1+n)%n]
	vCurr := tangents[i]

	pPrevOffset := converter.Point{X: p.X + d*nPrev.X, Y: p.Y + d*nPrev.Y}
	pCurrOffset := converter.Point{X: p.X + d*nCurr.X, Y: p.Y + d*nCurr.Y}

	denom := vPrev.X*vCurr.Y - vPrev.Y*vCurr.X
	u := ((pCurrOffset.X-pPrevOffset.X)*vCurr.Y - (pCurrOffset.Y-pPrevOffset.Y)*vCurr.X) / denom
	w := ((pCurrOffset.X-pPrevOffset.X)*vPrev.Y - (pCurrOffset.Y-pPrevOffset.Y)*vPrev.X) / denom
	inter1 := converter.Point{X: pPrevOffset.X + u*vPrev.X, Y: pPrevOffset.Y + u*vPrev.Y}
	inter2 := converter.Point{X: pCurrOffset.X + w*vCurr.X, Y: pCurrOffset.Y + w*vCurr.Y}

	fmt.Printf("p: (%.3f, %.3f)\n", p.X, p.Y)
	fmt.Printf("pPrevOffset: (%.3f, %.3f), pCurrOffset: (%.3f, %.3f)\n", pPrevOffset.X, pPrevOffset.Y, pCurrOffset.X, pCurrOffset.Y)
	fmt.Printf("denom: %.5f, u: %.3f, w: %.3f\n", denom, u, w)
	fmt.Printf("inter1: (%.3f, %.3f)\n", inter1.X, inter1.Y)
	fmt.Printf("inter2: (%.3f, %.3f)\n", inter2.X, inter2.Y)
	fmt.Printf("dist from p: %.3f\n", math.Hypot(inter1.X-p.X, inter1.Y-p.Y))
}
