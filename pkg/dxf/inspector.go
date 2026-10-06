package dxf

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// Tag represents a DXF group code and its associated value.
type Tag struct {
	Code  int
	Value string
}

// Diagnostic represents an issue, warning, or recommendation.
type Diagnostic struct {
	Severity string // "CRITICAL", "WARNING", "INFO", "SUCCESS"
	Message  string
}

// Report contains full health, geometry, and compatibility analysis of a DXF file.
type Report struct {
	Filename        string            `json:"filename"`
	Version         string            `json:"version"`          // e.g. "AC1009", "AC1015"
	VersionName     string            `json:"version_name"`     // e.g. "AutoCAD 2000 (AC1015)"
	Units           string            `json:"units"`            // e.g. "Millimeters"
	InsUnitsCode    int               `json:"insunits_code"`
	HandSeed        string            `json:"handseed,omitempty"`
	Sections        []string          `json:"sections"`
	MissingSections []string          `json:"missing_sections,omitempty"`
	Tables          []string          `json:"tables"`
	TotalEntities   int               `json:"total_entities"`
	EntityCounts    map[string]int    `json:"entity_counts"`
	ClosedLoops     int               `json:"closed_loops"`
	OpenCurves      int               `json:"open_curves"`
	BoundsMinX      float64           `json:"bounds_min_x"`
	BoundsMinY      float64           `json:"bounds_min_y"`
	BoundsMaxX      float64           `json:"bounds_max_x"`
	BoundsMaxY      float64           `json:"bounds_max_y"`
	Width           float64           `json:"width"`
	Height          float64           `json:"height"`
	Diagnostics     []Diagnostic      `json:"diagnostics"`
	CADCompatible   bool              `json:"cad_compatible"`
}

// InspectFile parses and analyzes a DXF file from disk.
func InspectFile(path string) (*Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DXF file: %w", err)
	}
	defer f.Close()

	rep, err := InspectReader(f)
	if err != nil {
		return nil, err
	}
	rep.Filename = path
	return rep, nil
}

// InspectReader parses and analyzes DXF stream from an io.Reader.
func InspectReader(r io.Reader) (*Report, error) {
	scanner := bufio.NewScanner(r)
	// Allow large tokens if needed
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var tags []Tag
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		code, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		if !scanner.Scan() {
			break
		}
		val := strings.TrimSpace(scanner.Text())
		tags = append(tags, Tag{Code: code, Value: val})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return AnalyzeTags(tags)
}

// AnalyzeTags analyzes parsed DXF tags and builds a comprehensive report.
func AnalyzeTags(tags []Tag) (*Report, error) {
	rep := &Report{
		EntityCounts: make(map[string]int),
		BoundsMinX:   math.MaxFloat64,
		BoundsMinY:   math.MaxFloat64,
		BoundsMaxX:   -math.MaxFloat64,
		BoundsMaxY:   -math.MaxFloat64,
	}

	var currentSection string
	var inEntities bool
	var inHeader bool
	var inTables bool
	var currentTable string
	var currentEntity string
	var currentEntityClosed bool

	var splineCount int
	var lineCount int
	var polylineCount int
	var hasHandles bool

	updateBounds := func(x, y float64) {
		if x < rep.BoundsMinX {
			rep.BoundsMinX = x
		}
		if x > rep.BoundsMaxX {
			rep.BoundsMaxX = x
		}
		if y < rep.BoundsMinY {
			rep.BoundsMinY = y
		}
		if y > rep.BoundsMaxY {
			rep.BoundsMaxY = y
		}
	}

	var currentX float64
	hasPendingX := false

	for i := 0; i < len(tags); i++ {
		tag := tags[i]

		// Check for handles
		if tag.Code == 5 {
			hasHandles = true
		}

		// Section boundaries
		if tag.Code == 0 && tag.Value == "SECTION" {
			if i+1 < len(tags) && tags[i+1].Code == 2 {
				currentSection = tags[i+1].Value
				rep.Sections = append(rep.Sections, currentSection)
				inEntities = (currentSection == "ENTITIES")
				inHeader = (currentSection == "HEADER")
				inTables = (currentSection == "TABLES")
			}
			continue
		}
		if tag.Code == 0 && tag.Value == "ENDSEC" {
			currentSection = ""
			inEntities = false
			inHeader = false
			inTables = false
			continue
		}

		// Header variables
		if inHeader {
			if tag.Code == 9 && tag.Value == "$ACADVER" {
				if i+1 < len(tags) && tags[i+1].Code == 1 {
					rep.Version = tags[i+1].Value
				}
			}
			if tag.Code == 9 && tag.Value == "$INSUNITS" {
				if i+1 < len(tags) && tags[i+1].Code == 70 {
					code, _ := strconv.Atoi(tags[i+1].Value)
					rep.InsUnitsCode = code
				}
			}
			if tag.Code == 9 && tag.Value == "$HANDSEED" {
				if i+1 < len(tags) && tags[i+1].Code == 5 {
					rep.HandSeed = tags[i+1].Value
				}
			}
		}

		// Tables
		if inTables {
			if tag.Code == 0 && tag.Value == "TABLE" {
				if i+1 < len(tags) && tags[i+1].Code == 2 {
					currentTable = tags[i+1].Value
					rep.Tables = append(rep.Tables, currentTable)
				}
			}
		}

		// Coordinates tracking
		if inEntities {
			if tag.Code == 10 {
				if v, err := strconv.ParseFloat(tag.Value, 64); err == nil {
					currentX = v
					hasPendingX = true
				}
			} else if tag.Code == 20 && hasPendingX {
				if v, err := strconv.ParseFloat(tag.Value, 64); err == nil {
					updateBounds(currentX, v)
					hasPendingX = false
				}
			}
		}

		// Entities parsing
		if inEntities && tag.Code == 0 && tag.Value != "ENDSEC" {
			// Finish previous entity
			if currentEntity != "" {
				rep.EntityCounts[currentEntity]++
				rep.TotalEntities++
				if currentEntity == "POLYLINE" || currentEntity == "LWPOLYLINE" {
					if currentEntityClosed {
						rep.ClosedLoops++
					} else {
						rep.OpenCurves++
					}
				}
			}

			currentEntity = tag.Value
			currentEntityClosed = false

			switch currentEntity {
			case "SPLINE":
				splineCount++
			case "LINE":
				lineCount++
			case "POLYLINE", "LWPOLYLINE":
				polylineCount++
			}
		}

		// Check closed flags for polylines and splines
		if inEntities && (currentEntity == "POLYLINE" || currentEntity == "LWPOLYLINE") {
			if tag.Code == 70 {
				flag, _ := strconv.Atoi(tag.Value)
				if (flag & 1) != 0 {
					currentEntityClosed = true
				}
			}
		}
		if inEntities && currentEntity == "SPLINE" {
			if tag.Code == 70 {
				flag, _ := strconv.Atoi(tag.Value)
				if (flag & 1) != 0 {
					currentEntityClosed = true
				}
			}
		}
	}

	// Final entity count
	if currentEntity != "" && inEntities {
		rep.EntityCounts[currentEntity]++
		rep.TotalEntities++
		if currentEntity == "POLYLINE" || currentEntity == "LWPOLYLINE" {
			if currentEntityClosed {
				rep.ClosedLoops++
			} else {
				rep.OpenCurves++
			}
		}
	}

	// Calculate bounds dimensions
	if rep.BoundsMinX < math.MaxFloat64 {
		rep.Width = rep.BoundsMaxX - rep.BoundsMinX
		rep.Height = rep.BoundsMaxY - rep.BoundsMinY
	} else {
		rep.BoundsMinX, rep.BoundsMinY = 0, 0
		rep.BoundsMaxX, rep.BoundsMaxY = 0, 0
		rep.Width, rep.Height = 0, 0
	}

	// Version naming
	switch rep.Version {
	case "AC1009":
		rep.VersionName = "AutoCAD Release 12 (AC1009)"
	case "AC1012":
		rep.VersionName = "AutoCAD Release 13 (AC1012)"
	case "AC1014":
		rep.VersionName = "AutoCAD Release 14 (AC1014)"
	case "AC1015":
		rep.VersionName = "AutoCAD 2000 (AC1015)"
	case "AC1018":
		rep.VersionName = "AutoCAD 2004 (AC1018)"
	case "AC1021":
		rep.VersionName = "AutoCAD 2007 (AC1021)"
	case "AC1024":
		rep.VersionName = "AutoCAD 2010 (AC1024)"
	case "AC1027":
		rep.VersionName = "AutoCAD 2013 (AC1027)"
	default:
		if rep.Version == "" {
			rep.Version = "Unspecified / R12"
			rep.VersionName = "Legacy / Unspecified"
		} else {
			rep.VersionName = "AutoCAD " + rep.Version
		}
	}

	// Units naming
	switch rep.InsUnitsCode {
	case 1:
		rep.Units = "Inches"
	case 4:
		rep.Units = "Millimeters"
	case 6:
		rep.Units = "Meters"
	default:
		rep.Units = "Unitless / Pixels"
	}

	// Compatibility Analysis & Diagnostics
	rep.CADCompatible = true

	// 1. Mandatory sections check
	hasSection := func(name string) bool {
		for _, s := range rep.Sections {
			if s == name {
				return true
			}
		}
		return false
	}

	if rep.Version == "AC1015" || strings.HasPrefix(rep.Version, "AC10") && rep.Version > "AC1014" {
		mandatory := []string{"CLASSES", "OBJECTS"}
		for _, m := range mandatory {
			if !hasSection(m) {
				rep.MissingSections = append(rep.MissingSections, m)
				rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
					Severity: "CRITICAL",
					Message:  fmt.Sprintf("Missing mandatory '%s' section for %s schema. Strict CAD importers (Onshape, ODA, SolidWorks) will reject this file as corrupt.", m, rep.Version),
				})
				rep.CADCompatible = false
			}
		}
		if rep.HandSeed == "" {
			rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
				Severity: "WARNING",
				Message:  "Missing $HANDSEED header variable in AutoCAD 2000+ database.",
			})
		}
	}

	// 2. Spline overload check
	if splineCount > 200 {
		rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
			Severity: "WARNING",
			Message:  fmt.Sprintf("File contains %d individual SPLINE entities. Large numbers of unjoined splines can severely lag or fail 2D constraint solvers in Onshape and Fusion 360 sketches.", splineCount),
		})
	}

	// 3. Entity connectivity / loops
	if polylineCount > 0 && rep.ClosedLoops > 0 {
		rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
			Severity: "SUCCESS",
			Message:  fmt.Sprintf("Found %d closed watertight loop(s). Ideal for 1-click sketch extrusion in Onshape, Fusion 360, and SolidWorks.", rep.ClosedLoops),
		})
	}

	if splineCount > 0 && polylineCount == 0 {
		rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
			Severity: "INFO",
			Message:  fmt.Sprintf("Entities are exported as cubic B-splines (%d). Supported by Fusion 360 and AutoCAD; for Onshape, export as Polylines (R12) for best sketch stability.", splineCount),
		})
	}

	if !hasHandles && rep.Version == "AC1015" {
		rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
			Severity: "WARNING",
			Message:  "Zero entity handles found in an AC1015 database.",
		})
	}

	return rep, nil
}

// FormatTerminalReport renders a formatted text diagnostic report.
func (r *Report) FormatTerminalReport() string {
	var sb strings.Builder
	sb.WriteString("======================================================================\n")
	sb.WriteString("   Text2SVG // DXF Health & CAD Compatibility Diagnostic\n")
	sb.WriteString("======================================================================\n")
	if r.Filename != "" {
		sb.WriteString(fmt.Sprintf("File:           %s\n", r.Filename))
	}
	sb.WriteString(fmt.Sprintf("Format Version: %s\n", r.VersionName))
	sb.WriteString(fmt.Sprintf("Units:          %s (INSUNITS = %d)\n", r.Units, r.InsUnitsCode))
	sb.WriteString(fmt.Sprintf("Dimensions:     %.2f × %.2f (%s)\n", r.Width, r.Height, r.Units))
	sb.WriteString(fmt.Sprintf("Bounding Box:   Min(%.2f, %.2f)  Max(%.2f, %.2f)\n", r.BoundsMinX, r.BoundsMinY, r.BoundsMaxX, r.BoundsMaxY))

	sb.WriteString("\n[SECTION ARCHITECTURE]\n")
	for _, sec := range r.Sections {
		sb.WriteString(fmt.Sprintf("  ✓ SECTION: %s\n", sec))
	}
	for _, miss := range r.MissingSections {
		sb.WriteString(fmt.Sprintf("  ✗ MISSING: %s (Required by %s)\n", miss, r.Version))
	}

	sb.WriteString("\n[ENTITY SUMMARY]\n")
	sb.WriteString(fmt.Sprintf("  Total Entities: %d\n", r.TotalEntities))
	for k, v := range r.EntityCounts {
		sb.WriteString(fmt.Sprintf("    - %-12s: %d\n", k, v))
	}
	if r.ClosedLoops > 0 || r.OpenCurves > 0 {
		sb.WriteString(fmt.Sprintf("  Loop Closures:  %d closed loops | %d open curves\n", r.ClosedLoops, r.OpenCurves))
	}

	sb.WriteString("\n[CAD COMPATIBILITY & HEALTH]\n")
	if len(r.Diagnostics) == 0 {
		sb.WriteString("  ✓ Clean DXF structure. No compatibility issues detected.\n")
	} else {
		for _, d := range r.Diagnostics {
			prefix := "  ℹ "
			switch d.Severity {
			case "CRITICAL":
				prefix = "  ✗ [CRITICAL] "
			case "WARNING":
				prefix = "  ⚠ [WARNING]  "
			case "SUCCESS":
				prefix = "  ✓ [SUCCESS]  "
			}
			sb.WriteString(prefix + d.Message + "\n")
		}
	}
	sb.WriteString("======================================================================\n")
	return sb.String()
}
