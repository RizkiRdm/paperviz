# Product Requirements Memory

Canonical specification: [`PRD.md`](./PRD.md). Do not duplicate product requirements here.

## Current Status

- **Chunk 0.2 (Product Flow Audit) — done:** `docs/product/current-user-flow.md` maps the current journey. Key finding: value delivered on first try, but zero retention loop (no history/saved papers). Phase 2 before Phase 4.
- **Chunk 1.1 (Evidence Provenance) — in progress:** product requirement is that every AI-generated explanation is traceable to the original paper (page, figure/table, section, source ref). Backend model implemented; display/population pending.
- Client enforces the existing 20 MB PDF and `application/pdf` requirements before upload.
