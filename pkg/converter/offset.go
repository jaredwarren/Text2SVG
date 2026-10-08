package converter

import (
	"fmt"
	"math"
	"os"

	"github.com/lestrrat-go/polyclip"
	"github.com/lestrrat-go/polyclip/geom"
)

// OffsetContour offsets a single closed contour by distance d.
// Positive d expands the region enclosed by the ring; negative d shrinks it.
// The returned ring keeps the input winding. An offset that collapses the
// region (over-inset) returns nil.
func OffsetContour(pts []Point, d float64, join CornerJoin, miterLimit float64) []Point {
	pts = CleanContour(pts, 1e-5)
	if len(pts) < 3 || math.Abs(d) < 1e-9 {
		return pts
	}

	wantCW := IsClockwise(pts)
	// The offset engine's outward normal is the right-hand normal, so a
	// positive distance grows a counter-clockwise ring and shrinks a
	// clockwise one. Feed it CCW, then restore the caller's winding.
	out := offsetSolid(geom.MultiPolygon{{Outer: toRing(EnsureOrientation(pts, false))}}, d, join, miterLimit)
	best := largestRing(out)
	if len(best) < 3 {
		return nil
	}
	return EnsureOrientation(best, wantCW)
}

// OffsetProfile applies contour offsetting to a classified manifold region.
// Positive d expands solid material: outer boundaries grow and holes shrink.
// Negative d does the opposite. Pieces that collide are unioned; a piece that
// shrinks away is dropped.
func OffsetProfile(classified []ClassifiedContour, d float64, join CornerJoin, miterLimit float64) []Contour {
	if len(classified) == 0 {
		return nil
	}
	if math.Abs(d) < 1e-5 {
		res := make([]Contour, len(classified))
		for i, c := range classified {
			res[i] = Contour{Points: c.Points, Closed: true}
		}
		return res
	}

	raw := make([]Contour, 0, len(classified))
	for _, c := range classified {
		if len(c.Points) < 3 {
			continue
		}
		raw = append(raw, Contour{Points: c.Points, Closed: true})
	}
	out := offsetSolid(materialMultiPolygon(raw, false), d, join, miterLimit)
	return multiToContours(out)
}

// materialMultiPolygon builds the filled text region. When fillHoles is set,
// only the outermost rings are kept, so counters become solid. Otherwise holes
// stay attached to the solid that contains them, and islands inside holes are
// separate pieces.
func materialMultiPolygon(contours []Contour, fillHoles bool) geom.MultiPolygon {
	classified := ClassifyAndOrientContours(contours, 1e-4)
	if len(classified) == 0 {
		return nil
	}

	type solid struct {
		pts   []Point
		holes [][]Point
		depth int
		area  float64
	}
	var solids []solid
	var holes []ClassifiedContour
	for _, c := range classified {
		if len(c.Points) < 3 || c.Area < 1e-8 {
			continue
		}
		if c.Depth%2 == 0 {
			if fillHoles && c.Depth != 0 {
				continue
			}
			solids = append(solids, solid{pts: c.Points, depth: c.Depth, area: c.Area})
			continue
		}
		if !fillHoles {
			holes = append(holes, c)
		}
	}

	for _, h := range holes {
		best := -1
		bestArea := math.Inf(1)
		for i, s := range solids {
			if s.depth != h.Depth-1 || s.area <= h.Area || s.area >= bestArea {
				continue
			}
			if IsContourInside(h.Points, s.pts) {
				best = i
				bestArea = s.area
			}
		}
		if best >= 0 {
			solids[best].holes = append(solids[best].holes, h.Points)
		}
	}

	mp := make(geom.MultiPolygon, 0, len(solids))
	for _, s := range solids {
		ex := geom.ExPolygon{Outer: toRing(s.pts)}
		for _, h := range s.holes {
			ex.Holes = append(ex.Holes, toRing(h))
		}
		if len(ex.Outer) >= 3 {
			mp = append(mp, ex)
		}
	}
	return mp
}

// offsetSolid inflates (d > 0) or erodes (d < 0) a filled region and unions
// any pieces that grow into each other.
//
// Pipeline:
//  1. Simplify self-crossing font outlines into a proper filled region.
//  2. Chamfer sharp reflex corners so outward offsets cannot grow needles
//     (JoinRound only fills convex gaps; reflex hairpins still miter).
//  3. Delegate the parallel-curve + topology resolve to polyclip.Offset,
//     which handles outer/hole winding natively.
//  4. Union colliding pieces, then pull any remaining extreme spikes back
//     onto the d-radius. We never delete far vertices — that leaves inward
//     chords (the classic V-notch artifact).
func offsetSolid(mp geom.MultiPolygon, d float64, join CornerJoin, miterLimit float64) geom.MultiPolygon {
	if len(mp) == 0 {
		return nil
	}
	if math.Abs(d) < 1e-9 {
		return mp
	}
	opts := offsetOptions(d, join, miterLimit)
	reach := offsetReach(d, join, miterLimit)
	source := mp

	// Font outlines sometimes cross themselves (engraved and script faces).
	// Resolve that to the filled region before offsetting, or the parallel
	// curve follows the crossing and collapses.
	if cleaned, err := polyclip.Simplify(orientForClip(mp)); err == nil && len(cleaned) > 0 {
		mp = cleaned
		source = cleaned
	}
	if os.Getenv("OFFSET_DEBUG") == "1" {
		for i, ex := range mp {
			fmt.Fprintf(os.Stderr, "piece %d area=%.2f holes=%d pts=%d\n", i, ex.Outer.Area(), len(ex.Holes), len(ex.Outer))
		}
	}

	// Offset each piece's outer as a solid, then punch counters back out.
	// polyclip.Offset's built-in hole handling expands CW holes under +d in
	// practice (opposite of the material-expand semantics we need), so we
	// keep holes as separate CCW solids offset by -d and Difference them.
	parts := make([]geom.MultiPolygon, 0, len(mp))
	for _, ex := range mp {
		if len(ex.Outer) < 3 {
			continue
		}
		outerRing := chamferSharpReflex(ccwRing(ex.Outer), d, reach)
		outer, err := polyclip.Offset(geom.MultiPolygon{{Outer: outerRing}}, d, opts)
		if os.Getenv("OFFSET_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "  offset d=%.2f err=%v pieces=%d\n", d, err, len(outer))
		}
		if err != nil || len(outer) == 0 {
			continue
		}

		var voids []geom.MultiPolygon
		for _, h := range ex.Holes {
			if len(h) < 3 {
				continue
			}
			holeRing := chamferSharpReflex(ccwRing(h), -d, reach)
			void, err := polyclip.Offset(geom.MultiPolygon{{Outer: holeRing}}, -d, opts)
			if err != nil || len(void) == 0 {
				continue
			}
			voids = append(voids, void)
		}
		piece := outer
		if len(voids) > 0 {
			if voidU, err := polyclip.UnionAll(voids...); err == nil && len(voidU) > 0 {
				if punched, err := polyclip.Difference(outer, voidU); err == nil {
					piece = punched
				}
			}
		}
		if len(piece) > 0 {
			parts = append(parts, piece)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	merged, err := polyclip.UnionAll(parts...)
	if err != nil {
		var flat geom.MultiPolygon
		for _, p := range parts {
			flat = append(flat, p...)
		}
		merged = flat
	}
	if simplified, err := polyclip.Simplify(merged); err == nil && len(simplified) > 0 {
		merged = simplified
	}

	// Pull runaway reflex needles back onto the d-circle. For miter joins the
	// apex may legitimately reach miterLimit·|d|, so the factor follows reach.
	spikeFactor := reach / math.Abs(d)
	if spikeFactor < 1.15 {
		spikeFactor = 1.15
	}
	merged = projectSpikeVertices(merged, source, math.Abs(d), spikeFactor)
	if simplified, err := polyclip.Simplify(merged); err == nil && len(simplified) > 0 {
		if os.Getenv("OFFSET_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "  afterProject=%.1f\n", simplified.Area())
		}
		merged = simplified
	}
	return merged
}

// offsetReach is the farthest a reflex-corner apex may sit from the source
// before we pre-chamfer that corner. Round/bevel joins should stay near |d|;
// miter joins may reach miterLimit·|d|.
func offsetReach(d float64, join CornerJoin, miterLimit float64) float64 {
	ad := math.Abs(d)
	slack := math.Max(0.05, ad*0.03)
	if join == JoinMiter {
		if miterLimit < 1 {
			miterLimit = 3
		}
		return ad*miterLimit + slack
	}
	// Round/bevel vertices sit near |d|; chamfer any reflex apex that would
	// poke meaningfully past that envelope.
	return ad*1.05 + slack
}

// chamferSharpReflex replaces reflex corners whose offset apex would shoot
// past reach with a short bevel, so the offset cannot grow a needle there.
// ring must be counter-clockwise. d is the signed distance applied to ring.
func chamferSharpReflex(ring geom.Polygon, d, reach float64) geom.Polygon {
	n := len(ring)
	if n < 3 || math.Abs(d) < 1e-9 {
		return ring
	}
	pts := make([]Point, n)
	for i, p := range ring {
		pts[i] = Point{X: p.X, Y: p.Y}
	}
	var out []Point
	for i := 0; i < n; i++ {
		prev := pts[(i+n-1)%n]
		cur := pts[i]
		next := pts[(i+1)%n]
		v1x, v1y := cur.X-prev.X, cur.Y-prev.Y
		v2x, v2y := next.X-cur.X, next.Y-cur.Y
		l1 := math.Hypot(v1x, v1y)
		l2 := math.Hypot(v2x, v2y)
		if l1 < 1e-9 || l2 < 1e-9 {
			continue
		}
		d1x, d1y := v1x/l1, v1y/l1
		d2x, d2y := v2x/l2, v2y/l2
		// Right-hand normals, matching the offset engine.
		n1x, n1y := d1y, -d1x
		n2x, n2y := d2y, -d2x
		cross := n1x*n2y - n1y*n2x
		// Positive cross·d is a convex corner (the join fills a gap).
		// Reflex corners are the ones that grow an unbounded apex.
		if cross*d > 1e-5 {
			out = append(out, cur)
			continue
		}
		dot := n1x*n2x + n1y*n2y
		denom := 1 + dot
		apex := math.Inf(1)
		if denom > 1e-8 {
			apex = math.Abs(d) * math.Hypot(n1x+n2x, n1y+n2y) / denom
		}
		if apex <= reach {
			out = append(out, cur)
			continue
		}
		cut := math.Min(math.Min(l1, l2)*0.45, math.Max(math.Abs(d), 1e-3))
		if cut < 1e-6 {
			out = append(out, cur)
			continue
		}
		out = append(out,
			Point{X: cur.X - d1x*cut, Y: cur.Y - d1y*cut},
			Point{X: cur.X + d2x*cut, Y: cur.Y + d2y*cut},
		)
	}
	out = CleanContour(out, 1e-5)
	if len(out) < 3 {
		return ring
	}
	ringOut := make(geom.Polygon, len(out))
	for i, p := range out {
		ringOut[i] = geom.Point{X: p.X, Y: p.Y}
	}
	return ringOut
}

// projectSpikeVertices pulls offset vertices that landed farther than
// factor·|d| from the source boundary back onto the d-circle around the
// nearest source point. Unlike deleting those vertices (which leaves an
// inward chord / V-notch), projection preserves local topology.
func projectSpikeVertices(m, source geom.MultiPolygon, d, factor float64) geom.MultiPolygon {
	if len(m) == 0 || d <= 0 || factor <= 1 {
		return m
	}
	var segs [][2]geom.Point
	add := func(ring geom.Polygon) {
		n := len(ring)
		for i := 0; i < n; i++ {
			segs = append(segs, [2]geom.Point{ring[i], ring[(i+1)%n]})
		}
	}
	for _, ex := range source {
		add(ex.Outer)
		for _, h := range ex.Holes {
			add(h)
		}
	}
	if len(segs) == 0 {
		return m
	}
	limit := d * factor
	project := func(ring geom.Polygon) geom.Polygon {
		out := make(geom.Polygon, len(ring))
		for i, p := range ring {
			cp, dist := closestOnSegs(p, segs)
			if dist > limit && dist > 1e-9 {
				ux := (p.X - cp.X) / dist
				uy := (p.Y - cp.Y) / dist
				out[i] = geom.Point{X: cp.X + ux*d, Y: cp.Y + uy*d}
			} else {
				out[i] = p
			}
		}
		return out
	}
	out := make(geom.MultiPolygon, 0, len(m))
	for _, ex := range m {
		outer := project(ex.Outer)
		if len(outer) < 3 {
			continue
		}
		piece := geom.ExPolygon{Outer: outer}
		for _, h := range ex.Holes {
			hr := project(h)
			if len(hr) >= 3 {
				piece.Holes = append(piece.Holes, hr)
			}
		}
		out = append(out, piece)
	}
	return out
}

// closestOnSegs returns the nearest point on any segment and its distance.
func closestOnSegs(p geom.Point, segs [][2]geom.Point) (geom.Point, float64) {
	best := geom.Point{}
	bestD := math.Inf(1)
	for _, s := range segs {
		a, b := s[0], s[1]
		dx := b.X - a.X
		dy := b.Y - a.Y
		l2 := dx*dx + dy*dy
		t := 0.0
		if l2 > 1e-18 {
			t = ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / l2
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
		}
		q := geom.Point{X: a.X + t*dx, Y: a.Y + t*dy}
		e := math.Hypot(p.X-q.X, p.Y-q.Y)
		if e < bestD {
			bestD = e
			best = q
		}
	}
	return best, bestD
}

// orientForClip makes outers counter-clockwise and holes clockwise, which
// is the nonzero-winding convention Simplify uses.
func orientForClip(mp geom.MultiPolygon) geom.MultiPolygon {
	out := make(geom.MultiPolygon, len(mp))
	for i, ex := range mp {
		out[i].Outer = ccwRing(ex.Outer)
		for _, h := range ex.Holes {
			hr := append(geom.Polygon(nil), h...)
			if hr.SignedArea() > 0 {
				hr.Reverse()
			}
			out[i].Holes = append(out[i].Holes, hr)
		}
	}
	return out
}

// ccwRing returns a copy of ring wound counter-clockwise.
func ccwRing(ring geom.Polygon) geom.Polygon {
	out := append(geom.Polygon(nil), ring...)
	if out.SignedArea() < 0 {
		out.Reverse()
	}
	return out
}

func offsetOptions(d float64, join CornerJoin, miterLimit float64) polyclip.OffsetOptions {
	if miterLimit <= 1 {
		miterLimit = 3
	}
	arcTol := math.Abs(d) * 0.01
	if arcTol < 1e-4 {
		arcTol = 1e-4
	}
	return polyclip.OffsetOptions{
		Join:       mapJoin(join),
		MiterLimit: miterLimit,
		ArcTol:     arcTol,
	}
}

func mapJoin(join CornerJoin) polyclip.JoinType {
	switch join {
	case JoinRound:
		return polyclip.JoinRound
	case JoinBevel:
		return polyclip.JoinBevel
	case JoinMiter:
		return polyclip.JoinMiter
	default:
		return polyclip.JoinRound
	}
}

func toRing(pts []Point) geom.Polygon {
	pts = CleanContour(pts, 1e-5)
	ring := make(geom.Polygon, 0, len(pts))
	for _, p := range pts {
		ring = append(ring, geom.Point{X: p.X, Y: p.Y})
	}
	return ring
}

func fromRing(ring geom.Polygon) []Point {
	if len(ring) < 3 {
		return nil
	}
	pts := make([]Point, len(ring))
	for i, p := range ring {
		pts[i] = Point{X: p.X, Y: p.Y}
	}
	return CleanContour(pts, 1e-4)
}

func multiToContours(m geom.MultiPolygon) []Contour {
	var out []Contour
	for _, ex := range m {
		if pts := fromRing(ex.Outer); len(pts) >= 3 && math.Abs(SignedArea(pts)) > 1e-4 {
			out = append(out, Contour{Points: pts, Closed: true})
		}
		for _, h := range ex.Holes {
			if pts := fromRing(h); len(pts) >= 3 && math.Abs(SignedArea(pts)) > 1e-4 {
				out = append(out, Contour{Points: pts, Closed: true})
			}
		}
	}
	return out
}

func largestRing(m geom.MultiPolygon) []Point {
	var best []Point
	bestArea := 0.0
	for _, c := range multiToContours(m) {
		a := math.Abs(SignedArea(c.Points))
		if a > bestArea {
			best = c.Points
			bestArea = a
		}
	}
	return best
}
