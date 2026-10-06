# Text2SVG // CAM & CAD Vector Studio

A high-performance Go application and interactive web studio that converts custom typography and fonts into CAM-ready **DXF** and **SVG** vector files.

Optimized for laser cutters, CNC routers, vinyl plotters, plasma cutters, and 3D CAD modeling (Fusion 360, SolidWorks, AutoCAD).

---

## Features

- **Dual-Mode Operation**:
  - **Interactive Web Studio**: Run `./text2svg` to launch a local browser studio with live preview, millimeter engineering grid, and real-time bounding box measurements.
  - **Headless CLI**: Run `text2svg --text "..."` for terminal scripting, automated batch processing, or CI/CD pipelines.
  - **"Copy CLI Command" Button**: Instantly copy the exact terminal command required to recreate any design.
- **Font Support**:
  - Load custom `.ttf`, `.otf`, and `.woff` font files via drag-and-drop or file picker.
  - Auto-scans installed system fonts and local `./fonts/` directory.
- **Precision Adjustments**:
  - Font Size / Em-Height (in `mm`, `in`, or `px`).
  - Kerning / Tracking adjustment.
  - Multi-line text with custom Line Height (Leading).
  - Selectable Datum Origin: **Bottom-Left (Standard CNC (0,0))**, **Center**, or **Top-Left**.
- **Vector & DXF Output**:
  - AutoCAD Release 12 / 2000 compliant DXF (`AC1009`).
  - Closed `POLYLINE` entities with clean vertices.
  - Standard CAM layer tagging: `CUT` (Red / ACI 1), `ENGRAVE` (Blue / ACI 5), `SCORE` (Yellow / ACI 2).
  - Standard SVG output with calibrated physical units (`mm`, `in`) and viewBox.

---

## Quick Start

### 1. Build the Binary
```bash
make build
# or: go build -o text2svg .
```

### 2. Launch the Web Studio
```bash
make run
# or: ./text2svg
```
This starts the local web server and automatically opens your default browser at `http://localhost:8080`.
Run `make dev` to start the server without automatically opening the browser.


### 3. Headless CLI Usage
Export directly from your terminal:
```bash
# Export DXF for laser cutting in millimeters
./text2svg --text "LASER CUT 2026" --font Arial --size 35 --kerning 1.5 --out cut.dxf

# Export SVG with center origin
./text2svg --text "CAD MODEL" --font Georgia --size 50 --datum center --format svg --out model.svg
```

#### CLI Flags
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--text` | `""` | Text to convert (triggers headless mode when set) |
| `--font` | `"Arial"` | Font path (`path/to/font.ttf`) or system font name |
| `--size` | `30.0` | Size / Em-height in target units |
| `--units` | `"mm"` | Physical measurement units: `mm`, `in`, `px` |
| `--kerning` | `0.0` | Extra letter spacing / tracking |
| `--leading` | `1.2` | Multi-line line spacing multiplier |
| `--datum` | `"bottom-left"` | Origin reference: `bottom-left`, `center`, `top-left` |
| `--format` | `"dxf"` | Output format: `dxf`, `svg` |
| `--dxf-format` | `"spline"` | DXF entity format: `spline` (AutoCAD 2000 AC1015 true cubic B-splines), `polyline` (R12/2000 LWPOLYLINE) |
| `--weld` | `false` | Weld overlapping letters into a continuous manifold loop (boolean union) |
| `--out` | `""` | Output filepath |
| `--port` | `8080` | Local port for Web Studio |
| `--no-browser`| `false` | Do not auto-launch browser on startup |

---

## Development

A comprehensive `Makefile` is included:

```bash
make help          # List all available targets
make build         # Compile text2svg binary
make run           # Build and launch Web Studio
make dev           # Start Web Studio without launching browser
make test          # Run test suite
make test-race     # Run tests with Go race detector
make coverage      # Generate HTML test coverage report
make fmt           # Format code with go fmt
make vet           # Run go vet static analysis
make lint          # Run linter (golangci-lint or go vet)
make tidy          # Tidy go.mod dependencies
make install       # Install binary to $GOBIN / $GOPATH/bin
make build-all     # Cross-compile for macOS, Linux, and Windows into dist/
make kill          # Free port and terminate any background instances
make clean         # Remove binaries, test coverage, and build artifacts
```

---

## License

This project is licensed under the [MIT License](file:///Users/jaredwarren/go/src/github.com/jaredwarren/Text2SVG/LICENSE).

