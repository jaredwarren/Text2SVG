package converter

import (
	"fmt"
	"strings"
)

// GenerateSVG produces a clean, CAD-optimized vector SVG string with physical dimensions,
// evenodd fill rules for inner counters, optional boundary enclosures, and optional reference construction frames.
func GenerateSVG(contours []Contour, segments []PathSegment, bounds BoundingBox, units Units, constructionBox bool, boundaryContours []Contour) string {
	width := bounds.Width()
	height := bounds.Height()

	if width <= 0 {
		width = 10.0
	}
	if height <= 0 {
		height = 10.0
	}

	unitSuffix := "mm"
	if units == UnitsInches {
		unitSuffix = "in"
	} else if units == UnitsPixels {
		unitSuffix = "px"
	}

	// We flip Y so that Cartesian (Y up) becomes SVG (Y down):
	// svgX = x - bounds.MinX
	// svgY = bounds.MaxY - y
	toSVG := func(p Point) (float64, float64) {
		return p.X - bounds.MinX, bounds.MaxY - p.Y
	}

	var pathD strings.Builder

	if len(segments) > 0 {
		// Use exact Bézier segments
		for _, seg := range segments {
			switch seg.Type {
			case SegmentMoveTo:
				x, y := toSVG(seg.Args[0])
				pathD.WriteString(fmt.Sprintf("M %.3f,%.3f ", x, y))
			case SegmentLineTo:
				x, y := toSVG(seg.Args[0])
				pathD.WriteString(fmt.Sprintf("L %.3f,%.3f ", x, y))
			case SegmentQuadTo:
				cx, cy := toSVG(seg.Args[0])
				x, y := toSVG(seg.Args[1])
				pathD.WriteString(fmt.Sprintf("Q %.3f,%.3f %.3f,%.3f ", cx, cy, x, y))
			case SegmentCubeTo:
				c1x, c1y := toSVG(seg.Args[0])
				c2x, c2y := toSVG(seg.Args[1])
				x, y := toSVG(seg.Args[2])
				pathD.WriteString(fmt.Sprintf("C %.3f,%.3f %.3f,%.3f %.3f,%.3f ", c1x, c1y, c2x, c2y, x, y))
			case SegmentClose:
				pathD.WriteString("Z ")
			}
		}
	} else {
		// Fallback to sampled contours
		for _, c := range contours {
			if len(c.Points) == 0 {
				continue
			}
			x0, y0 := toSVG(c.Points[0])
			pathD.WriteString(fmt.Sprintf("M %.3f,%.3f ", x0, y0))
			for i := 1; i < len(c.Points); i++ {
				x, y := toSVG(c.Points[i])
				pathD.WriteString(fmt.Sprintf("L %.3f,%.3f ", x, y))
			}
			if c.Closed {
				pathD.WriteString("Z ")
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.3f %.3f" width="%.3f%s" height="%.3f%s" style="overflow: visible;">`+"\n",
		width, height, width, unitSuffix, height, unitSuffix))
	sb.WriteString("  <defs>\n")
	sb.WriteString("    <style>\n")
	sb.WriteString("      .cut-path { fill: none; stroke: #e02424; stroke-width: 0.35; stroke-linecap: round; stroke-linejoin: round; fill-rule: evenodd; }\n")
	sb.WriteString("      .fill-path { fill: #f3f4f6; stroke: none; fill-rule: evenodd; }\n")
	sb.WriteString("      .boundary-path { fill: none; stroke: #38bdf8; stroke-width: 0.5; stroke-linecap: round; stroke-linejoin: round; fill-rule: evenodd; }\n")
	sb.WriteString("      .construction-frame { fill: none; stroke: #6b7280; stroke-width: 0.25; stroke-dasharray: 2,2; opacity: 0.75; }\n")
	sb.WriteString("    </style>\n")
	sb.WriteString("  </defs>\n")
	if len(boundaryContours) > 0 {
		var bD strings.Builder
		for _, bc := range boundaryContours {
			if len(bc.Points) < 3 {
				continue
			}
			x0, y0 := toSVG(bc.Points[0])
			bD.WriteString(fmt.Sprintf("M %.3f,%.3f ", x0, y0))
			for i := 1; i < len(bc.Points); i++ {
				x, y := toSVG(bc.Points[i])
				bD.WriteString(fmt.Sprintf("L %.3f,%.3f ", x, y))
			}
			bD.WriteString("Z ")
		}
		sb.WriteString(fmt.Sprintf(`  <path class="boundary-path" fill-rule="evenodd" d="%s" />`+"\n", strings.TrimSpace(bD.String())))
	}
	sb.WriteString(fmt.Sprintf(`  <path class="cut-path" fill-rule="evenodd" d="%s" />`+"\n", strings.TrimSpace(pathD.String())))
	if constructionBox {
		sb.WriteString(fmt.Sprintf(`  <rect class="construction-frame" x="0.0" y="0.0" width="%.3f" height="%.3f" />`+"\n", width, height))
	}
	sb.WriteString("</svg>")

	return sb.String()
}
