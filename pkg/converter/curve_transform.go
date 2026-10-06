package converter

import (
	"math"
)

// ApplySlant applies an affine shear transform along the X axis relative to baseline y=0.
// Positive slant angle in degrees slants the top of the text forward (italic).
func ApplySlant(contours []Contour, segments []PathSegment, angleDeg float64) ([]Contour, []PathSegment) {
	if math.Abs(angleDeg) < 1e-6 {
		return contours, segments
	}

	rad := angleDeg * (math.Pi / 180.0)
	tanAngle := math.Tan(rad)

	// Transform contours
	slantedContours := make([]Contour, len(contours))
	for i, c := range contours {
		pts := make([]Point, len(c.Points))
		for j, p := range c.Points {
			pts[j] = Point{
				X: p.X + p.Y*tanAngle,
				Y: p.Y,
			}
		}
		slantedContours[i] = Contour{
			Points: pts,
			Closed: c.Closed,
		}
	}

	// Transform segments
	var slantedSegments []PathSegment
	if len(segments) > 0 {
		slantedSegments = make([]PathSegment, len(segments))
		for i, seg := range segments {
			args := make([]Point, len(seg.Args))
			for j, p := range seg.Args {
				args[j] = Point{
					X: p.X + p.Y*tanAngle,
					Y: p.Y,
				}
			}
			slantedSegments[i] = PathSegment{
				Type: seg.Type,
				Args: args,
			}
		}
	}

	return slantedContours, slantedSegments
}

// SubdivideContour breaks long straight segments into shorter segments so they bend
// smoothly when mapped onto a curved surface.
func SubdivideContour(pts []Point, maxLen float64) []Point {
	if len(pts) < 2 || maxLen <= 0 {
		return pts
	}

	var res []Point
	n := len(pts)
	for i := 0; i < n; i++ {
		p0 := pts[i]
		p1 := pts[(i+1)%n]
		res = append(res, p0)

		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > maxLen {
			steps := int(math.Ceil(dist / maxLen))
			for s := 1; s < steps; s++ {
				t := float64(s) / float64(steps)
				res = append(res, Point{
					X: p0.X + t*dx,
					Y: p0.Y + t*dy,
				})
			}
		}
	}
	return res
}

// ApplyArcDeformation maps flat 2D contours along a circular arc of specified radius.
// If arcSweep is 0, sweep angle is automatically calculated from text width and radius.
func ApplyArcDeformation(contours []Contour, radius float64, sweepDeg float64, align ArcAlignment, inward bool) []Contour {
	if len(contours) == 0 {
		return contours
	}
	if radius <= 0 {
		radius = 50.0 // Default 50mm radius
	}

	// 1. Find bounding box of flat text
	b := ContourBoundingBox(nil)
	initialized := false
	for _, c := range contours {
		for _, p := range c.Points {
			if !initialized {
				b = BoundingBox{MinX: p.X, MinY: p.Y, MaxX: p.X, MaxY: p.Y}
				initialized = true
			} else {
				if p.X < b.MinX {
					b.MinX = p.X
				}
				if p.X > b.MaxX {
					b.MaxX = p.X
				}
				if p.Y < b.MinY {
					b.MinY = p.Y
				}
				if p.Y > b.MaxY {
					b.MaxY = p.Y
				}
			}
		}
	}

	textWidth := b.Width()
	if textWidth <= 0 {
		textWidth = 10.0
	}

	// 2. Determine sweep angle in radians
	var sweepRad float64
	if math.Abs(sweepDeg) > 1e-4 {
		sweepRad = sweepDeg * (math.Pi / 180.0)
	} else {
		// Natural arc angle for width W at radius R: theta = W / R
		sweepRad = textWidth / radius
	}

	// Maximum straight segment length before deformation: ~3 degrees of arc
	maxSegLen := math.Max(0.5, radius*(3.0*math.Pi/180.0))

	// 3. Alignment reference X
	var refX float64
	switch align {
	case ArcAlignLeft:
		refX = b.MinX
	case ArcAlignRight:
		refX = b.MaxX
	case ArcAlignCenter:
		fallthrough
	default:
		refX = (b.MinX + b.MaxX) / 2.0
	}

	// Map point (x, y) to curved coordinate (x', y')
	mapPoint := func(p Point) Point {
		// Arc-length position s along baseline
		var s float64
		if align == ArcAlignLeft {
			s = (p.X - refX) / textWidth
			// theta from 0 to sweepRad
			theta := s * sweepRad
			return polarToCartesian(p, radius, theta, inward)
		} else if align == ArcAlignRight {
			s = (p.X - refX) / textWidth
			theta := s * sweepRad
			return polarToCartesian(p, radius, theta, inward)
		} else {
			// Center align: -0.5 to +0.5
			s = (p.X - refX) / textWidth
			theta := s * sweepRad
			return polarToCartesian(p, radius, theta, inward)
		}
	}

	// Deform contours
	curvedContours := make([]Contour, len(contours))
	for i, c := range contours {
		subdivided := SubdivideContour(c.Points, maxSegLen)
		deformedPts := make([]Point, len(subdivided))
		for j, pt := range subdivided {
			deformedPts[j] = mapPoint(pt)
		}
		curvedContours[i] = Contour{
			Points: CleanContour(deformedPts, 1e-4),
			Closed: c.Closed,
		}
	}

	return curvedContours
}

// polarToCartesian computes conformal arc coordinates.
func polarToCartesian(p Point, radius, theta float64, inward bool) Point {
	var r float64
	var ySign float64 = 1.0

	if !inward {
		// Outward (convex arch over top of circle)
		r = radius + p.Y
		xPrime := r * math.Sin(theta)
		yPrime := r*math.Cos(theta) - radius
		return Point{X: xPrime, Y: yPrime}
	} else {
		// Inward (concave arch pointing toward center of curvature)
		r = radius - p.Y
		ySign = -1.0
		xPrime := r * math.Sin(theta)
		yPrime := ySign * (r*math.Cos(theta) - radius)
		return Point{X: xPrime, Y: yPrime}
	}
}
