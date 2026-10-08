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
	DatumBottomLeft     Datum = "bottom-left"
	DatumCenter         Datum = "center"
	DatumTopLeft        Datum = "top-left"
	DatumBaselineLeft   Datum = "baseline-left"
	DatumBaselineCenter Datum = "baseline-center"
)

// Units defines the physical measurement unit.
type Units string

const (
	UnitsMM     Units = "mm"
	UnitsInches Units = "in"
	UnitsPixels Units = "px"
)

// DXFFormat defines the DXF export entity format.
type DXFFormat string

const (
	DXFFormatSpline   DXFFormat = "spline"   // AutoCAD 2000 (AC1015) true cubic B-SPLINE + LINE entities
	DXFFormatPolyline DXFFormat = "polyline" // Closed POLYLINE / LWPOLYLINE entities
)

// CornerJoin specifies how corners are shaped during contour offsetting.
type CornerJoin string

const (
	JoinRound CornerJoin = "round"
	JoinMiter CornerJoin = "miter"
	JoinBevel CornerJoin = "bevel"
)

// ArcAlignment defines alignment along a circular arc.
type ArcAlignment string

const (
	ArcAlignLeft   ArcAlignment = "left"
	ArcAlignCenter ArcAlignment = "center"
	ArcAlignRight  ArcAlignment = "right"
)

// BoundaryMode defines the outer perimeter enclosure type.
type BoundaryMode string

const (
	BoundaryNone      BoundaryMode = "none"
	BoundaryConformal BoundaryMode = "conformal" // Contour bubble tracing letter profiles
	BoundaryBox       BoundaryMode = "box"       // Rectangle / rounded-box / pill enclosure
)

// TextParams defines all parameters for text shaping and vector export.
type TextParams struct {
	Text         string    `json:"text"`
	Size         float64   `json:"size"`          // Target height in chosen units
	Units        Units     `json:"units"`         // mm, in, px (default mm)
	Kerning      float64   `json:"kerning"`       // Extra tracking/letter-spacing in chosen units
	LineHeight   float64   `json:"line_height"`   // Multiplier (e.g. 1.2)
	Datum        Datum     `json:"datum"`         // bottom-left, center, top-left, baseline-left, baseline-center
	CurveSamples int       `json:"curve_samples"` // Segments per curve in DXF (default 16)
	LayerName    string    `json:"layer_name"`    // DXF layer (e.g. "CUT", "ENGRAVE")
	Weld         bool      `json:"weld"`          // Path welding: boolean union for overlapping letters
	DXFFormat    DXFFormat `json:"dxf_format"`    // "spline" (AutoCAD 2000 SPLINE) or "polyline"

	// Phase 2: 3D Surfaces & Tolerances
	SlantAngle float64      `json:"slant_angle"` // Shear slant angle in degrees (-45° to +45°)
	Offset     float64      `json:"offset"`      // Inset/Offset in chosen units (+ expands, - contracts)
	CornerJoin CornerJoin   `json:"corner_join"` // "round", "miter", "bevel"
	ArcEnabled bool         `json:"arc_enabled"` // Deform text along circular arc
	ArcRadius  float64      `json:"arc_radius"`  // Radius of arc in chosen units
	ArcSweep   float64      `json:"arc_sweep"`   // Sweep angle in degrees (-360° to +360°)
	ArcAlign   ArcAlignment `json:"arc_align"`   // "center", "left", "right"
	ArcInward  bool         `json:"arc_inward"`  // Point letters toward center of curvature

	// Phase 3: Alignment & CAD Helpers
	ConstructionBox bool `json:"construction_box"` // Export reference bounding box wireframe

	// Phase 4: Perimeter Boundaries, Badges & Backing Plates
	BoundaryMode      BoundaryMode `json:"boundary_mode"`       // "none", "conformal", "box"
	BoundaryOffset    float64      `json:"boundary_offset"`     // Offset distance for conformal, or uniform padding for box
	BoundaryPaddingY  float64      `json:"boundary_padding_y"`  // Optional asymmetric Y padding for box (if <= 0, uses BoundaryOffset)
	BoundaryRadius    float64      `json:"boundary_radius"`     // Corner radius for box mode (0 = sharp, >= height/2 = pill/capsule)
	BoundaryFillHoles bool         `json:"boundary_fill_holes"` // True = solid outer silhouette (fill interior counter holes)
	BoundaryShiftX    float64      `json:"boundary_shift_x"`    // Manual X offset shift of boundary relative to text
	BoundaryShiftY    float64      `json:"boundary_shift_y"`    // Manual Y offset shift of boundary relative to text
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
	Contours        []Contour           `json:"contours"`
	ClassifiedPaths []ClassifiedContour `json:"classified_paths,omitempty"`
	Bounds          BoundingBox         `json:"bounds"`
	SVG             string              `json:"svg"`
	DXF             string              `json:"dxf"`
	GlyphCount      int                 `json:"glyph_count"`
	PathCount       int                 `json:"path_count"`
	Welded           bool                `json:"welded"`
	DXFFormat        string              `json:"dxf_format"`
	BoundaryContours []Contour           `json:"boundary_contours,omitempty"`
	BoundaryMode     string              `json:"boundary_mode,omitempty"`
}
