package converter

import (
	"fmt"
	"strings"
)

// GenerateDXF produces a fully compliant, CAM-compatible DXF file (AutoCAD R12/2000).
// Each closed contour is emitted as a closed POLYLINE entity with 2D vertices.
func GenerateDXF(contours []Contour, layerName string, units Units) string {
	if layerName == "" {
		layerName = "CUT"
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

	var sb strings.Builder

	// Header section
	sb.WriteString("  0\nSECTION\n  2\nHEADER\n")
	sb.WriteString("  9\n$ACADVER\n  1\nAC1009\n") // AutoCAD Release 12 - universally compatible
	sb.WriteString(fmt.Sprintf("  9\n$INSUNITS\n 70\n%6d\n", insUnits))
	sb.WriteString("  0\nENDSEC\n")

	// Tables section
	sb.WriteString("  0\nSECTION\n  2\nTABLES\n")

	// Linetype table
	sb.WriteString("  0\nTABLE\n  2\nLTYPE\n 70\n     1\n")
	sb.WriteString("  0\nLTYPE\n  2\nCONTINUOUS\n 70\n     0\n  3\nSolid line\n 72\n    65\n 73\n     0\n 40\n0.0\n")
	sb.WriteString("  0\nENDTAB\n")

	// Layer table
	sb.WriteString("  0\nTABLE\n  2\nLAYER\n 70\n     2\n")
	sb.WriteString("  0\nLAYER\n  2\n0\n 70\n     0\n 62\n     7\n  6\nCONTINUOUS\n")
	sb.WriteString(fmt.Sprintf("  0\nLAYER\n  2\n%s\n 70\n     0\n 62\n%6d\n  6\nCONTINUOUS\n", layerName, color))
	sb.WriteString("  0\nENDTAB\n")

	sb.WriteString("  0\nENDSEC\n")

	// Entities section
	sb.WriteString("  0\nSECTION\n  2\nENTITIES\n")

	for _, contour := range contours {
		if len(contour.Points) < 2 {
			continue
		}

		closedFlag := 0
		if contour.Closed {
			closedFlag = 1
		}

		// POLYLINE header
		sb.WriteString("  0\nPOLYLINE\n")
		sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
		sb.WriteString(" 66\n     1\n") // Vertices follow flag
		sb.WriteString(fmt.Sprintf(" 70\n%6d\n", closedFlag))
		sb.WriteString(" 10\n0.0\n 20\n0.0\n 30\n0.0\n")

		// VERTEX entities
		for _, pt := range contour.Points {
			sb.WriteString("  0\nVERTEX\n")
			sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
			sb.WriteString(fmt.Sprintf(" 10\n%.4f\n", pt.X))
			sb.WriteString(fmt.Sprintf(" 20\n%.4f\n", pt.Y))
			sb.WriteString(" 30\n0.0\n")
		}

		// SEQEND
		sb.WriteString("  0\nSEQEND\n")
		sb.WriteString(fmt.Sprintf("  8\n%s\n", layerName))
	}

	sb.WriteString("  0\nENDSEC\n")
	sb.WriteString("  0\nEOF\n")

	return sb.String()
}
