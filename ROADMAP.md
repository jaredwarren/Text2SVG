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

## Phase 1: 3D CAD Sketch Readiness & Robust Profile Geometry

Targeted at ensuring imported sketch profiles extrude flawlessly in Fusion 360, SolidWorks, and FreeCAD on the very first click.

### 1.1 True DXF Splines (AutoCAD 2000+ `SPLINE` Entities)
- **Problem**: Linearized polylines produce hundreds of flat planar facets when extruded, resulting in heavy geometry and preventing smooth fillets/chamfers along letter edges.
- **Solution**: Implement DXF 2000 (`AC1015`) export generating native **cubic B-splines (`SPLINE` entities)** with exact control points and knot vectors matching the font's Bézier curves.
- **Impact**: Produces true smooth curved surfaces in 3D CAD with 80% smaller file sizes and zero sketch solver lag.

### 1.2 Path Welding (Boolean Union for Manifold Extrusions)
- **Problem**: When cursive or tightly-kerned letters overlap, the intersecting lines create multiple crossing profile regions. CAD extruders will either fail with "Self-intersecting curve" or require manually selecting dozens of tiny nested regions.
- **Solution**: Implement polygon boolean union (Clipper algorithm in Go) to automatically merge overlapping glyphs into a single continuous outer boundary.
- **Features**:
  - Automatically merges outer overlapping strokes into a unified closed loop.
  - Automatically preserves inner counter loops (holes in `e`, `o`, `a`, `d`, `b`, etc.).
  - Toggle: `[x] Weld Overlapping Letters`.

### 1.3 Manifold Profile Cleanup & Hole Hierarchy
- **Problem**: CAD sketch solvers require explicit closed loops and fail if start and end vertices don't connect with exact tolerance, or if zero-length micro-segments exist.
- **Solution**:
  - Strict vertex deduplication and loop closure verification.
  - Orientation tagging: Outer boundaries oriented clockwise, inner holes (counters) oriented counter-clockwise, enabling CAD engines to immediately recognize which areas are solid vs hollow.

---

## Phase 2: Text Shaping for 3D Surfaces & Embossing

Targeted at dials, bezels, cylindrical bosses, and custom 3D printed lettering.

### 2.1 Text-on-a-Curve (Arch & Circular Text)
- **Problem**: 3D parts frequently require curved typography (e.g. watch bezels, rotary dials, round knob labels, circular coin/token engravings).
- **Solution**: Deform glyph coordinates along an arc defined by radius and sweep angle.
- **Controls**:
  - Toggle: `[x] Curve / Arc Text`
  - Sliders: `Radius` (mm/in), `Sweep Angle` (-360° to +360°), `Alignment` (Center, Left, Right).
  - Inward vs Outward orientation (letters pointing toward or away from the center of curvature).

### 2.2 Inset / Offset Profiles (Draft & Tolerance Compensation)
- **Problem**: 3D printed inlays, push-fit text, and mold draft angles require letters to be slightly expanded or contracted.
- **Solution**: Inward and outward curve offsetting.
- **Controls**:
  - `Contour Offset`: Offset distance in mm (e.g. -0.2mm for tight press-fit inlays, +0.5mm for raised border backplates).
  - Choice of corner joins: Round (for CNC ball/end mills) or Sharp/Miter (for sharp 3D prints).

### 2.3 Simulated Slant (Oblique) & Baseline Offsets
- **Problem**: Many fonts lack a true italic version, or technical CAD text requires a specific 15° or 75° drafting slant angle.
- **Solution**: Shear matrix transform slider (`Slant Angle`: -45° to +45°) to create italicized sketches from any standard font.

---

## Phase 3: Multi-CAD Export Formats & Alignment Helpers

Optimizing interoperability and placement within CAD workspaces.

### 3.1 CAD-Optimized SVG (for Fusion 360 & FreeCAD)
- **Features**:
  - Fusion 360 has a direct "Insert > Insert SVG" sketch feature.
  - Generate cubic Bézier `<path>` SVG with explicit physical millimeter scaling that imports directly to scale on any selected CAD construction plane.
  - Zero transform matrices (bake all translations and scales into raw coordinates so CAD import doesn't misplace origins).

### 3.2 Precision Origin & Datum Alignment for CAD Mating
- **Problem**: Placing extruded text on a CAD part requires snapping to reference axes (center of boss, center of bounding box, or typographical baseline).
- **Solution**: Enhanced Origin / Datum options:
  - **Center of Bounding Box** (for centering on round knobs / faces)
  - **Typographical Baseline-Center** (for aligning with mechanical construction lines)
  - **Bottom-Left (0,0)** (standard Cartesian datum)
  - Export reference bounding box wireframe (optional bounding rectangle on a `CONSTRUCTION` layer).

### 3.3 3D Extrusion Web Preview (Lightweight WebGL)
- **Problem**: Hard to visualize how the text will look when extruded into 3D.
- **Solution**: A simple 3D toggle in the web studio using Three.js/WebGL to view the text with an interactive 3D extrusion slider (`Extrude Height: 5mm`, `Bevel: 0.5mm`).

---

## Prioritized Implementation Roadmap

| Priority | Feature | Complexity | CAD Impact | Primary Benefit |
| :---: | :--- | :---: | :---: | :--- |
| **1** | **Path Welding (Boolean Union)** | Medium | **Critical** | Eliminates self-intersecting curve errors on cursive/tight text |
| **2** | **True DXF Splines (`SPLINE`)** | Medium | **High** | Replaces faceted polylines with true smooth CAD curves & fillets |
| **3** | **Text-on-a-Curve (Circular / Arc)** | Medium | **High** | Enables dials, round bezels, circular embossed parts |
| **4** | **Manifold Loop & Hole Orientation** | Low | **High** | Instant automatic profile recognition in CAD extrude |
| **5** | **Contour Offset (Inlays / Tolerances)**| Medium | **Medium** | Perfect tolerances for 3D printed two-color inlays |
| **6** | **3D WebGL Extrude Preview** | Low-Med | **Medium** | Visualizes 3D look before exporting to CAD |
