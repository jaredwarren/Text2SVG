package converter

// Point represents a 2D coordinate in real-world units (e.g. millimeters).
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BoundingBox holds min and max extents of geometry.
type BoundingBox struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

// Width returns the bounding box width.
func (b BoundingBox) Width() float64 {
	return b.MaxX - b.MinX
}

// Height returns the bounding box height.
func (b BoundingBox) Height() float64 {
	return b.MaxY - b.MinY
}

// Datum defines the origin point reference for exports.
type Datum string

const (
	DatumBottomLeft Datum = "bottom-left"
	DatumCenter     Datum = "center"
	DatumTopLeft    Datum = "top-left"
)

// Units defines the physical measurement unit.
type Units string

const (
	UnitsMM     Units = "mm"
	UnitsInches Units = "in"
	UnitsPixels Units = "px"
)

// TextParams defines all parameters for text shaping and vector export.
type TextParams struct {
	Text         string  `json:"text"`
	Size         float64 `json:"size"`          // Target height in chosen units
	Units        Units   `json:"units"`         // mm, in, px (default mm)
	Kerning      float64 `json:"kerning"`       // Extra tracking/letter-spacing in chosen units
	LineHeight   float64 `json:"line_height"`   // Multiplier (e.g. 1.2)
	Datum        Datum   `json:"datum"`         // bottom-left, center, top-left
	CurveSamples int     `json:"curve_samples"` // Segments per curve in DXF (default 16)
	LayerName    string  `json:"layer_name"`    // DXF layer (e.g. "CUT", "ENGRAVE")
}

// Contour represents a continuous path loop (sequence of points).
type Contour struct {
	Points []Point `json:"points"`
	Closed bool    `json:"closed"`
}

// PathSegment represents a high-level vector command for exact SVG/Bézier reproduction.
type PathSegmentType string

const (
	SegmentMoveTo PathSegmentType = "M"
	SegmentLineTo PathSegmentType = "L"
	SegmentQuadTo PathSegmentType = "Q"
	SegmentCubeTo PathSegmentType = "C"
	SegmentClose  PathSegmentType = "Z"
)

// PathSegment represents a single vector draw operation.
type PathSegment struct {
	Type PathSegmentType `json:"type"`
	Args []Point         `json:"args"`
}

// GlyphGeometry stores both the exact Bézier segments (for SVG) and linearized contours (for DXF).
type GlyphGeometry struct {
	Segments []PathSegment
	Contours []Contour
}

// RenderResult contains the final computed vectors, dimensions, and outputs.
type RenderResult struct {
	Contours   []Contour   `json:"contours"`
	Bounds     BoundingBox `json:"bounds"`
	SVG        string      `json:"svg"`
	DXF        string      `json:"dxf"`
	GlyphCount int         `json:"glyph_count"`
	PathCount  int         `json:"path_count"`
}
