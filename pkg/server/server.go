package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
	"github.com/jaredwarren/Text2SVG/pkg/dxf"
	"github.com/jaredwarren/Text2SVG/pkg/fonts"
)

type Server struct {
	fontManager *fonts.Manager
	fs          http.FileSystem
}

func NewServer(fm *fonts.Manager, fs http.FileSystem) *Server {
	return &Server{
		fontManager: fm,
		fs:          fs,
	}
}

// ConvertRequest defines the incoming payload for preview and export.
type ConvertRequest struct {
	Text         string                 `json:"text"`
	FontID       string                 `json:"font_id"`
	Size         float64                `json:"size"`
	Units        converter.Units        `json:"units"`
	Kerning      float64                `json:"kerning"`
	LineHeight   float64                `json:"line_height"`
	Datum        converter.Datum        `json:"datum"`
	CurveSamples int                    `json:"curve_samples"`
	LayerName    string                 `json:"layer_name"`
	Weld         bool                   `json:"weld"`
	DXFFormat    converter.DXFFormat    `json:"dxf_format"`
	SlantAngle   float64                `json:"slant_angle"`
	Offset       float64                `json:"offset"`
	CornerJoin   converter.CornerJoin   `json:"corner_join"`
	ArcEnabled   bool                   `json:"arc_enabled"`
	ArcRadius    float64                `json:"arc_radius"`
	ArcSweep        float64                `json:"arc_sweep"`
	ArcAlign        converter.ArcAlignment `json:"arc_align"`
	ArcInward       bool                   `json:"arc_inward"`
	ConstructionBox bool                   `json:"construction_box"`
}

// ConvertResponse is the response returned to the Web UI.
type ConvertResponse struct {
	SVG        string                `json:"svg"`
	Bounds     converter.BoundingBox `json:"bounds"`
	Width      float64               `json:"width"`
	Height     float64               `json:"height"`
	Units      string                `json:"units"`
	GlyphCount int                   `json:"glyph_count"`
	PathCount  int                   `json:"path_count"`
	Welded     bool                  `json:"welded"`
	DXFFormat  string                `json:"dxf_format"`
	CLICommand string                `json:"cli_command"`
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/fonts", s.handleListFonts)
	mux.HandleFunc("POST /api/upload-font", s.handleUploadFont)
	mux.HandleFunc("POST /api/preview", s.handlePreview)
	mux.HandleFunc("POST /api/export/dxf", s.handleExportDXF)
	mux.HandleFunc("POST /api/export/svg", s.handleExportSVG)
	mux.HandleFunc("POST /api/inspect-dxf", s.handleInspectDXF)

	// Static asset handler
	fileServer := http.FileServer(s.fs)
	mux.Handle("/", fileServer)

	return mux
}

func (s *Server) handleListFonts(w http.ResponseWriter, r *http.Request) {
	fontsList := s.fontManager.ListFonts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fontsList)
}

func (s *Server) handleUploadFont(w http.ResponseWriter, r *http.Request) {
	// 32MB max font upload
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("font")
	if err != nil {
		http.Error(w, "Font file missing in form field 'font'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read font file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	loaded, err := s.fontManager.RegisterFont(header.Filename, data, "Uploaded")
	if err != nil {
		http.Error(w, "Failed to parse uploaded font: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      strings.ToLower(strings.ReplaceAll(loaded.FontName, " ", "-")),
		"name":    loaded.FontName,
	})
}

func (s *Server) parseRequest(r *http.Request) (*ConvertRequest, *converter.LoadedFont, error) {
	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, nil, fmt.Errorf("invalid json: %w", err)
	}

	if req.Text == "" {
		req.Text = "Text2SVG"
	}
	if req.Size <= 0 {
		req.Size = 25.0
	}
	if req.LineHeight <= 0 {
		req.LineHeight = 1.2
	}
	if req.Units == "" {
		req.Units = converter.UnitsMM
	}
	if req.Datum == "" {
		req.Datum = converter.DatumBottomLeft
	}
	if req.CurveSamples <= 0 {
		req.CurveSamples = 16
	}
	if req.LayerName == "" {
		req.LayerName = "CUT"
	}
	if req.DXFFormat == "" {
		req.DXFFormat = converter.DXFFormatSpline
	}
	if req.CornerJoin == "" {
		req.CornerJoin = converter.JoinRound
	}
	if req.ArcRadius <= 0 {
		req.ArcRadius = 50.0
	}
	if req.ArcAlign == "" {
		req.ArcAlign = converter.ArcAlignCenter
	}

	loadedFont, err := s.fontManager.GetFont(req.FontID)
	if err != nil {
		return nil, nil, fmt.Errorf("font not found: %w", err)
	}

	return &req, loadedFont, nil
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	req, font, err := s.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := font.LayoutText(converter.TextParams{
		Text:         req.Text,
		Size:         req.Size,
		Units:        req.Units,
		Kerning:      req.Kerning,
		LineHeight:   req.LineHeight,
		Datum:        req.Datum,
		CurveSamples: req.CurveSamples,
		LayerName:    req.LayerName,
		Weld:         req.Weld,
		DXFFormat:    req.DXFFormat,
		SlantAngle:   req.SlantAngle,
		Offset:       req.Offset,
		CornerJoin:   req.CornerJoin,
		ArcEnabled:   req.ArcEnabled,
		ArcRadius:    req.ArcRadius,
		ArcSweep:        req.ArcSweep,
		ArcAlign:        req.ArcAlign,
		ArcInward:       req.ArcInward,
		ConstructionBox: req.ConstructionBox,
	})
	if err != nil {
		http.Error(w, "Failed to render text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate CLI command equivalent
	escapedText := strings.ReplaceAll(req.Text, `"`, `\"`)
	escapedText = strings.ReplaceAll(escapedText, "\n", `\n`)
	weldFlag := ""
	if req.Weld {
		weldFlag = " --weld"
	}
	dxfFormatFlag := ""
	if req.DXFFormat == converter.DXFFormatPolyline {
		dxfFormatFlag = " --dxf-format polyline"
	}
	slantFlag := ""
	if req.SlantAngle != 0 {
		slantFlag = fmt.Sprintf(" --slant %.1f", req.SlantAngle)
	}
	offsetFlag := ""
	if req.Offset != 0 {
		offsetFlag = fmt.Sprintf(" --offset %.2f --corner-join %s", req.Offset, req.CornerJoin)
	}
	arcFlag := ""
	if req.ArcEnabled {
		inwardStr := ""
		if req.ArcInward {
			inwardStr = " --arc-inward"
		}
		arcFlag = fmt.Sprintf(" --arc --arc-radius %.1f --arc-sweep %.1f --arc-align %s%s", req.ArcRadius, req.ArcSweep, req.ArcAlign, inwardStr)
	}
	boxFlag := ""
	if req.ConstructionBox {
		boxFlag = " --construction-box"
	}

	cliCmd := fmt.Sprintf("text2svg --text \"%s\" --font \"%s\" --size %.1f --kerning %.2f --datum %s%s%s%s%s%s%s --format dxf --out export.dxf",
		escapedText, font.FontName, req.Size, req.Kerning, req.Datum, weldFlag, dxfFormatFlag, slantFlag, offsetFlag, arcFlag, boxFlag)

	resp := ConvertResponse{
		SVG:        result.SVG,
		Bounds:     result.Bounds,
		Width:      result.Bounds.Width(),
		Height:     result.Bounds.Height(),
		Units:      string(req.Units),
		GlyphCount: result.GlyphCount,
		PathCount:  result.PathCount,
		Welded:     result.Welded,
		DXFFormat:  result.DXFFormat,
		CLICommand: cliCmd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleExportDXF(w http.ResponseWriter, r *http.Request) {
	req, font, err := s.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := font.LayoutText(converter.TextParams{
		Text:            req.Text,
		Size:            req.Size,
		Units:           req.Units,
		Kerning:         req.Kerning,
		LineHeight:      req.LineHeight,
		Datum:           req.Datum,
		CurveSamples:    req.CurveSamples,
		LayerName:       req.LayerName,
		Weld:            req.Weld,
		DXFFormat:       req.DXFFormat,
		SlantAngle:      req.SlantAngle,
		Offset:          req.Offset,
		CornerJoin:      req.CornerJoin,
		ArcEnabled:      req.ArcEnabled,
		ArcRadius:       req.ArcRadius,
		ArcSweep:        req.ArcSweep,
		ArcAlign:        req.ArcAlign,
		ArcInward:       req.ArcInward,
		ConstructionBox: req.ConstructionBox,
	})
	if err != nil {
		http.Error(w, "Failed to generate DXF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := "text2svg-export.dxf"
	w.Header().Set("Content-Type", "application/dxf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write([]byte(result.DXF))
}

func (s *Server) handleExportSVG(w http.ResponseWriter, r *http.Request) {
	req, font, err := s.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := font.LayoutText(converter.TextParams{
		Text:            req.Text,
		Size:            req.Size,
		Units:           req.Units,
		Kerning:         req.Kerning,
		LineHeight:      req.LineHeight,
		Datum:           req.Datum,
		CurveSamples:    req.CurveSamples,
		LayerName:       req.LayerName,
		Weld:            req.Weld,
		DXFFormat:       req.DXFFormat,
		SlantAngle:      req.SlantAngle,
		Offset:          req.Offset,
		CornerJoin:      req.CornerJoin,
		ArcEnabled:      req.ArcEnabled,
		ArcRadius:       req.ArcRadius,
		ArcSweep:        req.ArcSweep,
		ArcAlign:        req.ArcAlign,
		ArcInward:       req.ArcInward,
		ConstructionBox: req.ConstructionBox,
	})
	if err != nil {
		http.Error(w, "Failed to generate SVG: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := "text2svg-export.svg"
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write([]byte(result.SVG))
}

func (s *Server) handleInspectDXF(w http.ResponseWriter, r *http.Request) {
	var reader io.Reader
	filename := "uploaded.dxf"

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing file in form field 'file'", http.StatusBadRequest)
			return
		}
		defer file.Close()
		filename = header.Filename
		reader = file
	} else {
		reader = r.Body
	}

	rep, err := dxf.InspectReader(reader)
	if err != nil {
		http.Error(w, "Failed to inspect DXF: "+err.Error(), http.StatusBadRequest)
		return
	}
	rep.Filename = filename

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rep)
}
