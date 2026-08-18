# Current Task

## Objective

Chunk 1.4 (Research-Oriented Summary) — complete. Phase 1 (Trust & Core Value) now fully complete. Next: Phase 2 (Activation & Retention), starting with Chunk 2.1 (Paper History).

## Requirements

- **Chunk 1.4 (DONE):** Structured research summary — Gemini prompt now outputs `## `-delimited sections (Research Question, Method, Main Findings, Evidence, Limitations, Key Figures, Key Tables, Conclusion). Frontend parses sections and renders each as a distinct card. Flat mode only; chapter mode unaffected. No schema change.

- **Chunk 1.1 (DONE):** Evidence provenance — migration 005, `evidence` table, `EvidenceRepo`, pipeline population, GET exposure, frontend Source chips + collapsible source text.

- **Chunk 1.2 (DONE):** Original vs Explained Figure — `/api/documents/:id/charts/:chartId/image` endpoint + 2-zone UI + chart-level provenance.

- **Chunk 1.3 (DONE):** Figure Explanation Quality — enriched prompts (x_axis, y_axis, key_takeaway, limitations, confidence), frontend display, chart_type bugfix.

- **Provenance scoping (decided):** Evidence rows only where original source text exists — image-origin charts. Chapter-derived charts have no original page mapping → no fabricated provenance (Rule 4).

## Constraints

- Follow `handlers → services → repository/external` layering
- No DB reset needed
- Existing functionality must remain intact

## Relevant Files

- `internal/services/simplification.go` — structured research prompt
- `internal/services/simplification_test.go` — prompt format tests
- `frontend/src/pages/result-page.jsx` — section parsing + card rendering

## Progress

- Chunk 1.4 complete: structured research summary (backend prompt + frontend section cards). 69 Go tests pass, frontend builds clean, gofmt clean.
- Chunk 1.3 complete: enriched prompts + frontend display + chart_type bugfix.
- Chunk 1.1 complete (backend + frontend): evidence provenance fully displayed.
- Chunk 1.2 complete: original-figure serving endpoint + clear original-vs-interpretation UI + tests.
- Chunk 0.2 audit deliverable: `docs/product/current-user-flow.md`.

## Next Action

1. Start Chunk 2.1 (Paper History) — list previously analyzed papers: title, date, status, summary, figures, explanations.
