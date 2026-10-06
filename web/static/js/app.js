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
    datum: "bottom-left",
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
    view_mode: "wireframe", // "wireframe" | "fill"
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
        body: JSON.stringify({
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
        })
      });

      if (!res.ok) {
        const errText = await res.text();
        console.warn('Preview error:', errText);
        return;
      }

      const data = await res.json();

      // Render SVG
      svgContainer.innerHTML = data.svg;
      applyViewMode();

      // Update Readouts
      readoutDims.textContent = `${data.width.toFixed(2)} × ${data.height.toFixed(2)} ${state.units}`;
      readoutDatum.textContent = formatDatumLabel(state.datum);
      readoutLayer.textContent = `${state.layer_name} (${getLayerColorName(state.layer_name)})`;
      if (readoutTopology) {
        readoutTopology.textContent = data.welded ? "Manifold (Welded)" : "Standard";
        readoutTopology.style.color = data.welded ? "#34d399" : "";
      }
      if (cadModeTag) {
        cadModeTag.textContent = state.dxf_format === 'spline' ? 'AC1015 Spline' : 'AC1009 Poly';
      }
      if (shapingModeTag) {
        const parts = [];
        if (state.arc_enabled) parts.push("Arc");
        if (state.slant_angle !== 0) parts.push(`${state.slant_angle > 0 ? '+' : ''}${state.slant_angle}°`);
        if (state.offset !== 0) parts.push(`${state.offset > 0 ? '+' : ''}${state.offset}${state.units}`);
        shapingModeTag.textContent = parts.length > 0 ? parts.join(' · ') : 'Direct';
        shapingModeTag.style.color = parts.length > 0 ? '#60a5fa' : '';
      }
      readoutStats.textContent = `${data.glyph_count} Glyphs | ${data.path_count} Loops`;

      // Update CLI snippet
      lastCLICommand = data.cli_command;
      cliPreviewCode.textContent = lastCLICommand;

      // Position Datum Crosshair relative to SVG box
      updateDatumCrosshair();
    } catch (err) {
      console.error('Preview fetch failed:', err);
    }
  }

  function formatDatumLabel(datum) {
    switch (datum) {
      case 'bottom-left': return 'Bottom-Left (0,0)';
      case 'center': return 'Center (0,0)';
      case 'top-left': return 'Top-Left (0,0)';
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
    if (!svgElem) return;

    const rect = svgElem.getBoundingClientRect();
    const wrapperRect = canvasWrapper.getBoundingClientRect();

    // Compute relative origin in wrapper
    let x = 0, y = 0;
    if (state.datum === 'bottom-left') {
      x = (rect.left - wrapperRect.left);
      y = (rect.bottom - wrapperRect.top);
    } else if (state.datum === 'center') {
      x = (rect.left - wrapperRect.left) + rect.width / 2;
      y = (rect.top - wrapperRect.top) + rect.height / 2;
    } else if (state.datum === 'top-left') {
      x = (rect.left - wrapperRect.left);
      y = (rect.top - wrapperRect.top);
    }

    originIndicator.style.left = `${x}px`;
    originIndicator.style.top = `${y}px`;
  }

  function applyViewMode() {
    if (state.view_mode === 'wireframe') {
      svgContainer.classList.add('view-mode-wireframe');
      svgContainer.classList.remove('view-mode-fill');
      btnModeWireframe.classList.add('active');
      btnModeFill.classList.remove('active');
    } else {
      svgContainer.classList.remove('view-mode-wireframe');
      svgContainer.classList.add('view-mode-fill');
      btnModeWireframe.classList.remove('active');
      btnModeFill.classList.add('active');
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

  // Datum Buttons
  datumButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      datumButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      state.datum = btn.dataset.datum;
      triggerUpdate();
    });
  });

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
        body: JSON.stringify({
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
        })
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

  // Start
  applyPreset('laser');
  loadFonts();
});
