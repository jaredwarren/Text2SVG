package dxf

import (
	"os"
	"strings"
	"testing"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
)

func TestInspectR12Polyline(t *testing.T) {
	// Generate a simple R12 DXF with closed contours
	c1 := converter.Contour{
		Points: []converter.Point{
			{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}, {X: 0, Y: 20}, {X: 0, Y: 0},
		},
		Closed: true,
	}
	dxfStr := converter.GenerateDXF([]converter.Contour{c1}, nil, converter.DXFFormatPolyline, "CUT", converter.UnitsMM, false, converter.BoundingBox{}, nil)

	rep, err := InspectReader(strings.NewReader(dxfStr))
	if err != nil {
		t.Fatalf("unexpected error inspecting DXF: %v", err)
	}

	if rep.Version != "AC1009" {
		t.Errorf("expected version AC1009, got %s", rep.Version)
	}
	if rep.ClosedLoops != 1 {
		t.Errorf("expected 1 closed loop, got %d", rep.ClosedLoops)
	}
	if rep.Width != 20.0 || rep.Height != 20.0 {
		t.Errorf("expected 20x20 dimensions, got %.2fx%.2f", rep.Width, rep.Height)
	}
	if !rep.CADCompatible {
		t.Errorf("expected R12 to be CAD compatible")
	}
}

func TestInspectAC1015Splines(t *testing.T) {
	// Generate an AC1015 DXF with Splines
	segments := []converter.PathSegment{
		{Type: converter.SegmentMoveTo, Args: []converter.Point{{X: 0, Y: 0}}},
		{Type: converter.SegmentLineTo, Args: []converter.Point{{X: 10, Y: 0}}},
		{Type: converter.SegmentQuadTo, Args: []converter.Point{{X: 15, Y: 5}, {X: 20, Y: 10}}},
		{Type: converter.SegmentClose, Args: nil},
	}
	dxfStr := converter.GenerateDXF(nil, segments, converter.DXFFormatSpline, "CUT", converter.UnitsMM, false, converter.BoundingBox{}, nil)

	rep, err := InspectReader(strings.NewReader(dxfStr))
	if err != nil {
		t.Fatalf("unexpected error inspecting AC1015 DXF: %v", err)
	}

	if rep.Version != "AC1015" {
		t.Errorf("expected version AC1015, got %s", rep.Version)
	}
	// Verify mandatory sections are present
	if len(rep.MissingSections) != 0 {
		t.Errorf("expected no missing sections, got: %v", rep.MissingSections)
	}
	if !rep.CADCompatible {
		t.Errorf("expected updated AC1015 to be CAD compatible")
	}
}

func TestInspectDownloadedFileIfPresent(t *testing.T) {
	path := "/Users/jaredwarren/Downloads/LASER_CUT_2026.dxf"
	if _, err := os.Stat(path); err != nil {
		t.Skip("downloaded test file not found on disk, skipping")
	}

	rep, err := InspectFile(path)
	if err != nil {
		t.Fatalf("failed to inspect file: %v", err)
	}

	t.Logf("Diagnosed %s:\n%s", path, rep.FormatTerminalReport())

	// Verify that the inspector correctly caught the missing sections in the downloaded file!
	foundClassesMissing := false
	for _, m := range rep.MissingSections {
		if m == "CLASSES" {
			foundClassesMissing = true
		}
	}
	if !foundClassesMissing {
		t.Errorf("expected inspector to catch missing CLASSES section in old downloaded file")
	}
	if rep.CADCompatible {
		t.Errorf("expected old downloaded file to be flagged as not CADCompatible")
	}
}
