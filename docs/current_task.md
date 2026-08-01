# Current Task

## Objective

Commit and push the verified React Doctor top-three fixes and chunk 4 client-side file validation.

## Requirements

- Defer Recharts until a data-extracted chart is rendered.
- Use native interactive semantics for PDF selection.
- Give pasted-text textarea a persistent visible label.
- Reject files over `20 * 1024 * 1024` bytes with existing `file_too_large` message.
- Reject `file.type !== "application/pdf"` with existing `invalid_file_type` message.

## Constraints

- Follow `DESIGN.md` tokens and existing component patterns.
- Do not fix the 10 remaining `result-page.jsx` React Doctor findings in this pass.
- Do not stage unrelated working-tree changes.
- Use DOM/runtime evidence instead of image-model analysis.

## Relevant Files

- `frontend/src/components/chart-card.jsx`
- `frontend/src/components/data-chart.jsx`
- `frontend/src/components/upload-dropzone.jsx`
- `frontend/src/pages/upload-page.jsx`

## Progress

- Build passes; Recharts is emitted as a separate async chunk.
- React Doctor confirms the requested three diagnostics are gone.
- Playwright confirms lazy chart rendering, semantic controls, associated label, both validation errors, and valid-file recovery.
- Full Go tests pass: 17 tests across 5 packages.
- Graphify index refreshed.

## Next Action

Collect independent review results, resolve blockers if any, create atomic commits, push `main`, then plan chunk 5.
