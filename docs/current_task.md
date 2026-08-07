# Current Task

## Objective

P0 Core Features COMPLETE. Await user direction for next phase.

## Requirements

- Batch A — Claim-diff verification:
  - A1: Backend exposes `claim_diff` in GET response (already persisted rows wired through)
  - A2: Frontend expandable claim comparison panel under Verified badge
- Batch B — Chapter structure:
  - B1: Migration `003_chapters.sql` (chapters table + index)
  - B2: `ChapterRepo` + `Chapter` type in repository layer
  - B3: Pipeline carries chapters through `PipelineOutput`
  - B4: `saveResult` persists chapters, Get handler exposes them
  - B5: Frontend "Sections in this paper" summary card above article

## Constraints

- No new Gemini calls for claim-diff (reads already-persisted rows)
- Chapter detection already existed (`DetectChapters`) — only persistence was missing
- One concern per chunk, sequential execution, BUKTI SELESAI after each
- DESIGN.md tokens: `rounded-[12px]`, `border-[#e5e5e5]`, `bg-[#f5f5f5]`

## Relevant Files

- `internal/handlers/documents.go` — getDocumentResponse, saveResult, Get handler
- `internal/repository/chapters.go` — ChapterRepo (new)
- `internal/repository/types.go` — Chapter struct (new)
- `internal/services/types.go` — PipelineOutput.Chapters field
- `internal/services/pipeline.go` — passes chapters through
- `migrations/003_chapters.sql` — new migration
- `frontend/src/components/ui/status-banners.jsx` — VerificationBadge (onClick), ClaimComparisonPanel (new)
- `frontend/src/pages/result-page.jsx` — chapter summary card, claim toggle state

## Progress

- All 7 chunks committed and pushed (0a9df0f through be738e9).
- Go build passes, npm build passes.
- DB reset required (new chapters table).

## Next Action

Await user direction for next phase.
