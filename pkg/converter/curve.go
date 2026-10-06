package converter

// SampleQuadBezier samples n points along a quadratic Bézier curve from p0 to p2 with control point p1.
// Note: does not include p0, but includes p2 as the final point.
func SampleQuadBezier(p0, p1, p2 Point, n int) []Point {
	if n < 2 {
		n = 2
	}
	pts := make([]Point, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		u := 1.0 - t
		tt := t * t
		uu := u * u

		x := uu*p0.X + 2*u*t*p1.X + tt*p2.X
		y := uu*p0.Y + 2*u*t*p1.Y + tt*p2.Y
		pts[i-1] = Point{X: x, Y: y}
	}
	return pts
}

// SampleCubeBezier samples n points along a cubic Bézier curve from p0 to p3 with control points p1, p2.
// Note: does not include p0, but includes p3 as the final point.
func SampleCubeBezier(p0, p1, p2, p3 Point, n int) []Point {
	if n < 2 {
		n = 2
	}
	pts := make([]Point, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		u := 1.0 - t
		uu := u * u
		tt := t * t
		uuu := uu * u
		ttt := tt * t

		x := uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X
		y := uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y
		pts[i-1] = Point{X: x, Y: y}
	}
	return pts
}

// QuadToCubic elevates a quadratic Bézier curve (p0, p1, p2) to an exact cubic Bézier curve (c0, c1, c2, c3).
func QuadToCubic(p0, p1, p2 Point) (Point, Point, Point, Point) {
	c0 := p0
	c1 := Point{
		X: p0.X + (2.0/3.0)*(p1.X-p0.X),
		Y: p0.Y + (2.0/3.0)*(p1.Y-p0.Y),
	}
	c2 := Point{
		X: p2.X + (2.0/3.0)*(p1.X-p2.X),
		Y: p2.Y + (2.0/3.0)*(p1.Y-p2.Y),
	}
	c3 := p2
	return c0, c1, c2, c3
}
