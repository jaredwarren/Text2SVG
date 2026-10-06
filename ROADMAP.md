# Text2SVG // Product & Engineering Roadmap: 3D CAD Typography

This roadmap re-focuses **Text2SVG** specifically on generating clean, high-precision vector typography for **3D CAD modeling, sketch extrusion, and parametric design** in tools like **Autodesk Fusion 360, SolidWorks, FreeCAD, Onshape, and Blender**.

---

## The 3D CAD Typography Challenge

Importing standard text into CAD software is notoriously painful:
1. **Faceted Polylines**: Most converters export thousands of tiny straight line segments. In CAD, this creates faceted surfaces, bloated sketch solvers, and makes adding edge fillets or chamfers fail.
2. **Self-Intersecting Profiles**: Overlapping letters (cursive, script, tight kerning) cause CAD extrude features to throw "Self-intersecting curve" errors.
3. **Broken Loop Closures**: Microscopic gaps or duplicate coincident vertices prevent CAD sketch engines from recognizing closed profiles for extrusion.
4. **Curved & Cylindrical Surfaces**: CAD text often needs to wrap around dials, bezels, cylinders, or circular bosses.

---

## Current Status: MVP v1.0 (Completed)

- [x] **Dual-Mode Architecture**: Standalone Go binary with embedded web studio + headless CLI mode.
- [x] **Font Ingestion**: Real-time TTF/OTF parsing via `sfnt`, drag-and-drop custom font upload, and auto-discovery of system fonts.
- [x] **Precision Typography Controls**: Font size (mm, in, px), kerning / letter-spacing, line height (leading), multi-line text layout.
- [x] **CNC/CAD Cartesian Coordinates**: $Y$ points UP, selectable Origin datum (**Bottom-Left (0,0)**, **Center**, **Top-Left**).
- [x] **DXF Exporter (R12)**: Standard polyline DXF export with unit scale headers.
- [x] **SVG Exporter**: Clean XML output with exact physical dimensions (`mm`, `in`) and non-clipping viewport.

---

## Phase 1: 3D CAD Sketch Readiness & Robust Profile Geometry (Completed)

Targeted at ensuring imported sketch profiles extrude flawlessly in Fusion 360, SolidWorks, and FreeCAD on the very first click.

### 1.1 True DXF Splines (AutoCAD 2000+ `SPLINE` Entities) - [x] Completed
- **Problem**: Linearized polylines produce hundreds of flat planar facets when extruded, resulting in heavy geometry and preventing smooth fillets/chamfers along letter edges.
- **Solution**: Implemented DXF 2000 (`AC1015`) export generating native **cubic B-splines (`SPLINE` entities)** with exact control points and knot vectors matching the font's Bézier curves, plus quadratic-to-cubic degree elevation and native `LINE` entities.
- **Impact**: Produces true smooth curved surfaces in 3D CAD with ~50% smaller file sizes and zero sketch solver lag.

### 1.2 Path Welding (Boolean Union for Manifold Extrusions) - [x] Completed
- **Problem**: When cursive or tightly-kerned letters overlap, the intersecting lines create multiple crossing profile regions. CAD extruders will either fail with "Self-intersecting curve" or require manually selecting dozens of tiny nested regions.
- **Solution**: Implemented polygon boolean union (Martínez algorithm in pure Go) to automatically merge overlapping glyphs into a single continuous outer boundary.
- **Features**:
  - Automatically merges outer overlapping strokes into a unified closed loop.
  - Automatically preserves inner counter loops (holes in `e`, `o`, `a`, `d`, `b`, etc.).
  - Toggle: `[x] Weld Overlapping Letters` (CLI `--weld` and Web Studio toggle).

### 1.3 Manifold Profile Cleanup & Hole Hierarchy - [x] Completed
- **Problem**: CAD sketch solvers require explicit closed loops and fail if start and end vertices don't connect with exact tolerance, or if zero-length micro-segments exist.
- **Solution**:
  - Strict vertex deduplication, collinear segment reduction, and loop closure verification.
  - Orientation tagging: Outer boundaries oriented clockwise (CW), inner holes (counters) oriented counter-clockwise (CCW), enabling CAD engines to immediately recognize which areas are solid vs hollow.
  - Containment nesting depth analysis (0=outer, 1=hole, 2=island).

---

## Phase 2: Text Shaping for 3D Surfaces & Embossing (Completed)

Targeted at dials, bezels, cylindrical bosses, press-fit inlays, and custom 3D printed lettering.

### 2.1 Text-on-a-Curve (Arch & Circular Text) - [x] Completed
- **Problem**: 3D parts frequently require curved typography (e.g. watch bezels, rotary dials, round knob labels, circular coin/token engravings).
- **Solution**: Implemented conformal arc mapping that dynamically bends contours along an arc defined by radius and sweep angle. Includes automatic contour segment subdivision (`SubdivideContour`) so straight glyph stems bend smoothly without facet distortion.
- **Features**:
  - Arc deformation along circumference with baseline preserving curvature.
  - Alignment options: `center`, `left`, `right`.
  - Inward vs Outward orientation (letters oriented toward or away from the center of curvature).
  - CLI flags: `--arc`, `--arc-radius`, `--arc-sweep`, `--arc-align`, `--arc-inward`.

### 2.2 Inset / Offset Profiles (Draft & Tolerance Compensation) - [x] Completed
- **Problem**: 3D printed inlays, push-fit text, and mold draft angles require letters to be slightly expanded or contracted.
- **Solution**: Implemented parallel edge offsetting with line-line intersection solving and convex corner joins (`round`, `miter`, `bevel`).
- **Features**:
  - Physical tolerance logic: positive offset expands solid boundaries and contracts inner holes, matching CAM kerf compensation and press-fit assembly.
  - Corner join styles: `round` (CNC ballnose / smooth curves), `miter` (sharp 90° corners), `bevel` (chamfered joins).
  - CLI flags: `--offset`, `--corner-join`.

### 2.3 Simulated Slant (Oblique) & Baseline Offsets - [x] Completed
- **Problem**: Many fonts lack a true italic version, or technical CAD text requires a specific 15° or 75° drafting slant angle.
- **Solution**: Implemented affine shear matrix transform along baseline $y=0$ ($x' = x + y\tan\theta$, $y'=y$) transforming contour vertices and Bézier control points.
- **Features**:
  - Full angle range: -45° to +45°.
  - CLI flag: `--slant`.

---

## Phase 3: Multi-CAD Export Formats & Alignment Helpers (Completed)

Optimizing interoperability and placement within CAD workspaces.

### 3.1 CAD-Optimized SVG (for Fusion 360 & FreeCAD) - [x] Completed
- **Features**:
  - Direct 1:1 physical scale millimeter/inch sketch import without manual scaling or unit conversion prompts.
  - Generates cubic Bézier `<path>` SVG (`M... C... Z`) with zero matrix transforms (all offsets and scales are baked directly into geometric coordinates).
  - Explicit `fill-rule="evenodd"` XML attribute and CSS stylesheet ensuring inner counters (e.g., 'O', 'A', 'B', '8', '0') are auto-recognized as cutouts in CAD extrusion profiles.

### 3.2 Precision Origin & Datum Alignment for CAD Mating - [x] Completed
- **Problem**: Placing extruded text on a CAD part requires snapping to reference axes (center of boss, center of bounding box, or typographical baseline).
- **Solution**: Enhanced Origin / Datum options:
  - **Bottom-Left (0,0)**: Standard Cartesian/CNC machining datum.
  - **Center (0,0)**: Symmetric origin for centering on circular dials, knobs, and faces.
  - **Top-Left (0,0)**: Graphic design and desktop publishing convention.
  - **Baseline-Left (0,0)**: Mathematical baseline datum ($y=0$) aligned with the left bounding box limit ($x=0$) for technical drawing alignment.
  - **Baseline-Center (0,0)**: Mathematical baseline datum ($y=0$) centered horizontally ($x=0$) along the text line.
  - **Construction Box Wireframe**: Optional bounding envelope wireframe exported on layer `CONSTRUCTION` (ACI 8 / gray in DXF) and `<rect class="construction-frame">` with dashed stroke in SVG (`--construction-box` CLI flag & Studio toggle).

### 3.3 3D Extrusion Web Preview (Lightweight WebGL) - [x] Completed
- **Problem**: Hard to visualize how the text will look when extruded into 3D.
- **Solution**: Interactive WebGL 3D solid viewer built directly into Web Studio using pure vendored Three.js, OrbitControls, and SVGLoader:
  - 3D Viewport toggle (`Wireframe` | `Fill` | `3D Solid`) with 360° orbit, pan, and zoom controls.
  - Real-time Extrusion Depth slider (0.5 mm to 100 mm).
  - Real-time Bevel / Chamfer slider (0.0 mm to 5.0 mm).
  - PBR Material finishes: Brushed Aluminum, Red Anodized Aluminum, Polished Brass / Gold, Matte Black Acrylic, Natural Delrin / White POM.
  - Engineering millimeter ground bed with coordinate lighting.

### 3.4 DXF Health Inspector & CAD Compatibility Linter - [x] Completed
- **Problem**: When a DXF fails to import into Onshape, Fusion 360, SolidWorks, or LightBurn, CAD software provides generic, unhelpful errors ("Translation failed", "Invalid database").
- **Solution**: A built-in pure Go DXF tokenizer, AST analyzer, and compatibility validator solving DXF import issues:
  - **CLI Command & Flag**: `text2svg --inspect <file.dxf>` and `text2svg inspect <file.dxf>`.
  - **REST API Endpoint**: `POST /api/inspect-dxf` accepting raw binary or multipart form file.
  - **Studio Diagnostic Modal**: 1-click inspection of current in-memory designs or drag-and-drop of any external third-party `.dxf` file.
  - **Schema Architecture Verification**: Audits AutoCAD R12 (`AC1009`) and AutoCAD 2000 (`AC1015`) section completeness (`HEADER`, `CLASSES`, `TABLES`, `BLOCKS`, `ENTITIES`, `OBJECTS`, and `$HANDSEED`).
  - **Loop Closure & Geometry Stats**: Detects watertight closed loops vs open curves, entity counts, dimensions, and extents.
  - **Target CAD Compatibility Matrix**: Diagnostic reporting for PTC Onshape, Autodesk Fusion 360, AutoCAD, FreeCAD, and Laser Cutters / LightBurn.

---

## Prioritized Implementation Roadmap

| Priority | Feature | Phase | Complexity | CAD Impact | Primary Benefit |
| :---: | :--- | :---: | :---: | :---: | :--- |
| **1** | **Path Welding (Boolean Union)** | Phase 1 | Medium | **Critical** | [x] Eliminates self-intersecting curve errors on cursive/tight text |
| **2** | **True DXF Splines (`SPLINE`)** | Phase 1 | Medium | **High** | [x] Replaces faceted polylines with true smooth CAD curves & fillets |
| **3** | **Manifold Loop & Hole Orientation** | Phase 1 | Low | **High** | [x] Instant automatic profile recognition in CAD extrude |
| **4** | **Text-on-a-Curve (Circular / Arc)** | Phase 2 | Medium | **High** | [x] Enables dials, round bezels, circular embossed parts |
| **5** | **Contour Offset (Inlays / Tolerances)**| Phase 2 | Medium | **High** | [x] Perfect tolerances for 3D printed two-color inlays & kerf |
| **6** | **Simulated Slant (Oblique)** | Phase 2 | Low | **Medium** | [x] Drafting slants and simulated italics for any font |
| **7** | **CAD-Optimized SVG (Fusion/FreeCAD)** | Phase 3 | Low | **High** | [x] Direct 1:1 scale millimeter sketch import |
| **8** | **DXF Health Inspector & Linter** | Phase 3 | Low-Med | **High** | [x] Instant diagnosis of DXF import failures in Onshape/Fusion |
| **9** | **3D WebGL Extrude Preview** | Phase 3 | Low-Med | **Medium** | [x] Visualizes 3D look before exporting to CAD |
