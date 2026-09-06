> ARCHIVED 2026-09-04 — superseded by docs/PRD.md. Do not use
> this file as canonical product requirements.

# Product Requirements Memory

Canonical specification: [`PRD.md`](./PRD.md). Do not duplicate product requirements here.

## Current Status

- **Chunk 0.2 (Product Flow Audit) — done:** `docs/product/current-user-flow.md` maps the current journey. Key finding: value delivered on first try, but zero retention loop (no history/saved papers). Phase 2 before Phase 4.
- **Chunk 1.1 (Evidence Provenance) — backend done:** every AI-generated explanation traceable to the original paper (page, figure/table, section, source ref). Evidence rows populated for image-origin figures + exposed in API. Frontend display of provenance pending.
- **Chunk 1.2 (Original vs Explained Figure) — done:** users clearly see original figure vs generated interpretation; provenance (page/chapter) shown. Provenance is chart-level only; evidence-table provenance is backend-only pending frontend display.
- Client enforces the existing 20 MB PDF and `application/pdf` requirements before upload.
