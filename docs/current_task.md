# Current Task

## Objective

Chapter-tabbed view implemented. Awaiting DB reset + rebuild + testing.

## Requirements

- **Backend:**
  - Migration 004: `chapter_id` column on charts table
  - Services `Chart.ChapterIndex` field
  - Repository `Chart.ChapterID` field
  - `Insert` + `ListByDocument` handle `chapter_id`
  - Handler exposes `content` on chapters, `chapter_id` on charts
  - `saveResult` links charts to chapters via `chapterIndexToID` map

- **Frontend:**
  - Horizontal scrollable tabs (only when 2+ chapters)
  - ARIA roles: `tablist`, `tab`, `tabpanel`
  - Keyboard navigation (arrow keys)
  - Chapter-scoped content + charts
  - Fallback to linear view when 0-1 chapters

- **UX copy:**
  - Error messages use "We couldn't..." pattern
  - Processing text uses student-friendly language
  - Chart empty states simplified

## Constraints

- DB reset required for migration 004 (schema change)
- Rebuild required after DB reset
- Live Gemini regression tests still outstanding

## Relevant Files

- `migrations/004_chapter_charts.sql` — new migration
- `internal/services/types.go` — Chart.ChapterIndex field
- `internal/services/charts.go` — GenerateChapterChart sets ChapterIndex
- `internal/repository/types.go` — Chart.ChapterID field
- `internal/repository/charts.go` — Insert + ListByDocument handle chapter_id
- `internal/handlers/documents.go` — chapterResponse.Content, chartResponse.ChapterID, saveResult links charts to chapters
- `frontend/src/pages/result-page.jsx` — tabbed chapter view with ARIA + keyboard nav

## Progress

- All backend + frontend changes implemented.
- Tests pass (21/21).
- Design detector clean (0 findings).
- Committed but not yet pushed.

## Next Action

1. DB reset: `rm paperviz.db*`
2. Rebuild: `go build -o paperviz ./cmd/server`
3. Test: Upload PDF with chapters → verify tabbed view
4. Push: `git push`
