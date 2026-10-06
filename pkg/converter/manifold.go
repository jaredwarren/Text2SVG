package converter

import (
	"math"
	"sort"
)

// ContourRole specifies whether a closed contour represents an outer solid boundary or an inner hole.
type ContourRole string

const (
	RoleOuter ContourRole = "outer"
	RoleHole  ContourRole = "hole"
)

// ClassifiedContour represents a closed vector loop with topological classification.
type ClassifiedContour struct {
	Points   []Point     `json:"points"`
	Closed   bool        `json:"closed"`
	Role     ContourRole `json:"role"`
	Depth    int         `json:"depth"`     // Nesting depth: 0=outer, 1=hole, 2=island, etc.
	ParentID int         `json:"parent_id"` // Index of immediate enclosing contour (-1 if none)
	Area     float64     `json:"area"`      // Absolute area
}

// ClassifyAndOrientContours analyzes a list of closed contours, cleans vertices,
// determines hole/outer hierarchy, and enforces CAD standard orientation:
//   - Outer boundaries: Clockwise (CW) in Cartesian coordinates (Y up)
//   - Inner holes (counters): Counter-Clockwise (CCW) in Cartesian coordinates (Y up)
func ClassifyAndOrientContours(rawContours []Contour, tol float64) []ClassifiedContour {
	if tol <= 0 {
		tol = 1e-4
	}

	// 1. Clean and deduplicate each contour
	var cleaned []Contour
	for _, c := range rawContours {
		pts := CleanContour(c.Points, tol)
		if len(pts) >= 3 {
			cleaned = append(cleaned, Contour{
				Points: pts,
				Closed: true,
			})
		}
	}

	n := len(cleaned)
	if n == 0 {
		return nil
	}

	// Compute areas and bounding boxes
	areas := make([]float64, n)
	for i := 0; i < n; i++ {
		areas[i] = math.Abs(SignedArea(cleaned[i].Points))
	}

	// 2. Determine containment matrix
	// contains[j][i] is true if contour j geometrically contains contour i
	contains := make([][]bool, n)
	for i := 0; i < n; i++ {
		contains[i] = make([]bool, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			// j can only contain i if j's area is greater than i's area
			if areas[j] > areas[i] && IsContourInside(cleaned[i].Points, cleaned[j].Points) {
				contains[j][i] = true
			}
		}
	}

	// 3. Compute depth and immediate parent for each contour
	result := make([]ClassifiedContour, n)
	for i := 0; i < n; i++ {
		depth := 0
		parentID := -1
		minParentArea := 1e18

		for j := 0; j < n; j++ {
			if contains[j][i] {
				depth++
				if areas[j] < minParentArea {
					minParentArea = areas[j]
					parentID = j
				}
			}
		}

		isOuter := (depth % 2) == 0
		role := RoleOuter
		if !isOuter {
			role = RoleHole
		}

		// Enforce orientation:
		// Outer boundaries: Clockwise (CW) -> makeCW = true
		// Inner holes: Counter-Clockwise (CCW) -> makeCW = false
		orientedPts := EnsureOrientation(cleaned[i].Points, isOuter)

		result[i] = ClassifiedContour{
			Points:   orientedPts,
			Closed:   true,
			Role:     role,
			Depth:    depth,
			ParentID: parentID,
			Area:     areas[i],
		}
	}

	// Sort so outer boundaries come first, followed by holes (stable for CAD engines)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		return result[i].Area > result[j].Area
	})

	return result
}
