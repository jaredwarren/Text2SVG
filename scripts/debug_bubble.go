//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
	"github.com/lestrrat-go/polyclip"
	"github.com/lestrrat-go/polyclip/geom"
)

func main() {
	probeAcademy()
	unused()
}

func probeAcademy() {
	data, _ := os.ReadFile("/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf")
	lf, _ := converter.ParseFont(data, "a")
	res, _ := lf.LayoutText(converter.TextParams{Text: "6", Size: 30, Units: converter.UnitsMM})
	cls := converter.ClassifyAndOrientContours(res.Contours, 1e-4)
	var ring geom.Polygon
	for _, p := range cls[0].Points {
		ring = append(ring, geom.Point{X: p.X, Y: p.Y})
	}
	if ring.SignedArea() < 0 {
		ring.Reverse()
	}
	for _, d := range []float64{0.5, 1, 2, 4} {
		out, err := polyclip.Offset(geom.MultiPolygon{{Outer: ring}}, d, polyclip.OffsetOptions{Join: polyclip.JoinRound, ArcTol: d * 0.01, MiterLimit: 3})
		fmt.Printf("direct d=%.1f err=%v pieces=%d", d, err, len(out))
		for i, ex := range out {
			fmt.Printf(" [%d area=%.1f pts=%d holes=%d]", i, ex.Outer.SignedArea(), len(ex.Outer), len(ex.Holes))
		}
		fmt.Println()
		if s, err := polyclip.Simplify(out); err == nil {
			fmt.Printf("  simplify pieces=%d", len(s))
			for i, ex := range s {
				fmt.Printf(" [%d area=%.1f pts=%d]", i, ex.Outer.SignedArea(), len(ex.Outer))
			}
			fmt.Println()
		}
	}
}

func unused() {
	cases := []struct {
		font, text, name string
		d                float64
		join             converter.CornerJoin
	}{
		{"/System/Library/Fonts/Supplemental/Arial.ttf", "LASER", "arial_laser05", 0.5, converter.JoinRound},
		{"/System/Library/Fonts/Supplemental/Academy Engraved LET Fonts.ttf", "6", "acad6_d05", 0.5, converter.JoinRound},
		{"/System/Library/Fonts/Supplemental/Times New Roman.ttf", "6", "times6_d05", 0.5, converter.JoinRound},
		{"/System/Library/Fonts/Supplemental/Arial.ttf", "6", "arial6_d8", 8, converter.JoinRound},
		{"/System/Library/Fonts/Supplemental/Brush Script.ttf", "6", "brush6_d2", 2, converter.JoinRound},
		{"/System/Library/Fonts/Supplemental/SnellRoundhand.ttf", "Ray", "snell_d3", 3, converter.JoinRound},
	}
	for _, c := range cases {
		data, err := os.ReadFile(c.font)
		if err != nil {
			fmt.Println("skip", c.name, err)
			continue
		}
		lf, err := converter.ParseFont(data, c.font)
		if err != nil {
			fmt.Println("parse", c.name, err)
			continue
		}
		res, err := lf.LayoutText(converter.TextParams{
			Text:              c.text,
			Size:              30,
			Units:             converter.UnitsMM,
			BoundaryMode:      converter.BoundaryConformal,
			BoundaryOffset:    c.d,
			BoundaryFillHoles: true,
			CornerJoin:        c.join,
			Datum:             converter.DatumCenter,
		})
		if err != nil {
			fmt.Println("layout", c.name, err)
			continue
		}
		cls := converter.ClassifyAndOrientContours(res.Contours, 1e-4)
		fmt.Printf("%s contours=%d boundary=%d classified=%d\n", c.name, len(res.Contours), len(res.BoundaryContours), len(cls))
		for i, cc := range cls {
			minE := 1e9
			n := len(cc.Points)
			for k := 0; k < n; k++ {
				a, b := cc.Points[k], cc.Points[(k+1)%n]
				if e := math.Hypot(a.X-b.X, a.Y-b.Y); e < minE {
					minE = e
				}
			}
			fmt.Printf("  src%d role=%s depth=%d area=%.1f pts=%d minEdge=%.4f\n", i, cc.Role, cc.Depth, cc.Area, len(cc.Points), minE)
		}
		for i, b := range res.BoundaryContours {
			fmt.Printf("  b%d pts=%d area=%.1f\n", i, len(b.Points), converter.SignedArea(b.Points))
		}
		writeSVG("/tmp/"+c.name+".svg", res)
	}
}

func writeSVG(path string, res *converter.RenderResult) {
	var b []byte
	b = append(b, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="900" height="700">`)...)
	b = append(b, []byte(fmt.Sprintf(`<g transform="translate(450,350) scale(6,-6)">`))...)
	for _, c := range res.Contours {
		b = append(b, []byte(fmt.Sprintf(`<path d="%s" fill="none" stroke="#c0392b" stroke-width="0.15"/>`, dPath(c.Points)))...)
	}
	for _, c := range res.BoundaryContours {
		b = append(b, []byte(fmt.Sprintf(`<path d="%s" fill="none" stroke="#2980b9" stroke-width="0.2"/>`, dPath(c.Points)))...)
	}
	b = append(b, []byte(`</g></svg>`)...)
	_ = os.WriteFile(path, b, 0644)
}

func dPath(pts []converter.Point) string {
	s := ""
	for i, p := range pts {
		if i == 0 {
			s += fmt.Sprintf("M %.3f %.3f ", p.X, p.Y)
		} else {
			s += fmt.Sprintf("L %.3f %.3f ", p.X, p.Y)
		}
	}
	return s + "Z"
}
