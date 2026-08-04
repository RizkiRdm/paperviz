# Engineering Decisions

Historical entries are append-only. Deprecate decisions instead of deleting them.

## 2026-08-04 — Chart type switching via chartData.type

- **Context:** `data-chart.jsx` only rendered BarCharts. Pipeline generates charts with `type` field (bar/line/pie/scatter) but frontend ignored it.
- **Decision:** Switch on `chartData.type` and render the correct Recharts component (BarChart, LineChart, PieChart, ScatterChart). Pie uses `Cell` with color array. Scatter uses category X-axis.
- **Alternatives considered:** Always render bar; add a chart-type selector UI.
- **Consequences:** Charts now render in their intended type. Pie charts show percentage labels. No new dependencies.

## 2026-08-04 — Processing stage via DB column + pipeline callback

- **Context:** Processing spinner showed generic "Simplifying & Verifying..." for entire 30-60s pipeline. User had no visibility into progress.
- **Decision:** Add `processing_stage TEXT` column to documents table. Pipeline receives `OnStage` callback, calls it at each stage transition (simplifying/verifying/generating_charts). Handler wires callback to `docRepo.UpdateStage`. GET response returns `processing_stage`. Frontend maps stages to labels.
- **Alternatives considered:** Return stage from GET without persisting (in-memory only); use existing `error_message` field for stage.
- **Consequences:** DB schema change requires reset. Stage is cleared on pipeline completion (`processing_stage = NULL` in UpdateStatus). Minimal overhead — one UPDATE per stage transition.

## 2026-08-04 — Share dialog as modal overlay

- **Context:** CopyLinkButton was inline in header. For a shareable-link product, a proper share dialog with URL, copy, and expiry info is expected.
- **Decision:** New `ShareDialog` component rendered as fixed overlay. Shows URL input (read-only, auto-select on focus), copy button, 7-day expiry note. Triggered from header "Share" button.
- **Alternatives considered:** Keep inline CopyLinkButton; add a dropdown popover.
- **Consequences:** One new component inside result-page.jsx (not a separate file — avoids unnecessary file proliferation for MVP). Modal dismisses on backdrop click or X button.

## 2026-08-02 — Load Recharts on demand

- **Context:** React Doctor reported the static Recharts import as expensive initial-page JavaScript.
- **Decision:** `ChartCard` lazy-loads `DataChart`; `DataChart` uses runtime `import("recharts")`, handles load failure, and ignores completion after unmount.
- **Alternatives considered:** Keep the eager import; move a static import into a lazy wrapper; dynamic import in the existing component.
- **Consequences:** Initial app JS falls from about 598 kB to 247 kB; Recharts becomes a separate async chunk. Chart display briefly shows existing-token loading text.

## 2026-08-02 — Validate selected files in UploadPage

- **Context:** `UploadDropzone` receives files, while `UploadPage` owns selected-file and error state plus server-aligned messages.
- **Decision:** Validate size and MIME type in `UploadPage.onFileChange`; invalid files clear selection and reuse `file_too_large` or `invalid_file_type` messages.
- **Alternatives considered:** Export error messages into the dropzone; pass validation errors through new props; duplicate checks in selection and submit paths.
- **Consequences:** Picker and drag/drop share one validation boundary without prop drilling. Strict MIME checking may reject valid PDFs whose browser reports an empty MIME type.

## 2026-08-02 — Use native semantics for upload interaction

- **Context:** Clickable dropzone markup was a static `div`, inaccessible by keyboard and unnamed as an interactive control.
- **Decision:** Use a native `button type="button"` beside the hidden file input and add a visible associated label to pasted-text textarea.
- **Alternatives considered:** Add `role`, `tabIndex`, and custom keyboard handlers to the `div`.
- **Consequences:** Browser-native focus, keyboard activation, and screen-reader semantics replace custom behavior.
