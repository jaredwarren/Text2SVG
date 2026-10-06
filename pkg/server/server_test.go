package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
	"github.com/jaredwarren/Text2SVG/pkg/fonts"
	"github.com/jaredwarren/Text2SVG/web"
)

func TestServerPreviewAndExportPhase2(t *testing.T) {
	fm := fonts.NewManager()
	s := NewServer(fm, web.AssetFS())
	handler := s.Routes()

	// 1. Test Fonts listing
	reqList := httptest.NewRequest("GET", "/api/fonts", nil)
	recList := httptest.NewRecorder()
	handler.ServeHTTP(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/fonts, got %d", recList.Code)
	}

	var fontList []fonts.FontInfo
	if err := json.NewDecoder(recList.Body).Decode(&fontList); err != nil {
		t.Fatalf("failed to decode font list: %v", err)
	}
	if len(fontList) == 0 {
		t.Skip("no fonts installed on system, skipping server convert tests")
	}

	firstFontID := fontList[0].ID

	// 2. Test Preview with Phase 2 options (Slant, Offset, Arc)
	payload := ConvertRequest{
		Text:       "CURVE & SLANT",
		FontID:     firstFontID,
		Size:       30.0,
		Units:      converter.UnitsMM,
		SlantAngle: 15.0,
		Offset:     0.2,
		CornerJoin: converter.JoinRound,
		ArcEnabled: true,
		ArcRadius:  60.0,
		ArcSweep:   0.0,
		ArcAlign:   converter.ArcAlignCenter,
		ArcInward:  false,
		Weld:       true,
		DXFFormat:  converter.DXFFormatSpline,
	}

	bodyBytes, _ := json.Marshal(payload)
	reqPrev := httptest.NewRequest("POST", "/api/preview", bytes.NewReader(bodyBytes))
	reqPrev.Header.Set("Content-Type", "application/json")
	recPrev := httptest.NewRecorder()
	handler.ServeHTTP(recPrev, reqPrev)

	if recPrev.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/preview, got %d: %s", recPrev.Code, recPrev.Body.String())
	}

	var prevResp ConvertResponse
	if err := json.NewDecoder(recPrev.Body).Decode(&prevResp); err != nil {
		t.Fatalf("failed to decode preview response: %v", err)
	}

	if prevResp.SVG == "" {
		t.Errorf("expected non-empty SVG")
	}
	if prevResp.Width <= 0 || prevResp.Height <= 0 {
		t.Errorf("expected positive dimensions, got width=%.2f, height=%.2f", prevResp.Width, prevResp.Height)
	}
	if prevResp.CLICommand == "" {
		t.Errorf("expected generated CLI command")
	}

	// 3. Test Export DXF
	reqDXF := httptest.NewRequest("POST", "/api/export/dxf", bytes.NewReader(bodyBytes))
	reqDXF.Header.Set("Content-Type", "application/json")
	recDXF := httptest.NewRecorder()
	handler.ServeHTTP(recDXF, reqDXF)

	if recDXF.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/export/dxf, got %d", recDXF.Code)
	}
	if recDXF.Body.Len() == 0 {
		t.Errorf("expected DXF content")
	}

	// 4. Test Export SVG
	reqSVG := httptest.NewRequest("POST", "/api/export/svg", bytes.NewReader(bodyBytes))
	reqSVG.Header.Set("Content-Type", "application/json")
	recSVG := httptest.NewRecorder()
	handler.ServeHTTP(recSVG, reqSVG)

	if recSVG.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/export/svg, got %d", recSVG.Code)
	}
	if recSVG.Body.Len() == 0 {
		t.Errorf("expected SVG content")
	}

	// 5. Test Inspect DXF endpoint with the DXF generated in step 3
	reqInspect := httptest.NewRequest("POST", "/api/inspect-dxf", bytes.NewReader(recDXF.Body.Bytes()))
	reqInspect.Header.Set("Content-Type", "application/dxf")
	recInspect := httptest.NewRecorder()
	handler.ServeHTTP(recInspect, reqInspect)

	if recInspect.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/inspect-dxf, got %d: %s", recInspect.Code, recInspect.Body.String())
	}
}
