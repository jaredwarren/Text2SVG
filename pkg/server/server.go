package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
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
	Text         string          `json:"text"`
	FontID       string          `json:"font_id"`
	Size         float64         `json:"size"`
	Units        converter.Units `json:"units"`
	Kerning      float64         `json:"kerning"`
	LineHeight   float64         `json:"line_height"`
	Datum        converter.Datum `json:"datum"`
	CurveSamples int             `json:"curve_samples"`
	LayerName    string          `json:"layer_name"`
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
	CLICommand string                `json:"cli_command"`
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/fonts", s.handleListFonts)
	mux.HandleFunc("POST /api/upload-font", s.handleUploadFont)
	mux.HandleFunc("POST /api/preview", s.handlePreview)
	mux.HandleFunc("POST /api/export/dxf", s.handleExportDXF)
	mux.HandleFunc("POST /api/export/svg", s.handleExportSVG)

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
	})
	if err != nil {
		http.Error(w, "Failed to render text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate CLI command equivalent
	escapedText := strings.ReplaceAll(req.Text, `"`, `\"`)
	escapedText = strings.ReplaceAll(escapedText, "\n", `\n`)
	cliCmd := fmt.Sprintf("text2svg --text \"%s\" --font \"%s\" --size %.1f --kerning %.2f --datum %s --format dxf --out export.dxf",
		escapedText, font.FontName, req.Size, req.Kerning, req.Datum)

	resp := ConvertResponse{
		SVG:        result.SVG,
		Bounds:     result.Bounds,
		Width:      result.Bounds.Width(),
		Height:     result.Bounds.Height(),
		Units:      string(req.Units),
		GlyphCount: result.GlyphCount,
		PathCount:  result.PathCount,
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
		Text:         req.Text,
		Size:         req.Size,
		Units:        req.Units,
		Kerning:      req.Kerning,
		LineHeight:   req.LineHeight,
		Datum:        req.Datum,
		CurveSamples: req.CurveSamples,
		LayerName:    req.LayerName,
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
		Text:         req.Text,
		Size:         req.Size,
		Units:        req.Units,
		Kerning:      req.Kerning,
		LineHeight:   req.LineHeight,
		Datum:        req.Datum,
		CurveSamples: req.CurveSamples,
		LayerName:    req.LayerName,
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
