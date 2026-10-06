package converter

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// LoadedFont wraps sfnt.Font with convenience methods.
type LoadedFont struct {
	SFNT       *sfnt.Font
	UnitsPerEm int
	FontName   string
	Buffer     sfnt.Buffer
}

// ParseFont parses TTF or OTF font bytes into a LoadedFont.
func ParseFont(data []byte, name string) (*LoadedFont, error) {
	f, err := sfnt.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}
	units := int(f.UnitsPerEm())
	if units <= 0 {
		units = 2048 // fallback default
	}

	return &LoadedFont{
		SFNT:       f,
		UnitsPerEm: units,
		FontName:   name,
	}, nil
}

// RawGlyph contains extracted contours and segments for a single glyph at unit scale.
type RawGlyph struct {
	Advance  float64       // Advance width in font units
	Segments []PathSegment // Segments in Cartesian (Y points UP)
	Contours []Contour     // Sampled loops in Cartesian (Y points UP)
}

// LoadGlyph extracts vector paths and linearized contours for a single rune in font units (Y points UP).
func (lf *LoadedFont) LoadGlyph(r rune, curveSamples int) (*RawGlyph, error) {
	if curveSamples < 4 {
		curveSamples = 16
	}

	ppem := fixed.Int26_6(lf.UnitsPerEm) * 64
	idx, err := lf.SFNT.GlyphIndex(&lf.Buffer, r)
	if err != nil {
		return nil, fmt.Errorf("glyph index error for '%c': %w", r, err)
	}

	advFixed, err := lf.SFNT.GlyphAdvance(&lf.Buffer, idx, ppem, font.HintingNone)
	if err != nil {
		return nil, fmt.Errorf("glyph advance error for '%c': %w", r, err)
	}
	advance := float64(advFixed) / 64.0

	// Handle spaces or empty glyphs
	if r == ' ' || r == '\t' {
		return &RawGlyph{
			Advance:  advance,
			Segments: nil,
			Contours: nil,
		}, nil
	}

	segs, err := lf.SFNT.LoadGlyph(&lf.Buffer, idx, ppem, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load glyph '%c': %w", r, err)
	}

	var rawContours []Contour
	var pathSegments []PathSegment
	var currentContour Contour
	var currentPos Point

	// Helper to flip sfnt Y to Cartesian (Y points UP)
	flipPoint := func(p fixed.Point26_6) Point {
		return Point{
			X: float64(p.X) / 64.0,
			Y: -(float64(p.Y) / 64.0), // negate because sfnt Y points DOWN
		}
	}

	for _, seg := range segs {
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			// Finish previous contour if it exists
			if len(currentContour.Points) > 0 {
				currentContour.Closed = true
				rawContours = append(rawContours, currentContour)
				currentContour = Contour{}
			}
			dest := flipPoint(seg.Args[0])
			currentPos = dest
			currentContour.Points = append(currentContour.Points, dest)
			pathSegments = append(pathSegments, PathSegment{
				Type: SegmentMoveTo,
				Args: []Point{dest},
			})

		case sfnt.SegmentOpLineTo:
			dest := flipPoint(seg.Args[0])
			currentContour.Points = append(currentContour.Points, dest)
			currentPos = dest
			pathSegments = append(pathSegments, PathSegment{
				Type: SegmentLineTo,
				Args: []Point{dest},
			})

		case sfnt.SegmentOpQuadTo:
			ctrl := flipPoint(seg.Args[0])
			dest := flipPoint(seg.Args[1])
			// Sample curve for linearized polyline
			curvePts := SampleQuadBezier(currentPos, ctrl, dest, curveSamples)
			currentContour.Points = append(currentContour.Points, curvePts...)
			currentPos = dest
			pathSegments = append(pathSegments, PathSegment{
				Type: SegmentQuadTo,
				Args: []Point{ctrl, dest},
			})

		case sfnt.SegmentOpCubeTo:
			ctrl1 := flipPoint(seg.Args[0])
			ctrl2 := flipPoint(seg.Args[1])
			dest := flipPoint(seg.Args[2])
			// Sample curve for linearized polyline
			curvePts := SampleCubeBezier(currentPos, ctrl1, ctrl2, dest, curveSamples)
			currentContour.Points = append(currentContour.Points, curvePts...)
			currentPos = dest
			pathSegments = append(pathSegments, PathSegment{
				Type: SegmentCubeTo,
				Args: []Point{ctrl1, ctrl2, dest},
			})
		}
	}

	if len(currentContour.Points) > 0 {
		currentContour.Closed = true
		rawContours = append(rawContours, currentContour)
	}

	return &RawGlyph{
		Advance:  advance,
		Segments: pathSegments,
		Contours: rawContours,
	}, nil
}

// LayoutText transforms text into positioned vector geometry according to params.
func (lf *LoadedFont) LayoutText(params TextParams) (*RenderResult, error) {
	if params.Size <= 0 {
		params.Size = 20.0 // default 20mm
	}
	if params.LineHeight <= 0 {
		params.LineHeight = 1.2
	}
	if params.CurveSamples < 4 {
		params.CurveSamples = 16
	}
	if params.LayerName == "" {
		params.LayerName = "CUT"
	}
	if params.Datum == "" {
		params.Datum = DatumBottomLeft
	}

	scale := params.Size / float64(lf.UnitsPerEm)
	lineSpacing := params.Size * params.LineHeight
	extraTracking := params.Kerning // in output units

	lines := strings.Split(params.Text, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		lines = []string{" "}
	}

	var allContours []Contour
	var allSegments []PathSegment
	var glyphGroups []GlyphContourGroup
	glyphCount := 0

	// We lay out top-down in line order, with baseline Y starting at 0 for the first line
	// and decreasing by lineSpacing for subsequent lines.
	cursorY := 0.0

	for _, line := range lines {
		cursorX := 0.0
		runes := []rune(line)

		for _, r := range runes {
			glyph, err := lf.LoadGlyph(r, params.CurveSamples)
			if err != nil {
				// skip or fallback
				continue
			}
			glyphCount++

			var currentGlyphContours []Contour

			// Translate and scale contours
			for _, contour := range glyph.Contours {
				var transformedPts []Point
				for _, pt := range contour.Points {
					transformedPts = append(transformedPts, Point{
						X: cursorX + pt.X*scale,
						Y: cursorY + pt.Y*scale,
					})
				}
				c := Contour{
					Points: transformedPts,
					Closed: contour.Closed,
				}
				currentGlyphContours = append(currentGlyphContours, c)
				allContours = append(allContours, c)
			}

			if len(currentGlyphContours) > 0 {
				glyphGroups = append(glyphGroups, GlyphContourGroup{
					Contours: currentGlyphContours,
				})
			}

			// Translate and scale segments
			for _, seg := range glyph.Segments {
				var transformedArgs []Point
				for _, arg := range seg.Args {
					transformedArgs = append(transformedArgs, Point{
						X: cursorX + arg.X*scale,
						Y: cursorY + arg.Y*scale,
					})
				}
				allSegments = append(allSegments, PathSegment{
					Type: seg.Type,
					Args: transformedArgs,
				})
			}

			// Advance cursor
			cursorX += glyph.Advance*scale + extraTracking
		}

		cursorY -= lineSpacing
	}

	// 1. Apply Slant (§2.3) if requested
	if math.Abs(params.SlantAngle) > 1e-4 {
		allContours, allSegments = ApplySlant(allContours, allSegments, params.SlantAngle)
		for i := range glyphGroups {
			glyphGroups[i].Contours, _ = ApplySlant(glyphGroups[i].Contours, nil, params.SlantAngle)
		}
	}

	// 2. Apply Arc Deformation (§2.1) if requested
	if params.ArcEnabled {
		allContours = ApplyArcDeformation(allContours, params.ArcRadius, params.ArcSweep, params.ArcAlign, params.ArcInward)
		for i := range glyphGroups {
			glyphGroups[i].Contours = ApplyArcDeformation(glyphGroups[i].Contours, params.ArcRadius, params.ArcSweep, params.ArcAlign, params.ArcInward)
		}
		allSegments = nil // sampled deformed contours take precedence
	}

	// 3. Apply Path Welding (§1.2) if requested and we have glyph groups
	var classifiedPaths []ClassifiedContour
	isWelded := false
	if params.Weld && len(glyphGroups) > 0 {
		weldedContours, classified := WeldGlyphs(glyphGroups, 1e-4)
		if len(weldedContours) > 0 {
			allContours = weldedContours
			classifiedPaths = classified
			allSegments = nil // in welded mode, use the welded boundary contours
			isWelded = true
		}
	}

	if !isWelded && len(allContours) > 0 {
		classifiedPaths = ClassifyAndOrientContours(allContours, 1e-4)
	}

	// 4. Apply Inset / Offset (§2.2) if requested
	if math.Abs(params.Offset) > 1e-5 && len(classifiedPaths) > 0 {
		join := params.CornerJoin
		if join == "" {
			join = JoinRound
		}
		offsetContours := OffsetProfile(classifiedPaths, params.Offset, join, 3.0)
		if len(offsetContours) > 0 {
			allContours = offsetContours
			classifiedPaths = ClassifyAndOrientContours(allContours, 1e-4)
			allSegments = nil
		}
	}

	// Compute initial bounding box
	if len(allContours) == 0 {
		return &RenderResult{
			Contours:        nil,
			ClassifiedPaths: nil,
			Bounds:          BoundingBox{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10},
			SVG:             "",
			DXF:             "",
			GlyphCount:      0,
			PathCount:       0,
			Welded:          false,
			DXFFormat:       string(params.DXFFormat),
		}, nil
	}

	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9

	for _, c := range allContours {
		for _, p := range c.Points {
			if p.X < minX {
				minX = p.X
			}
			if p.X > maxX {
				maxX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
			if p.Y > maxY {
				maxY = p.Y
			}
		}
	}

	// Apply Datum offset
	var offsetX, offsetY float64
	switch params.Datum {
	case DatumBottomLeft:
		offsetX = -minX
		offsetY = -minY
	case DatumCenter:
		offsetX = -(minX + maxX) / 2.0
		offsetY = -(minY + maxY) / 2.0
	case DatumTopLeft:
		offsetX = -minX
		offsetY = -maxY
	default:
		offsetX = -minX
		offsetY = -minY
	}

	// Shift all points by datum offset
	for i := range allContours {
		for j := range allContours[i].Points {
			allContours[i].Points[j].X += offsetX
			allContours[i].Points[j].Y += offsetY
		}
	}
	for i := range classifiedPaths {
		for j := range classifiedPaths[i].Points {
			classifiedPaths[i].Points[j].X += offsetX
			classifiedPaths[i].Points[j].Y += offsetY
		}
	}
	for i := range allSegments {
		for j := range allSegments[i].Args {
			allSegments[i].Args[j].X += offsetX
			allSegments[i].Args[j].Y += offsetY
		}
	}

	// Recalculate bounds
	bounds := BoundingBox{
		MinX: minX + offsetX,
		MinY: minY + offsetY,
		MaxX: maxX + offsetX,
		MaxY: maxY + offsetY,
	}

	// Generate DXF and SVG strings
	dxfFormat := params.DXFFormat
	if dxfFormat == "" {
		dxfFormat = DXFFormatSpline
	}

	dxfStr := GenerateDXF(allContours, allSegments, dxfFormat, params.LayerName, params.Units)
	svgStr := GenerateSVG(allContours, allSegments, bounds, params.Units)

	return &RenderResult{
		Contours:        allContours,
		ClassifiedPaths: classifiedPaths,
		Bounds:          bounds,
		SVG:             svgStr,
		DXF:             dxfStr,
		GlyphCount:      glyphCount,
		PathCount:       len(allContours),
		Welded:          isWelded,
		DXFFormat:       string(dxfFormat),
	}, nil
}
