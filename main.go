package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
	"github.com/jaredwarren/Text2SVG/pkg/fonts"
	"github.com/jaredwarren/Text2SVG/pkg/server"
	"github.com/jaredwarren/Text2SVG/web"
)

func main() {
	// CLI Flags
	textFlag := flag.String("text", "", "Text to convert (if set, runs in headless CLI mode)")
	fontFlag := flag.String("font", "Arial", "Font file path or installed font name")
	sizeFlag := flag.Float64("size", 30.0, "Font size / em-height in units (default 30.0)")
	kerningFlag := flag.Float64("kerning", 0.0, "Extra letter spacing / kerning in units (default 0.0)")
	leadingFlag := flag.Float64("leading", 1.2, "Line height multiplier for multi-line text (default 1.2)")
	datumFlag := flag.String("datum", "bottom-left", "Origin datum point: bottom-left, center, top-left")
	unitsFlag := flag.String("units", "mm", "Physical units: mm, in, px")
	formatFlag := flag.String("format", "dxf", "Export format: dxf, svg")
	outFlag := flag.String("out", "", "Output destination filepath (e.g. output.dxf)")
	portFlag := flag.Int("port", 8080, "Port for web studio server")
	noBrowserFlag := flag.Bool("no-browser", false, "Do not automatically launch web browser")

	flag.Parse()

	// Initialize Font Manager
	fontMgr := fonts.NewManager()

	// 1. Headless CLI Mode
	if *textFlag != "" {
		runHeadless(fontMgr, *textFlag, *fontFlag, *sizeFlag, *kerningFlag, *leadingFlag, *datumFlag, *unitsFlag, *formatFlag, *outFlag)
		return
	}

	// 2. Interactive Web Server Mode
	runServer(fontMgr, *portFlag, !*noBrowserFlag)
}

func runHeadless(fm *fonts.Manager, text, fontPath string, size, kerning, leading float64, datumStr, unitsStr, formatStr, outPath string) {
	// Try loading font from path if it's a file, otherwise lookup in manager
	var lf *converter.LoadedFont
	if data, err := os.ReadFile(fontPath); err == nil {
		name := filepath.Base(fontPath)
		lf, err = converter.ParseFont(data, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing font file '%s': %v\n", fontPath, err)
			os.Exit(1)
		}
	} else {
		var err error
		lf, err = fm.GetFont(fontPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: font '%s' not found: %v\n", fontPath, err)
			os.Exit(1)
		}
	}

	params := converter.TextParams{
		Text:         text,
		Size:         size,
		Units:        converter.Units(unitsStr),
		Kerning:      kerning,
		LineHeight:   leading,
		Datum:        converter.Datum(datumStr),
		CurveSamples: 20,
		LayerName:    "CUT",
	}

	result, err := lf.LayoutText(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering text: %v\n", err)
		os.Exit(1)
	}

	format := strings.ToLower(formatStr)
	var outputContent string
	defaultExt := ".dxf"
	if format == "svg" {
		outputContent = result.SVG
		defaultExt = ".svg"
	} else {
		outputContent = result.DXF
		defaultExt = ".dxf"
	}

	if outPath == "" {
		cleanName := strings.ReplaceAll(text, " ", "_")
		if len(cleanName) > 16 {
			cleanName = cleanName[:16]
		}
		outPath = cleanName + defaultExt
	}

	if err := os.WriteFile(outPath, []byte(outputContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file '%s': %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully exported %s to: %s (%.2f x %.2f %s, %d glyphs, %d loops)\n",
		strings.ToUpper(format), outPath, result.Bounds.Width(), result.Bounds.Height(), unitsStr, result.GlyphCount, result.PathCount)
}

func runServer(fm *fonts.Manager, port int, openBrowser bool) {
	srv := server.NewServer(fm, web.AssetFS())

	// Find free port starting at port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Fallback to random free port
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to bind to network port: %v\n", err)
			os.Exit(1)
		}
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	fmt.Println("=================================================================")
	fmt.Println("   Text2SVG // CAM & CAD Vector Studio (v1.0)")
	fmt.Println("=================================================================")
	fmt.Printf("-> Studio running at: %s\n", url)
	fmt.Println("-> Press Ctrl+C to terminate.")
	fmt.Println("=================================================================")

	if openBrowser {
		go func() {
			switch runtime.GOOS {
			case "darwin":
				_ = exec.Command("open", url).Start()
			case "windows":
				_ = exec.Command("cmd", "/c", "start", url).Start()
			case "linux":
				_ = exec.Command("xdg-open", url).Start()
			}
		}()
	}

	if err := http.Serve(listener, srv.Routes()); err != nil {
		fmt.Fprintf(os.Stderr, "Server exited: %v\n", err)
	}
}
