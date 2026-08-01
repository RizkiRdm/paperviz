# Engineering Decisions

Historical entries are append-only. Deprecate decisions instead of deleting them.

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
