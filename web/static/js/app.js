/**
 * Text2SVG // Studio Frontend Controller
 */

document.addEventListener('DOMContentLoaded', () => {
  // App State
  const state = {
    text: "LASER CUT 2026",
    font_id: "",
    size: 30.0,
    units: "mm",
    kerning: 0.0,
    line_height: 1.20,
    datum: "center", // "center" | "bottom-left" | "top-left" | "baseline-left" | "baseline-center"
    layer_name: "CUT",
    curve_samples: 20,
    weld: false,
    dxf_format: "spline", // "spline" | "polyline"
    slant_angle: 0.0,
    offset: 0.0,
    corner_join: "round",
    arc_enabled: false,
    arc_radius: 60.0,
    arc_sweep: 0.0,
    arc_align: "center",
    arc_inward: false,
    construction_box: false,
    boundary_mode: "none", // "none" | "conformal" | "box"
    boundary_offset: 4.0,
    boundary_padding_y: 0.0,
    boundary_radius: 3.0,
    boundary_fill_holes: true,
    boundary_shift_x: 0.0,
    boundary_shift_y: 0.0,
    extrude_depth: 5.0,
    bevel_size: 0.3,
    material_finish: "aluminum", // "aluminum" | "red-anodized" | "brass" | "acrylic-black" | "delrin-white"
    view_mode: "wireframe", // "wireframe" | "fill" | "3d"
    zoom: 1.0,
  };

  // DOM Elements
  const inputText = document.getElementById('input-text');
  const selectFont = document.getElementById('select-font');
  const fontDropzone = document.getElementById('font-dropzone');
  const fontFileInput = document.getElementById('font-file-input');
  const fontSourceTag = document.getElementById('font-source-tag');

  const inputSizeRange = document.getElementById('input-size-range');
  const inputSizeNum = document.getElementById('input-size-num');
  const inputKerningRange = document.getElementById('input-kerning-range');
  const inputKerningNum = document.getElementById('input-kerning-num');
  const inputLeadingRange = document.getElementById('input-leading-range');
  const inputLeadingNum = document.getElementById('input-leading-num');

  // Phase 2 Elements
  const inputSlantRange = document.getElementById('input-slant-range');
  const inputSlantNum = document.getElementById('input-slant-num');
  const inputOffsetRange = document.getElementById('input-offset-range');
  const inputOffsetNum = document.getElementById('input-offset-num');
  const selectCornerJoin = document.getElementById('select-corner-join');
  const toggleArc = document.getElementById('toggle-arc');
  const arcOptionsPanel = document.getElementById('arc-options-panel');
  const inputArcRadiusRange = document.getElementById('input-arc-radius-range');
  const inputArcRadiusNum = document.getElementById('input-arc-radius-num');
  const inputArcSweepRange = document.getElementById('input-arc-sweep-range');
  const inputArcSweepNum = document.getElementById('input-arc-sweep-num');
  const arcAlignButtons = document.querySelectorAll('#arc-align-pill-group .pill-btn');
  const toggleArcInward = document.getElementById('toggle-arc-inward');
  const shapingModeTag = document.getElementById('shaping-mode-tag');

  // Phase 4 Elements (Perimeter Boundary & Badge)
  const selectBoundaryMode = document.getElementById('select-boundary-mode');
  const boundaryModeTag = document.getElementById('boundary-mode-tag');
  const boundaryOptionsPanel = document.getElementById('boundary-options-panel');
  const labelBoundaryOffset = document.getElementById('label-boundary-offset');
  const hintBoundaryOffset = document.getElementById('hint-boundary-offset');
  const inputBoundaryOffsetRange = document.getElementById('input-boundary-offset-range');
  const inputBoundaryOffsetNum = document.getElementById('input-boundary-offset-num');
  const groupBoundaryPaddy = document.getElementById('group-boundary-paddy');
  const inputBoundaryPaddyRange = document.getElementById('input-boundary-paddy-range');
  const inputBoundaryPaddyNum = document.getElementById('input-boundary-paddy-num');
  const groupBoundaryRadius = document.getElementById('group-boundary-radius');
  const inputBoundaryRadiusRange = document.getElementById('input-boundary-radius-range');
  const inputBoundaryRadiusNum = document.getElementById('input-boundary-radius-num');
  const groupBoundaryFillHoles = document.getElementById('group-boundary-fill-holes');
  const toggleBoundaryFillHoles = document.getElementById('toggle-boundary-fill-holes');
  const boundaryDatumButtons = document.querySelectorAll('#boundary-datum-pill-group .pill-btn');
  const inputBoundaryShiftXRange = document.getElementById('input-boundary-shiftx-range');
  const inputBoundaryShiftXNum = document.getElementById('input-boundary-shiftx-num');
  const inputBoundaryShiftYRange = document.getElementById('input-boundary-shifty-range');
  const inputBoundaryShiftYNum = document.getElementById('input-boundary-shifty-num');

  // Phase 3 Elements
  const toggleConstructionBox = document.getElementById('toggle-construction-box');
  const btnMode3d = document.getElementById('btn-mode-3d');
  const webglStage = document.getElementById('webgl-stage');
  const inputExtrudeRange = document.getElementById('input-extrude-range');
  const inputExtrudeNum = document.getElementById('input-extrude-num');
  const inputBevelRange = document.getElementById('input-bevel-range');
  const inputBevelNum = document.getElementById('input-bevel-num');
  const selectMaterial = document.getElementById('select-material');
  const badge3dMaterial = document.getElementById('badge-3d-material');

  // Inspector Elements
  const btnInspectDxf = document.getElementById('btn-inspect-dxf');
  const inspectorModal = document.getElementById('inspector-modal');
  const btnCloseInspector = document.getElementById('btn-close-inspector');
  const btnInspectCurrent = document.getElementById('btn-inspect-current');
  const inspectorDropzone = document.getElementById('inspector-dropzone');
  const inspectorFileInput = document.getElementById('inspector-file-input');
  const inspectorReport = document.getElementById('inspector-report');

  const unitButtons = document.querySelectorAll('.unit-btn');
  const datumButtons = document.querySelectorAll('#datum-pill-group .pill-btn');
  const selectLayer = document.getElementById('select-layer');
  const selectSamples = document.getElementById('select-samples');
  const toggleWeld = document.getElementById('toggle-weld');
  const selectDxfFormat = document.getElementById('select-dxf-format');
  const cadModeTag = document.getElementById('cad-mode-tag');

  const btnModeWireframe = document.getElementById('btn-mode-wireframe');
  const btnModeFill = document.getElementById('btn-mode-fill');
  const btnZoomIn = document.getElementById('btn-zoom-in');
  const btnZoomOut = document.getElementById('btn-zoom-out');
  const btnZoomReset = document.getElementById('btn-zoom-reset');
  const canvasStage = document.getElementById('canvas-stage');
  const canvasWrapper = document.getElementById('canvas-wrapper');
  const svgContainer = document.getElementById('svg-container');
  const originIndicator = document.getElementById('origin-indicator');

  const readoutDims = document.getElementById('readout-dims');
  const readoutDatum = document.getElementById('readout-datum');
  const readoutLayer = document.getElementById('readout-layer');
  const readoutTopology = document.getElementById('readout-topology');
  const readoutStats = document.getElementById('readout-stats');
  const charCounter = document.getElementById('char-counter');
  const cliPreviewCode = document.getElementById('cli-preview-code');
  const toast = document.getElementById('toast');

  const btnExportDXF = document.getElementById('btn-export-dxf');
  const btnExportSVG = document.getElementById('btn-export-svg');
  const btnQuickDXF = document.getElementById('btn-quick-dxf');
  const btnQuickSVG = document.getElementById('btn-quick-svg');
  const btnCopyCLI = document.getElementById('btn-copy-cli');
  const btnCopyCLIMini = document.getElementById('btn-copy-cli-mini');
  const presetChips = document.querySelectorAll('.preset-chip');

  let debounceTimer = null;
  let lastCLICommand = "";
  let currentSVG = "";
  let currentBounds = null;

  function getCurrentPayload() {
    return {
      text: state.text,
      font_id: state.font_id,
      size: parseFloat(state.size),
      units: state.units,
      kerning: parseFloat(state.kerning),
      line_height: parseFloat(state.line_height),
      datum: state.datum,
      curve_samples: parseInt(state.curve_samples, 10),
      layer_name: state.layer_name,
      weld: Boolean(state.weld),
      dxf_format: state.dxf_format,
      slant_angle: parseFloat(state.slant_angle),
      offset: parseFloat(state.offset),
      corner_join: state.corner_join,
      arc_enabled: Boolean(state.arc_enabled),
      arc_radius: parseFloat(state.arc_radius),
      arc_sweep: parseFloat(state.arc_sweep),
      arc_align: state.arc_align,
      arc_inward: Boolean(state.arc_inward),
      construction_box: Boolean(state.construction_box),
      boundary_mode: state.boundary_mode,
      boundary_offset: parseFloat(state.boundary_offset),
      boundary_padding_y: parseFloat(state.boundary_padding_y),
      boundary_radius: parseFloat(state.boundary_radius),
      boundary_fill_holes: Boolean(state.boundary_fill_holes),
      boundary_shift_x: parseFloat(state.boundary_shift_x) || 0.0,
      boundary_shift_y: parseFloat(state.boundary_shift_y) || 0.0,
    };
  }

  // 1. Initial Font Fetch
  async function loadFonts() {
    try {
      const res = await fetch('/api/fonts');
      if (!res.ok) throw new Error('Failed to fetch fonts');
      const fonts = await res.json();

      selectFont.innerHTML = '';
      const grouped = {};
      fonts.forEach(f => {
        if (!grouped[f.category]) grouped[f.category] = [];
        grouped[f.category].push(f);
      });

      for (const [category, items] of Object.entries(grouped)) {
        const optgroup = document.createElement('optgroup');
        optgroup.label = `${category} Fonts`;
        items.forEach(f => {
          const opt = document.createElement('option');
          opt.value = f.id;
          opt.textContent = f.name;
          optgroup.appendChild(opt);
        });
        selectFont.appendChild(optgroup);
      }

      if (fonts.length > 0) {
        state.font_id = fonts[0].id;
        selectFont.value = state.font_id;
        fontSourceTag.textContent = fonts[0].category;
      }

      triggerUpdate();
    } catch (err) {
      console.error('Font load error:', err);
    }
  }

  // 2. Debounced API Preview Update
  function triggerUpdate() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(fetchPreview, 60);
  }

  async function fetchPreview() {
    state.text = inputText.value || " ";
    charCounter.textContent = `${state.text.length} chars`;

    try {
      const res = await fetch('/api/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(getCurrentPayload())
      });

      if (!res.ok) {
        const errText = await res.text();
        console.warn('Preview error:', errText);
        return;
      }

      const data = await res.json();
      currentSVG = data.svg;
      currentBounds = data.bounds;

      // Render SVG in 2D container
      svgContainer.innerHTML = currentSVG;
      applyViewMode();

      // Update Readouts
      readoutDims.textContent = `${data.width.toFixed(2)} × ${data.height.toFixed(2)} ${state.units}`;
      readoutDatum.textContent = formatDatumLabel(state.datum);
      if (state.boundary_mode !== 'none') {
        readoutLayer.textContent = `${state.layer_name} + BORDER (Cyan)`;
      } else {
        readoutLayer.textContent = `${state.layer_name} (${getLayerColorName(state.layer_name)})`;
      }
      if (readoutTopology) {
        let topoStr = data.welded ? "Manifold (Welded)" : "Standard";
        if (state.boundary_mode === 'conformal') topoStr += " · Conformal Badge";
        else if (state.boundary_mode === 'box') topoStr += " · Box Enclosure";
        readoutTopology.textContent = topoStr;
        readoutTopology.style.color = (data.welded || state.boundary_mode !== 'none') ? "#34d399" : "";
      }
      if (cadModeTag) {
        cadModeTag.textContent = state.dxf_format === 'spline' ? 'AC1015 Spline' : 'AC1009 Poly';
      }
      if (shapingModeTag) {
        const parts = [];
        if (state.boundary_mode !== 'none') {
          parts.push(state.boundary_mode === 'conformal' ? 'Conformal' : 'Box Badge');
        }
        if (state.arc_enabled) parts.push("Arc");
        if (state.slant_angle !== 0) parts.push(`${state.slant_angle > 0 ? '+' : ''}${state.slant_angle}°`);
        if (state.offset !== 0) parts.push(`${state.offset > 0 ? '+' : ''}${state.offset}${state.units}`);
        if (state.construction_box) parts.push("Guide Frame");
        shapingModeTag.textContent = parts.length > 0 ? parts.join(' · ') : 'Direct';
        shapingModeTag.style.color = parts.length > 0 ? '#60a5fa' : '';
      }
      readoutStats.textContent = `${data.glyph_count} Glyphs | ${data.path_count} Loops`;

      // Update CLI snippet
      lastCLICommand = data.cli_command;
      cliPreviewCode.textContent = lastCLICommand;

      // Position Datum Crosshair relative to SVG box
      updateDatumCrosshair();

      // If in 3D mode, update the WebGL extrusion solid
      if (state.view_mode === '3d') {
        render3DExtrusion();
      }
    } catch (err) {
      console.error('Preview fetch failed:', err);
    }
  }

  function formatDatumLabel(datum) {
    switch (datum) {
      case 'bottom-left': return 'Bottom-Left (0,0)';
      case 'center': return 'Center (0,0)';
      case 'top-left': return 'Top-Left (0,0)';
      case 'baseline-left': return 'Baseline-Left (0,0)';
      case 'baseline-center': return 'Baseline-Center (0,0)';
      default: return datum;
    }
  }

  function getLayerColorName(layer) {
    if (layer.includes('SCORE')) return 'ACI 2 / Yellow';
    if (layer.includes('ENGRAVE')) return 'ACI 5 / Blue';
    return 'ACI 1 / Red';
  }

  function updateDatumCrosshair() {
    const svgElem = svgContainer.querySelector('svg');
    if (!svgElem || !currentBounds) return;

    const rect = svgElem.getBoundingClientRect();
    const wrapperRect = canvasWrapper.getBoundingClientRect();

    const w = currentBounds.max_x - currentBounds.min_x;
    const h = currentBounds.max_y - currentBounds.min_y;

    let x = (rect.left - wrapperRect.left);
    let y = (rect.bottom - wrapperRect.top);

    if (w > 0 && h > 0) {
      const normX = -currentBounds.min_x / w;
      const normY = currentBounds.max_y / h;
      x = (rect.left - wrapperRect.left) + normX * rect.width;
      y = (rect.top - wrapperRect.top) + normY * rect.height;
    }

    originIndicator.style.left = `${x}px`;
    originIndicator.style.top = `${y}px`;
  }

  function applyViewMode() {
    if (state.view_mode === 'wireframe') {
      canvasWrapper.style.display = 'flex';
      webglStage.style.display = 'none';
      svgContainer.classList.add('view-mode-wireframe');
      svgContainer.classList.remove('view-mode-fill');
      btnModeWireframe.classList.add('active');
      btnModeFill.classList.remove('active');
      if (btnMode3d) btnMode3d.classList.remove('active');
    } else if (state.view_mode === 'fill') {
      canvasWrapper.style.display = 'flex';
      webglStage.style.display = 'none';
      svgContainer.classList.remove('view-mode-wireframe');
      svgContainer.classList.add('view-mode-fill');
      btnModeWireframe.classList.remove('active');
      btnModeFill.classList.add('active');
      if (btnMode3d) btnMode3d.classList.remove('active');
    } else if (state.view_mode === '3d') {
      canvasWrapper.style.display = 'none';
      webglStage.style.display = 'block';
      btnModeWireframe.classList.remove('active');
      btnModeFill.classList.remove('active');
      if (btnMode3d) btnMode3d.classList.add('active');
      render3DExtrusion();
    }
  }

  function applyZoom() {
    canvasWrapper.style.transform = `scale(${state.zoom})`;
    btnZoomReset.textContent = `${Math.round(state.zoom * 100)}%`;
    updateDatumCrosshair();
  }

  // 3. Event Listeners for Controls

  // Text Change
  inputText.addEventListener('input', triggerUpdate);

  // Font Selection
  selectFont.addEventListener('change', () => {
    state.font_id = selectFont.value;
    triggerUpdate();
  });

  // Size Sync
  inputSizeRange.addEventListener('input', () => {
    state.size = parseFloat(inputSizeRange.value);
    inputSizeNum.value = state.size.toFixed(1);
    triggerUpdate();
  });
  inputSizeNum.addEventListener('input', () => {
    state.size = parseFloat(inputSizeNum.value) || 20.0;
    inputSizeRange.value = state.size;
    triggerUpdate();
  });

  // Kerning Sync
  inputKerningRange.addEventListener('input', () => {
    state.kerning = parseFloat(inputKerningRange.value);
    inputKerningNum.value = state.kerning.toFixed(1);
    triggerUpdate();
  });
  inputKerningNum.addEventListener('input', () => {
    state.kerning = parseFloat(inputKerningNum.value) || 0.0;
    inputKerningRange.value = state.kerning;
    triggerUpdate();
  });

  // Leading Sync
  inputLeadingRange.addEventListener('input', () => {
    state.line_height = parseFloat(inputLeadingRange.value);
    inputLeadingNum.value = state.line_height.toFixed(2);
    triggerUpdate();
  });
  inputLeadingNum.addEventListener('input', () => {
    state.line_height = parseFloat(inputLeadingNum.value) || 1.2;
    inputLeadingRange.value = state.line_height;
    triggerUpdate();
  });

  // Phase 2: Slant Angle Sync
  if (inputSlantRange && inputSlantNum) {
    inputSlantRange.addEventListener('input', () => {
      state.slant_angle = parseFloat(inputSlantRange.value);
      inputSlantNum.value = Math.round(state.slant_angle);
      triggerUpdate();
    });
    inputSlantNum.addEventListener('input', () => {
      state.slant_angle = parseFloat(inputSlantNum.value) || 0.0;
      inputSlantRange.value = state.slant_angle;
      triggerUpdate();
    });
  }

  // Phase 2: Tolerance Offset Sync
  if (inputOffsetRange && inputOffsetNum) {
    inputOffsetRange.addEventListener('input', () => {
      state.offset = parseFloat(inputOffsetRange.value);
      inputOffsetNum.value = state.offset.toFixed(2);
      triggerUpdate();
    });
    inputOffsetNum.addEventListener('input', () => {
      state.offset = parseFloat(inputOffsetNum.value) || 0.0;
      inputOffsetRange.value = state.offset;
      triggerUpdate();
    });
  }

  // Phase 2: Corner Join Style
  if (selectCornerJoin) {
    selectCornerJoin.addEventListener('change', () => {
      state.corner_join = selectCornerJoin.value;
      triggerUpdate();
    });
  }

  // Phase 2: Text-on-a-Curve (Arc) Toggle
  if (toggleArc) {
    toggleArc.addEventListener('change', () => {
      state.arc_enabled = toggleArc.checked;
      if (arcOptionsPanel) {
        arcOptionsPanel.style.display = state.arc_enabled ? 'flex' : 'none';
      }
      triggerUpdate();
    });
  }

  // Phase 2: Arc Radius Sync
  if (inputArcRadiusRange && inputArcRadiusNum) {
    inputArcRadiusRange.addEventListener('input', () => {
      state.arc_radius = parseFloat(inputArcRadiusRange.value);
      inputArcRadiusNum.value = Math.round(state.arc_radius);
      triggerUpdate();
    });
    inputArcRadiusNum.addEventListener('input', () => {
      state.arc_radius = parseFloat(inputArcRadiusNum.value) || 60.0;
      inputArcRadiusRange.value = state.arc_radius;
      triggerUpdate();
    });
  }

  // Phase 2: Arc Sweep Sync
  if (inputArcSweepRange && inputArcSweepNum) {
    inputArcSweepRange.addEventListener('input', () => {
      state.arc_sweep = parseFloat(inputArcSweepRange.value);
      inputArcSweepNum.value = Math.round(state.arc_sweep);
      triggerUpdate();
    });
    inputArcSweepNum.addEventListener('input', () => {
      state.arc_sweep = parseFloat(inputArcSweepNum.value) || 0.0;
      inputArcSweepRange.value = state.arc_sweep;
      triggerUpdate();
    });
  }

  // Phase 2: Arc Align Pill Buttons
  arcAlignButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      arcAlignButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      state.arc_align = btn.dataset.align;
      triggerUpdate();
    });
  });

  // Phase 2: Arc Inward Toggle
  if (toggleArcInward) {
    toggleArcInward.addEventListener('change', () => {
      state.arc_inward = toggleArcInward.checked;
      triggerUpdate();
    });
  }

  // Phase 4: Perimeter Boundary & Badge Controls
  function updateBoundaryUI() {
    if (!selectBoundaryMode) return;
    state.boundary_mode = selectBoundaryMode.value;

    if (state.boundary_mode === 'none') {
      if (boundaryOptionsPanel) boundaryOptionsPanel.style.display = 'none';
      if (boundaryModeTag) {
        boundaryModeTag.textContent = 'None';
        boundaryModeTag.className = 'badge-mini';
        boundaryModeTag.style.color = '';
      }
    } else if (state.boundary_mode === 'conformal') {
      if (boundaryOptionsPanel) boundaryOptionsPanel.style.display = 'block';
      if (boundaryModeTag) {
        boundaryModeTag.textContent = 'Conformal';
        boundaryModeTag.className = 'badge-mini badge-active';
        boundaryModeTag.style.color = '#06b6d4';
      }
      if (labelBoundaryOffset) labelBoundaryOffset.textContent = 'Contour Offset Distance';
      if (hintBoundaryOffset) hintBoundaryOffset.textContent = 'Outward bubble contour distance around letter profiles';
      if (groupBoundaryPaddy) groupBoundaryPaddy.style.display = 'none';
      if (groupBoundaryRadius) groupBoundaryRadius.style.display = 'none';
      if (groupBoundaryFillHoles) groupBoundaryFillHoles.style.display = 'flex';
    } else if (state.boundary_mode === 'box') {
      if (boundaryOptionsPanel) boundaryOptionsPanel.style.display = 'block';
      if (boundaryModeTag) {
        boundaryModeTag.textContent = 'Box Enclosure';
        boundaryModeTag.className = 'badge-mini badge-active';
        boundaryModeTag.style.color = '#06b6d4';
      }
      if (labelBoundaryOffset) labelBoundaryOffset.textContent = 'Padding X (or Uniform)';
      if (hintBoundaryOffset) hintBoundaryOffset.textContent = 'Horizontal margin / uniform boundary padding around envelope';
      if (groupBoundaryPaddy) groupBoundaryPaddy.style.display = 'block';
      if (groupBoundaryRadius) groupBoundaryRadius.style.display = 'block';
      if (groupBoundaryFillHoles) groupBoundaryFillHoles.style.display = 'none';
    }
  }

  if (selectBoundaryMode) {
    selectBoundaryMode.addEventListener('change', () => {
      updateBoundaryUI();
      triggerUpdate();
    });
  }

  if (inputBoundaryOffsetRange && inputBoundaryOffsetNum) {
    inputBoundaryOffsetRange.addEventListener('input', () => {
      state.boundary_offset = parseFloat(inputBoundaryOffsetRange.value);
      inputBoundaryOffsetNum.value = state.boundary_offset.toFixed(1);
      triggerUpdate();
    });
    inputBoundaryOffsetNum.addEventListener('input', () => {
      state.boundary_offset = parseFloat(inputBoundaryOffsetNum.value) || 4.0;
      inputBoundaryOffsetRange.value = state.boundary_offset;
      triggerUpdate();
    });
  }

  if (inputBoundaryPaddyRange && inputBoundaryPaddyNum) {
    inputBoundaryPaddyRange.addEventListener('input', () => {
      state.boundary_padding_y = parseFloat(inputBoundaryPaddyRange.value);
      inputBoundaryPaddyNum.value = state.boundary_padding_y.toFixed(1);
      triggerUpdate();
    });
    inputBoundaryPaddyNum.addEventListener('input', () => {
      state.boundary_padding_y = parseFloat(inputBoundaryPaddyNum.value) || 0.0;
      inputBoundaryPaddyRange.value = state.boundary_padding_y;
      triggerUpdate();
    });
  }

  if (inputBoundaryRadiusRange && inputBoundaryRadiusNum) {
    inputBoundaryRadiusRange.addEventListener('input', () => {
      state.boundary_radius = parseFloat(inputBoundaryRadiusRange.value);
      inputBoundaryRadiusNum.value = state.boundary_radius.toFixed(1);
      triggerUpdate();
    });
    inputBoundaryRadiusNum.addEventListener('input', () => {
      state.boundary_radius = parseFloat(inputBoundaryRadiusNum.value) || 0.0;
      inputBoundaryRadiusRange.value = state.boundary_radius;
      triggerUpdate();
    });
  }

  if (toggleBoundaryFillHoles) {
    toggleBoundaryFillHoles.addEventListener('change', () => {
      state.boundary_fill_holes = toggleBoundaryFillHoles.checked;
      triggerUpdate();
    });
  }

  // Boundary Shift X Sync
  if (inputBoundaryShiftXRange && inputBoundaryShiftXNum) {
    inputBoundaryShiftXRange.addEventListener('input', () => {
      state.boundary_shift_x = parseFloat(inputBoundaryShiftXRange.value) || 0.0;
      inputBoundaryShiftXNum.value = state.boundary_shift_x.toFixed(1);
      triggerUpdate();
    });
    inputBoundaryShiftXNum.addEventListener('input', () => {
      state.boundary_shift_x = parseFloat(inputBoundaryShiftXNum.value) || 0.0;
      inputBoundaryShiftXRange.value = state.boundary_shift_x;
      triggerUpdate();
    });
  }

  // Boundary Shift Y Sync
  if (inputBoundaryShiftYRange && inputBoundaryShiftYNum) {
    inputBoundaryShiftYRange.addEventListener('input', () => {
      state.boundary_shift_y = parseFloat(inputBoundaryShiftYRange.value) || 0.0;
      inputBoundaryShiftYNum.value = state.boundary_shift_y.toFixed(1);
      triggerUpdate();
    });
    inputBoundaryShiftYNum.addEventListener('input', () => {
      state.boundary_shift_y = parseFloat(inputBoundaryShiftYNum.value) || 0.0;
      inputBoundaryShiftYRange.value = state.boundary_shift_y;
      triggerUpdate();
    });
  }

  // Units Buttons
  unitButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      unitButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      state.units = btn.dataset.unit;
      document.querySelectorAll('.unit-label').forEach(el => {
        if (el.textContent !== 'x' && el.textContent !== '°') el.textContent = state.units;
      });
      triggerUpdate();
    });
  });

  // Datum Sync Function
  function setDatum(datum) {
    state.datum = datum;
    datumButtons.forEach(b => b.classList.toggle('active', b.dataset.datum === datum));
    if (boundaryDatumButtons) {
      boundaryDatumButtons.forEach(b => b.classList.toggle('active', b.dataset.datum === datum));
    }
    triggerUpdate();
  }

  // Datum Buttons (CAD & CAM panel)
  datumButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      setDatum(btn.dataset.datum);
    });
  });

  // Boundary Datum Buttons (Perimeter Boundary panel)
  if (boundaryDatumButtons) {
    boundaryDatumButtons.forEach(btn => {
      btn.addEventListener('click', () => {
        setDatum(btn.dataset.datum);
      });
    });
  }

  // Layer Select
  selectLayer.addEventListener('change', () => {
    state.layer_name = selectLayer.value;
    triggerUpdate();
  });

  // Curve Samples Select
  selectSamples.addEventListener('change', () => {
    state.curve_samples = parseInt(selectSamples.value, 10);
    triggerUpdate();
  });

  // Path Welding Toggle
  if (toggleWeld) {
    toggleWeld.addEventListener('change', () => {
      state.weld = toggleWeld.checked;
      triggerUpdate();
    });
  }

  // DXF Entity Format Select
  if (selectDxfFormat) {
    selectDxfFormat.addEventListener('change', () => {
      state.dxf_format = selectDxfFormat.value;
      if (cadModeTag) {
        cadModeTag.textContent = state.dxf_format === 'spline' ? 'AC1015 Spline' : 'AC1009 Poly';
      }
      triggerUpdate();
    });
  }

  // View Mode Buttons
  btnModeWireframe.addEventListener('click', () => {
    state.view_mode = 'wireframe';
    applyViewMode();
  });
  btnModeFill.addEventListener('click', () => {
    state.view_mode = 'fill';
    applyViewMode();
  });
  if (btnMode3d) {
    btnMode3d.addEventListener('click', () => {
      state.view_mode = '3d';
      applyViewMode();
    });
  }

  // Phase 3: Reference Bounding Box (Construction Box) Toggle
  if (toggleConstructionBox) {
    toggleConstructionBox.addEventListener('change', () => {
      state.construction_box = toggleConstructionBox.checked;
      triggerUpdate();
    });
  }

  // Phase 3: 3D Extrusion Sliders & Materials
  if (inputExtrudeRange && inputExtrudeNum) {
    inputExtrudeRange.addEventListener('input', () => {
      state.extrude_depth = parseFloat(inputExtrudeRange.value);
      inputExtrudeNum.value = state.extrude_depth.toFixed(1);
      if (state.view_mode === '3d') render3DExtrusion();
    });
    inputExtrudeNum.addEventListener('input', () => {
      state.extrude_depth = parseFloat(inputExtrudeNum.value) || 5.0;
      inputExtrudeRange.value = state.extrude_depth;
      if (state.view_mode === '3d') render3DExtrusion();
    });
  }

  if (inputBevelRange && inputBevelNum) {
    inputBevelRange.addEventListener('input', () => {
      state.bevel_size = parseFloat(inputBevelRange.value);
      inputBevelNum.value = state.bevel_size.toFixed(1);
      if (state.view_mode === '3d') render3DExtrusion();
    });
    inputBevelNum.addEventListener('input', () => {
      state.bevel_size = parseFloat(inputBevelNum.value) || 0.0;
      inputBevelRange.value = state.bevel_size;
      if (state.view_mode === '3d') render3DExtrusion();
    });
  }

  if (selectMaterial) {
    selectMaterial.addEventListener('change', () => {
      state.material_finish = selectMaterial.value;
      if (badge3dMaterial) {
        badge3dMaterial.textContent = selectMaterial.options[selectMaterial.selectedIndex].text.split(' ')[0];
      }
      if (state.view_mode === '3d') render3DExtrusion();
    });
  }

  // Zoom Buttons
  btnZoomIn.addEventListener('click', () => {
    state.zoom = Math.min(state.zoom + 0.15, 3.0);
    applyZoom();
  });
  btnZoomOut.addEventListener('click', () => {
    state.zoom = Math.max(state.zoom - 0.15, 0.3);
    applyZoom();
  });
  btnZoomReset.addEventListener('click', () => {
    state.zoom = 1.0;
    applyZoom();
  });

  // Mouse wheel zoom on canvas
  canvasStage.addEventListener('wheel', (e) => {
    e.preventDefault();
    if (e.deltaY < 0) {
      state.zoom = Math.min(state.zoom + 0.08, 3.0);
    } else {
      state.zoom = Math.max(state.zoom - 0.08, 0.3);
    }
    applyZoom();
  }, { passive: false });

  // 4. Drag & Drop Font Upload
  fontDropzone.addEventListener('click', () => fontFileInput.click());
  fontDropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    fontDropzone.classList.add('drag-over');
  });
  fontDropzone.addEventListener('dragleave', () => fontDropzone.classList.remove('drag-over'));
  fontDropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    fontDropzone.classList.remove('drag-over');
    if (e.dataTransfer.files.length > 0) {
      uploadFontFile(e.dataTransfer.files[0]);
    }
  });
  fontFileInput.addEventListener('change', () => {
    if (fontFileInput.files.length > 0) {
      uploadFontFile(fontFileInput.files[0]);
    }
  });

  async function uploadFontFile(file) {
    const formData = new FormData();
    formData.append('font', file);

    showToast(`Uploading ${file.name}...`);
    try {
      const res = await fetch('/api/upload-font', {
        method: 'POST',
        body: formData,
      });

      if (!res.ok) {
        const err = await res.text();
        showToast(`Upload failed: ${err}`);
        return;
      }

      const result = await res.json();
      showToast(`Font "${result.name}" loaded!`);
      // Reload fonts list and auto-select
      await loadFonts();
      selectFont.value = result.id;
      state.font_id = result.id;
      fontSourceTag.textContent = 'Uploaded';
      triggerUpdate();
    } catch (err) {
      showToast(`Upload error: ${err.message}`);
    }
  }

  // 5. Presets
  function applyPreset(preset) {
    presetChips.forEach(c => c.classList.toggle('active', c.dataset.preset === preset));

    if (preset === 'laser') {
      state.datum = 'bottom-left';
      state.layer_name = 'CUT';
      state.view_mode = 'wireframe';
      state.units = 'mm';
      state.weld = true;
      state.dxf_format = 'polyline';
      selectLayer.value = 'CUT';
      if (toggleWeld) toggleWeld.checked = true;
      if (selectDxfFormat) selectDxfFormat.value = 'polyline';
      if (cadModeTag) cadModeTag.textContent = 'AC1009 Poly';
    } else if (preset === 'cad') {
      state.datum = 'center';
      state.layer_name = 'CUT';
      state.view_mode = 'wireframe';
      state.units = 'mm';
      state.curve_samples = 36;
      state.weld = true;
      state.dxf_format = 'spline';
      selectSamples.value = '36';
      if (toggleWeld) toggleWeld.checked = true;
      if (selectDxfFormat) selectDxfFormat.value = 'spline';
      if (cadModeTag) cadModeTag.textContent = 'AC1015 Spline';
    } else if (preset === 'vector') {
      state.view_mode = 'fill';
      state.datum = 'top-left';
      state.units = 'px';
      state.weld = false;
      state.dxf_format = 'spline';
      if (toggleWeld) toggleWeld.checked = false;
    }

    // Sync UI buttons
    datumButtons.forEach(b => b.classList.toggle('active', b.dataset.datum === state.datum));
    if (boundaryDatumButtons) boundaryDatumButtons.forEach(b => b.classList.toggle('active', b.dataset.datum === state.datum));
    unitButtons.forEach(b => b.classList.toggle('active', b.dataset.unit === state.units));
    applyViewMode();
    triggerUpdate();
  }

  presetChips.forEach(chip => {
    chip.addEventListener('click', () => {
      applyPreset(chip.dataset.preset);
    });
  });

  // 6. Exports & Downloads
  async function downloadExport(type) {
    const url = `/api/export/${type}`;
    showToast(`Generating ${type.toUpperCase()}...`);

    try {
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(getCurrentPayload())
      });

      if (!res.ok) throw new Error('Export failed');

      const blob = await res.blob();
      const cleanName = (state.text.trim().substring(0, 16) || 'text').replace(/[^a-zA-Z0-9_-]/g, '_');
      const filename = `${cleanName}.${type}`;

      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(link.href);

      showToast(`Downloaded ${filename}`);
    } catch (err) {
      showToast(`Export error: ${err.message}`);
    }
  }

  btnExportDXF.addEventListener('click', () => downloadExport('dxf'));
  btnQuickDXF.addEventListener('click', () => downloadExport('dxf'));
  btnExportSVG.addEventListener('click', () => downloadExport('svg'));
  btnQuickSVG.addEventListener('click', () => downloadExport('svg'));

  // 7. Copy CLI Snippet
  function copyCLI() {
    if (!lastCLICommand) return;
    navigator.clipboard.writeText(lastCLICommand).then(() => {
      showToast('CLI command copied!');
    }).catch(() => {
      showToast('Failed to copy to clipboard');
    });
  }

  btnCopyCLI.addEventListener('click', copyCLI);
  btnCopyCLIMini.addEventListener('click', copyCLI);

  function showToast(msg) {
    toast.textContent = msg;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 2200);
  }

  // ==========================================================================
  // Phase 3: Three.js 3D Extrusion Engine (WebGL)
  // ==========================================================================

  let scene, camera, renderer, controls, textGroup, gridHelper;
  let is3DInitialized = false;

  function initThreeScene() {
    if (is3DInitialized || typeof THREE === 'undefined') return;

    const width = canvasStage.clientWidth || 800;
    const height = canvasStage.clientHeight || 500;

    scene = new THREE.Scene();

    camera = new THREE.PerspectiveCamera(40, width / height, 0.1, 3000);
    camera.position.set(0, 140, 220);

    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;

    webglStage.innerHTML = '';
    webglStage.appendChild(renderer.domElement);

    if (THREE.OrbitControls) {
      controls = new THREE.OrbitControls(camera, renderer.domElement);
      controls.enableDamping = true;
      controls.dampingFactor = 0.08;
      controls.maxPolarAngle = Math.PI / 2 + 0.05; // keep above table
      controls.target.set(0, 0, 0);
    }

    // Lights
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.7);
    scene.add(ambientLight);

    const keyLight = new THREE.DirectionalLight(0xffffff, 0.95);
    keyLight.position.set(120, 200, 150);
    keyLight.castShadow = true;
    scene.add(keyLight);

    const fillLight = new THREE.DirectionalLight(0x93c5fd, 0.45);
    fillLight.position.set(-150, 100, -100);
    scene.add(fillLight);

    const rimLight = new THREE.DirectionalLight(0xfca5a5, 0.35);
    rimLight.position.set(0, -100, 120);
    scene.add(rimLight);

    // CAM Bed Grid (Millimeter scale)
    gridHelper = new THREE.GridHelper(300, 30, 0xef4444, 0x1e293b);
    gridHelper.position.y = 0;
    scene.add(gridHelper);

    textGroup = new THREE.Group();
    scene.add(textGroup);

    window.addEventListener('resize', onWindowResize);

    function onWindowResize() {
      if (!webglStage || webglStage.style.display === 'none') return;
      const w = canvasStage.clientWidth;
      const h = canvasStage.clientHeight;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    }

    function animate() {
      requestAnimationFrame(animate);
      if (controls) controls.update();
      renderer.render(scene, camera);
    }
    animate();

    is3DInitialized = true;
  }

  function getThreeMaterial(finish) {
    switch (finish) {
      case 'red-anodized':
        return new THREE.MeshStandardMaterial({
          color: 0xdc2626,
          roughness: 0.22,
          metalness: 0.72,
        });
      case 'brass':
        return new THREE.MeshStandardMaterial({
          color: 0xd97706,
          roughness: 0.25,
          metalness: 0.92,
        });
      case 'acrylic-black':
        return new THREE.MeshStandardMaterial({
          color: 0x18181b,
          roughness: 0.12,
          metalness: 0.05,
        });
      case 'delrin-white':
        return new THREE.MeshStandardMaterial({
          color: 0xf8fafc,
          roughness: 0.45,
          metalness: 0.02,
        });
      case 'aluminum':
      default:
        return new THREE.MeshStandardMaterial({
          color: 0xc8d0d8,
          roughness: 0.28,
          metalness: 0.88,
        });
    }
  }

  function render3DExtrusion() {
    if (!is3DInitialized) {
      initThreeScene();
    }
    if (!currentSVG || typeof THREE === 'undefined' || !THREE.SVGLoader) return;

    // Clear previous geometries
    while (textGroup.children.length > 0) {
      const child = textGroup.children[0];
      textGroup.remove(child);
      if (child.geometry) child.geometry.dispose();
      if (child.material) {
        if (Array.isArray(child.material)) child.material.forEach(m => m.dispose());
        else child.material.dispose();
      }
    }

    const loader = new THREE.SVGLoader();
    let svgData;
    try {
      svgData = loader.parse(currentSVG);
    } catch (e) {
      console.error("SVGLoader parsing error:", e);
      return;
    }

    const material = getThreeMaterial(state.material_finish);
    const depth = Math.max(0.5, parseFloat(state.extrude_depth));
    const bevel = parseFloat(state.bevel_size);
    const extrudeSettings = {
      depth: depth,
      bevelEnabled: bevel > 0.01,
      bevelThickness: bevel,
      bevelSize: bevel,
      bevelOffset: 0,
      bevelSegments: 3,
      curveSegments: 16,
    };

    const hasBoundary = svgData.paths.some(p => {
      const cls = p.userData?.node?.getAttribute('class') || '';
      return cls.includes('boundary-path') || p.userData?.node?.classList?.contains('boundary-path');
    });

    const baseDepth = Math.max(1.5, Math.min(4.0, depth * 0.6));
    const baseplateSettings = {
      depth: baseDepth,
      bevelEnabled: bevel > 0.01,
      bevelThickness: Math.min(bevel, 0.4),
      bevelSize: Math.min(bevel, 0.4),
      bevelOffset: 0,
      bevelSegments: 2,
      curveSegments: 16,
    };

    // Contrasting material for physical backing plate
    const baseplateMaterial = new THREE.MeshStandardMaterial({
      color: state.material_finish === 'acrylic-black' ? 0x94a3b8 : 0x1e293b,
      roughness: 0.35,
      metalness: state.material_finish === 'acrylic-black' ? 0.75 : 0.4,
    });

    for (const path of svgData.paths) {
      const isConstruction = path.userData?.node?.classList?.contains('construction-frame');

      if (isConstruction) {
        // Wireframe line envelope on CAM table
        const shapes = THREE.SVGLoader.createShapes(path);
        shapes.forEach(shape => {
          const points = shape.getPoints();
          const lineGeo = new THREE.BufferGeometry().setFromPoints(points);
          const lineMat = new THREE.LineDashedMaterial({
            color: 0x94a3b8,
            dashSize: 3,
            gapSize: 2,
            linewidth: 1,
          });
          const line = new THREE.Line(lineGeo, lineMat);
          line.computeLineDistances();
          line.rotation.x = -Math.PI / 2;
          textGroup.add(line);
        });
        continue;
      }

      const isBoundary = (path.userData?.node?.getAttribute('class') || '').includes('boundary-path') ||
                         path.userData?.node?.classList?.contains('boundary-path');

      const shapes = THREE.SVGLoader.createShapes(path);
      for (const shape of shapes) {
        try {
          const geom = new THREE.ExtrudeGeometry(shape, isBoundary ? baseplateSettings : extrudeSettings);
          // Invert Y so text baseline is upright in Three.js coordinates
          geom.scale(1, -1, 1);
          const mat = isBoundary ? baseplateMaterial : material;
          const mesh = new THREE.Mesh(geom, mat);
          mesh.castShadow = true;
          mesh.receiveShadow = true;
          if (isBoundary) {
            mesh.position.z = 0;
          } else if (hasBoundary) {
            mesh.position.z = baseDepth;
          }
          textGroup.add(mesh);
        } catch (err) {
          console.warn("Failed to extrude shape:", err);
        }
      }
    }

    // Lie flat on CAM bed with Z-depth extruding upwards
    textGroup.rotation.x = -Math.PI / 2;

    const bbox = new THREE.Box3().setFromObject(textGroup);
    const center = bbox.getCenter(new THREE.Vector3());
    const size = bbox.getSize(new THREE.Vector3());

    if (controls) {
      controls.target.set(center.x, 0, center.z);
      const maxDim = Math.max(size.x, size.z, 50);
      camera.position.set(center.x, maxDim * 1.3, center.z + maxDim * 1.6);
      controls.update();
    }
  }

  // ==========================================================================
  // Phase 3: DXF Health Inspector & Compatibility Linter
  // ==========================================================================

  function openInspector() {
    inspectorModal.style.display = 'flex';
  }

  function closeInspector() {
    inspectorModal.style.display = 'none';
  }

  if (btnInspectDxf) btnInspectDxf.addEventListener('click', openInspector);
  if (btnCloseInspector) btnCloseInspector.addEventListener('click', closeInspector);
  if (inspectorModal) {
    inspectorModal.addEventListener('click', (e) => {
      if (e.target === inspectorModal) closeInspector();
    });
  }

  if (btnInspectCurrent) {
    btnInspectCurrent.addEventListener('click', async () => {
      showToast("Generating DXF for inspection...");
      try {
        const exportRes = await fetch('/api/export/dxf', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(getCurrentPayload())
        });
        if (!exportRes.ok) throw new Error("Failed to export current DXF");
        const dxfData = await exportRes.arrayBuffer();

        showToast("Linting DXF with AST Inspector...");
        const inspectRes = await fetch('/api/inspect-dxf', {
          method: 'POST',
          headers: { 'Content-Type': 'application/octet-stream' },
          body: dxfData
        });
        if (!inspectRes.ok) throw new Error("Inspector error: " + await inspectRes.text());
        const report = await inspectRes.json();
        const baseName = (state.text.trim().substring(0, 16) || 'current').replace(/[^a-zA-Z0-9_-]/g, '_');
        report.filename = `${baseName}.dxf (Current Studio Export)`;
        renderInspectorReport(report);
      } catch (err) {
        showToast("Inspection failed: " + err.message);
      }
    });
  }

  if (inspectorDropzone && inspectorFileInput) {
    inspectorDropzone.addEventListener('click', () => inspectorFileInput.click());
    inspectorDropzone.addEventListener('dragover', (e) => {
      e.preventDefault();
      inspectorDropzone.classList.add('drag-over');
    });
    inspectorDropzone.addEventListener('dragleave', () => inspectorDropzone.classList.remove('drag-over'));
    inspectorDropzone.addEventListener('drop', (e) => {
      e.preventDefault();
      inspectorDropzone.classList.remove('drag-over');
      if (e.dataTransfer.files.length > 0) {
        inspectUploadedFile(e.dataTransfer.files[0]);
      }
    });
    inspectorFileInput.addEventListener('change', () => {
      if (inspectorFileInput.files.length > 0) {
        inspectUploadedFile(inspectorFileInput.files[0]);
      }
    });
  }

  async function inspectUploadedFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    showToast(`Analyzing ${file.name}...`);
    try {
      const res = await fetch('/api/inspect-dxf', {
        method: 'POST',
        body: formData
      });
      if (!res.ok) throw new Error(await res.text());
      const report = await res.json();
      renderInspectorReport(report);
    } catch (err) {
      showToast("Analysis error: " + err.message);
    }
  }

  function renderInspectorReport(rep) {
    let verdictClass = 'badge-pass';
    let verdictText = 'PASS // CAD READY';
    let verdictIcon = '✓';

    const criticals = (rep.diagnostics || []).filter(d => d.Severity === 'CRITICAL');
    const warnings = (rep.diagnostics || []).filter(d => d.Severity === 'WARNING');

    if (criticals.length > 0 || !rep.cad_compatible) {
      verdictClass = 'badge-crit';
      verdictText = 'CRITICAL // CAD CONFLICT';
      verdictIcon = '✕';
    } else if (warnings.length > 0) {
      verdictClass = 'badge-warn';
      verdictText = 'WARNING // CHECK SOLVER';
      verdictIcon = '⚠';
    }

    // Target CAD Matrix calculations
    const splines = (rep.entity_counts && rep.entity_counts['SPLINE']) || 0;
    const isAC1015 = rep.version === 'AC1015';
    const hasObjects = (rep.sections || []).includes('OBJECTS');

    let onshapeStatus = 'cad-status-ok', onshapeText = 'PASS (Native)';
    if (splines > 250) {
      onshapeStatus = 'cad-status-warn';
      onshapeText = `SLOW (${splines} Splines)`;
    }
    if (isAC1015 && (!hasObjects || (rep.missing_sections && rep.missing_sections.length > 0))) {
      onshapeStatus = 'cad-status-fail';
      onshapeText = 'FAIL (Corrupt Schema)';
    }

    let fusionStatus = 'cad-status-ok', fusionText = 'PASS (1:1 Sketch)';
    if (isAC1015 && rep.missing_sections && rep.missing_sections.length > 0) {
      fusionStatus = 'cad-status-fail';
      fusionText = 'FAIL (Missing Headers)';
    }

    let autoCADStatus = 'cad-status-ok', autoCADText = 'PASS (Canonical)';
    let freeCADStatus = 'cad-status-ok', freeCADText = 'PASS (Draft / Sketch)';
    let laserStatus = 'cad-status-ok', laserText = 'PASS (LightBurn / CNC)';
    if (splines > 0 && rep.closed_loops === 0) {
      laserStatus = 'cad-status-warn';
      laserText = 'WARN (Check Polyline)';
    }

    const entitySummary = Object.entries(rep.entity_counts || {})
      .map(([k, v]) => `${k}: ${v}`)
      .join(' · ') || 'None';

    const findingsHtml = (rep.diagnostics || []).map(d => {
      let icon = '•';
      if (d.Severity === 'SUCCESS') icon = '✓';
      else if (d.Severity === 'CRITICAL') icon = '✕';
      else if (d.Severity === 'WARNING') icon = '⚠';
      else if (d.Severity === 'INFO') icon = 'ℹ';
      return `<div class="finding-item finding-${d.Severity}">
        <span class="finding-icon">${icon}</span>
        <div>${escapeHtml(d.Message)}</div>
      </div>`;
    }).join('');

    inspectorReport.innerHTML = `
      <!-- Summary Header -->
      <div class="rep-header">
        <div class="rep-file-info">
          <h4>${escapeHtml(rep.filename || 'exported.dxf')}</h4>
          <div class="rep-file-meta">${rep.version_name || rep.version} · ${rep.units || 'Millimeters'} · ${entitySummary}</div>
        </div>
        <div class="badge-verdict ${verdictClass}">${verdictIcon} ${verdictText}</div>
      </div>

      <!-- CAD Matrix -->
      <div class="cad-matrix-wrap">
        <div class="cad-matrix-title">Target CAD & CAM Compatibility</div>
        <div class="cad-matrix-grid">
          <div class="cad-card">
            <div class="cad-card-name">Autodesk Fusion 360</div>
            <div class="cad-card-status ${fusionStatus}">${fusionText}</div>
          </div>
          <div class="cad-card">
            <div class="cad-card-name">PTC Onshape</div>
            <div class="cad-card-status ${onshapeStatus}">${onshapeText}</div>
          </div>
          <div class="cad-card">
            <div class="cad-card-name">AutoCAD (Autodesk)</div>
            <div class="cad-card-status ${autoCADStatus}">${autoCADText}</div>
          </div>
          <div class="cad-card">
            <div class="cad-card-name">FreeCAD</div>
            <div class="cad-card-status ${freeCADStatus}">${freeCADText}</div>
          </div>
          <div class="cad-card">
            <div class="cad-card-name">Laser / LightBurn</div>
            <div class="cad-card-status ${laserStatus}">${laserText}</div>
          </div>
        </div>
      </div>

      <!-- Numerical Extents & Entities -->
      <div class="rep-stats-grid">
        <div class="stat-box">
          <div class="stat-box-val">${rep.width.toFixed(2)} × ${rep.height.toFixed(2)}</div>
          <div class="stat-box-lbl">Dimensions (${rep.units || 'mm'})</div>
        </div>
        <div class="stat-box">
          <div class="stat-box-val">${rep.total_entities}</div>
          <div class="stat-box-lbl">Total Entities</div>
        </div>
        <div class="stat-box">
          <div class="stat-box-val" style="color: #34d399;">${rep.closed_loops}</div>
          <div class="stat-box-lbl">Closed Loops</div>
        </div>
        <div class="stat-box">
          <div class="stat-box-val" style="color: ${rep.open_curves > 0 ? '#f87171' : '#9ca3af'};">${rep.open_curves}</div>
          <div class="stat-box-lbl">Open Curves</div>
        </div>
      </div>

      <!-- Diagnostics findings -->
      <div>
        <div class="findings-title">Integrity & Geometry Diagnostics (${(rep.diagnostics || []).length})</div>
        <div class="findings-list">
          ${findingsHtml}
        </div>
      </div>
    `;
  }

  function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  // Start
  applyPreset('laser');
  loadFonts();
});
