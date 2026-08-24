
**Last updated:** 2026-08-24 — Chunk 4.1 merged: shareable figure explanation with public share pages

- **Working on:** Chunk 4.2 — shareable paper explanation
- **Active task file:** `goals/shareable-figure-explanation/`
- **Blocked on / pending decision:** none
- **Next action if resuming:** review 4.2 scope in roadmap, check if document-level share endpoint exists or needs creation

## Known Gaps
- Share page doesn't include interactive Recharts render — only image + text explanation: deferred to avoid scope creep in 4.1

## Key Decisions Log
| Decision | Choice | Why |
|---|---|---|
| Token generation timing | Lazy (on first share) | Avoids wasting tokens on documents never shared |
| Figure visibility | Inherit document visibility | Simpler, no per-figure toggle needed for MVP |
| Image serving auth | No extra check on GetChartImage | Share service gates data layer; URL only exposed for non-private docs |

## Where Things Live
- `internal/services/share.go` — GenerateShareToken, RevokeShareToken, GetSharedFigure
- `internal/handlers/share.go` — POST/DELETE token endpoints, GET /share/fig/{token}
- `frontend/src/pages/share-figure-page.jsx` — public share page component
- `migrations/009_share_tokens.sql` — share_token + visibility columns
