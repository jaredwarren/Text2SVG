package converter

import (
	"fmt"
	"strings"
)

// GenerateDXF produces a fully compliant, CAM/CAD-compatible DXF file.
// Supports:
//   - DXFFormatSpline (AutoCAD 2000 AC1015): Native cubic B-spline (SPLINE) and LINE entities
//     for true mathematically smooth surfaces in Fusion 360, SolidWorks, FreeCAD, and AutoCAD.
//   - DXFFormatPolyline: Lightweight (LWPOLYLINE) or standard closed polylines.
func GenerateDXF(contours []Contour, segments []PathSegment, format DXFFormat, layerName string, units Units, constructionBox bool, bounds BoundingBox, boundaryContours []Contour) string {
	if layerName == "" {
		layerName = "CUT"
	}
	if format == "" {
		format = DXFFormatSpline
	}

	// Insunits code: 4 = Millimeters, 1 = Inches, 0 = Unitless
	insUnits := 4
	if units == UnitsInches {
		insUnits = 1
	} else if units == UnitsPixels {
		insUnits = 0
	}

	// Layer color (AutoCAD Color Index):
	// CUT -> 1 (Red), SCORE -> 2 (Yellow), ENGRAVE -> 5 (Blue)
	color := 1
	upperLayer := strings.ToUpper(layerName)
	if strings.Contains(upperLayer, "SCORE") {
		color = 2
	} else if strings.Contains(upperLayer, "ENGRAVE") {
		color = 5
	}

	layerCount := 2
	if constructionBox {
		layerCount++
	}
	if len(boundaryContours) > 0 {
		layerCount++
	}

	var sb strings.Builder

	// Header section
	sb.WriteString("  0\nSECTION\n  2\nHEADER\n")
	if format == DXFFormatSpline {
		sb.WriteString("  9\n$ACADVER\n  1\nAC1015\n") // AutoCAD 2000
		sb.WriteString("  9\n$HANDSEED\n  5\nFFFF\n")
	} else {
		sb.WriteString("  9\n$ACADVER\n  1\nAC1009\n") // AutoCAD Release 12
	}
	sb.WriteString(fmt.Sprintf("  9\n$INSUNITS\n 70\n%6d\n", insUnits))
	sb.WriteString("  0\nENDSEC\n")

	// Classes section (required in AC1015+)
	if format == DXFFormatSpline {
		sb.WriteString("  0\nSECTION\n  2\nCLASSES\n  0\nENDSEC\n")
	}

	// Tables section
	sb.WriteString("  0\nSECTION\n  2\nTABLES\n")

	// Linetype table
	sb.WriteString("  0\nTABLE\n  2\nLTYPE\n 70\n     1\n")
	sb.WriteString("  0\nLTYPE\n  2\nCONTINUOUS\n 70\n     0\n  3\nSolid line\n 72\n    65\n 73\n     0\n 40\n0.0\n")
	sb.WriteString("  0\nENDTAB\n")

	// Layer table
	sb.WriteString("  0\nTABLE\n  2\nLAYER\n")
	sb.WriteString(fmt.Sprintf(" 70\n%6d\n", layerCount))
	sb.WriteString("  0\nLAYER\n  2\n0\n 70\n     0\n 62\n     7\n  6\nCONTINUOUS\n")
	sb.WriteString(fmt.Sprintf("  0\nLAYER\n  2\n%s\n 70\n     0\n 62\n%6d\n  6\nCONTINUOUS\n", layerName, color))
	if constructionBox {
		// Gray color 8 for construction references
		sb.WriteString("  0\nLAYER\n  2\nCONSTRUCTION\n 70\n     0\n 62\n     8\n  6\nCONTINUOUS\n")
	}
	if len(boundaryContours) > 0 {
		// Cyan color 4 for perimeter BORDER
		sb.WriteString("  0\nLAYER\n  2\nBORDER\n 70\n     0\n 62\n     4\n  6\nCONTINUOUS\n")
	}
	sb.WriteString("  0\nENDTAB\n")

	// Block Record table (required in AC1015)
	if format == DXFFormatSpline {
		sb.WriteString("  0\nTABLE\n  2\nBLOCK_RECORD\n 70\n     1\n")
		sb.WriteString("  0\nBLOCK_RECORD\n  2\n*MODEL_SPACE\n")
		sb.WriteString("  0\nENDTAB\n")
	}

	sb.WriteString("  0\nENDSEC\n")

	// Blocks section (required in AC1015)
	if format == DXFFormatSpline {
		sb.WriteString("  0\nSECTION\n  2\nBLOCKS\n  0\nENDSEC\n")
	}

	// Entities section
	sb.WriteString("  0\nSECTION\n  2\nENTITIES\n")

	// Emit construction bounding box if requested
	if constructionBox {
		rectPts := []Point{
			{X: bounds.MinX, Y: bounds.MinY},
			{X: bounds.MaxX, Y: bounds.MinY},
			{X: bounds.MaxX, Y: bounds.MaxY},
			{X: bounds.MinX, Y: bounds.MaxY},
			{X: bounds.MinX, Y: bounds.MinY},
		}
		if format == DXFFormatSpline {
			emitLWPolyline(&sb, rectPts, "CONSTRUCTION", true)
		} else {
			emitR12Polyline(&sb, rectPts, "CONSTRUCTION", true)
		}
	}

	// Emit perimeter boundary contours on layer BORDER
	if len(boundaryContours) > 0 {
		for _, bc := range boundaryContours {
			if len(bc.Points) < 3 {
				continue
			}
			if format == DXFFormatSpline {
				emitLWPolyline(&sb, bc.Points, "BORDER", true)
			} else {
				emitR12Polyline(&sb, bc.Points, "BORDER", true)
			}
		}
	}

	if format == DXFFormatSpline && len(segments) > 0 {
		// Output True DXF Splines and Lines from exact Bézier segments
		emitSplineEntities(&sb, segments, layerName, color)
	} else if format == DXFFormatSpline && len(contours) > 0 {
		// AutoCAD 2000 Lightweight Polylines (LWPOLYLINE) for closed loops
		for _, c := range contours {
			emitLWPolyline(&sb, c.Points, layerName, c.Closed)
		}
	} else {
		// Standard R12 POLYLINE entities
		for _, contour := range contours {
			emitR12Polyline(&sb, contour.Points, layerName, contour.Closed)
		}
	}

	sb.WriteString("  0\nENDSEC\n")

	// Objects section (required in AC1015+)
	if format == DXFFormatSpline {
		sb.WriteString("  0\nSECTION\n  2\nOBJECTS\n  0\nENDSEC\n")
	}

	sb.WriteString("  0\nEOF\n")

	return sb.String()
}

// emitSplineEntities writes LINE and cubic SPLINE entities corresponding to Bézier path segments.
func emitSplineEntities(sb *strings.Builder, segments []PathSegment, layerName string, color int) {
	var currentPos Point
	var startPos Point

	for _, seg := range segments {
		switch seg.Type {
		case SegmentMoveTo:
			currentPos = seg.Args[0]
			startPos = currentPos

		case SegmentLineTo:
			dest := seg.Args[0]
			emitDXFLine(sb, layerName, currentPos, dest)
			currentPos = dest

		case SegmentQuadTo:
			// Elevate quadratic to cubic Bézier
			ctrl := seg.Args[0]
			dest := seg.Args[1]
			c0, c1, c2, c3 := QuadToCubic(currentPos, ctrl, dest)
			emitDXFSpline(sb, layerName, c0, c1, c2, c3)
			currentPos = dest

		case SegmentCubeTo:
			ctrl1 := seg.Args[0]
			ctrl2 := seg.Args[1]
			dest := seg.Args[2]
			emitDXFSpline(sb, layerName, currentPos, ctrl1, ctrl2, dest)
			currentPos = dest

		case SegmentClose:
			if (currentPos.X != startPos.X) || (currentPos.Y != startPos.Y) {
				emitDXFLine(sb, layerName, currentPos, startPos)
				currentPos = startPos
			}
		}
	}
}

// emitDXFLine outputs an AutoCAD LINE entity.
func emitDXFLine(sb *strings.Builder, layerName string, p0, p1 Point) {
	sb.WriteString("  0\nLINE\n")
	sb.WriteString("100\nAcDbEntity\n")
	sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
	sb.WriteString("100\nAcDbLine\n")
	sb.WriteString(fmt.Sprintf(" 10\n%.4f\n 20\n%.4f\n 30\n0.0\n", p0.X, p0.Y))
	sb.WriteString(fmt.Sprintf(" 11\n%.4f\n 21\n%.4f\n 31\n0.0\n", p1.X, p1.Y))
}

// emitDXFSpline outputs an AutoCAD 2000 native cubic B-spline (SPLINE entity).
// Degree 3, 4 control points, knot vector [0, 0, 0, 0, 1, 1, 1, 1].
func emitDXFSpline(sb *strings.Builder, layerName string, p0, p1, p2, p3 Point) {
	sb.WriteString("  0\nSPLINE\n")
	sb.WriteString("100\nAcDbEntity\n")
	sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
	sb.WriteString("100\nAcDbSpline\n")
	sb.WriteString("210\n0.0\n220\n0.0\n230\n1.0\n") // Normal vector Z = 1
	sb.WriteString(" 70\n     8\n")                  // Planar = 8
	sb.WriteString(" 71\n     3\n")                  // Degree = 3 (cubic)
	sb.WriteString(" 72\n     8\n")                  // 8 knots
	sb.WriteString(" 73\n     4\n")                  // 4 control points
	sb.WriteString(" 74\n     0\n")                  // 0 fit points
	sb.WriteString(" 42\n1.000000e-07\n")            // Knot tolerance
	sb.WriteString(" 43\n1.000000e-07\n")            // Control-point tolerance
	sb.WriteString(" 44\n1.000000e-07\n")            // Fit tolerance

	// Knot vector for clamped cubic Bézier: [0, 0, 0, 0, 1, 1, 1, 1]
	for i := 0; i < 4; i++ {
		sb.WriteString(" 40\n0.0\n")
	}
	for i := 0; i < 4; i++ {
		sb.WriteString(" 40\n1.0\n")
	}

	// 4 Control points
	ctrls := []Point{p0, p1, p2, p3}
	for _, cp := range ctrls {
		sb.WriteString(fmt.Sprintf(" 10\n%.4f\n 20\n%.4f\n 30\n0.0\n", cp.X, cp.Y))
	}
}

// emitLWPolyline writes an AutoCAD 2000 lightweight 2D polyline.
func emitLWPolyline(sb *strings.Builder, pts []Point, layerName string, closed bool) {
	if len(pts) < 2 {
		return
	}
	closedFlag := 0
	if closed {
		closedFlag = 1
	}

	sb.WriteString("  0\nLWPOLYLINE\n")
	sb.WriteString("100\nAcDbEntity\n")
	sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
	sb.WriteString("100\nAcDbPolyline\n")
	sb.WriteString(fmt.Sprintf(" 90\n%6d\n", len(pts)))
	sb.WriteString(fmt.Sprintf(" 70\n%6d\n", closedFlag))
	sb.WriteString(" 43\n0.0\n")

	for _, pt := range pts {
		sb.WriteString(fmt.Sprintf(" 10\n%.4f\n 20\n%.4f\n", pt.X, pt.Y))
	}
}

// emitR12Polyline writes an AutoCAD Release 12 closed 2D polyline with vertex sequence.
func emitR12Polyline(sb *strings.Builder, pts []Point, layerName string, closed bool) {
	if len(pts) < 2 {
		return
	}
	closedFlag := 0
	if closed {
		closedFlag = 1
	}

	sb.WriteString("  0\nPOLYLINE\n")
	sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
	sb.WriteString(" 66\n     1\n") // Vertices follow
	sb.WriteString(fmt.Sprintf(" 70\n%6d\n", closedFlag))
	sb.WriteString(" 10\n0.0\n 20\n0.0\n 30\n0.0\n")

	for _, pt := range pts {
		sb.WriteString("  0\nVERTEX\n")
		sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
		sb.WriteString(fmt.Sprintf(" 10\n%.4f\n 20\n%.4f\n 30\n0.0\n", pt.X, pt.Y))
	}

	sb.WriteString("  0\nSEQEND\n")
	sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
}
