# Shareable Figure Explanation — Plan

## Goal
Build public share pages for individual figures with unique URLs, visibility controls, and clear distinction between original figure and PaperViz explanation.

## Solution Approach
1. Add `share_token` column to `charts` table (nullable, unique, indexed)
2. New handler: generate/revoke share tokens for individual charts
3. New handler: serve public figure share page at `/share/fig/{shareToken}`
4. Frontend: new `ShareFigurePage` component for public share view
5. Frontend: add per-chart share button to `ChartCard`

## Steps

### Step 1: Database Migration
**File:** `migrations/009_share_tokens.sql`
- Add `share_token TEXT UNIQUE` to `charts` table
- Add index on `share_token`
- Add `visibility TEXT DEFAULT 'private' CHECK (visibility IN ('public', 'unlisted', 'private'))` to `documents` table (if not present)

**Verification:** `go test ./...` passes, migration loads correctly

### Step 2: Repository Layer
**File:** `internal/repository/chart_repo.go`
- Add `ShareToken *string` field to Chart struct
- Add `SetShareToken(chartID, token string) error`
- Add `GetByShareToken(token string) (*Chart, error)`
- Add `RevokeShareToken(chartID string) error`

**File:** `internal/repository/document_repo.go`
- Add `Visibility` field to Document struct
- Add `SetVisibility(docID, visibility string) error`
- Add `GetVisibility(docID string) (string, error)`

**Verification:** Unit tests for each new method

### Step 3: Service Layer
**File:** `internal/services/share.go` (new file)
- `GenerateShareToken(ctx, db, docID, chartID) (string, error)` — checks doc visibility, generates token
- `RevokeShareToken(ctx, db, docID, chartID) error` — checks ownership, clears token
- `GetSharedFigure(ctx, db, token) (*SharedFigure, error)` — fetches chart + doc, checks visibility, returns combined data

**Verification:** Unit tests with mock DB

### Step 4: Handler Layer
**File:** `internal/handlers/share.go` (new file)
- `POST /api/documents/{id}/charts/{chartId}/share` — requires auth, generates token
- `DELETE /api/documents/{id}/charts/{chartId}/share` — requires auth, revokes token
- `GET /share/fig/{shareToken}` — public, serves share page

**File:** `internal/handlers/router.go`
- Register new routes

**Verification:** `go test ./...` passes

### Step 5: Frontend Share Figure Page
**File:** `frontend/src/pages/share-figure-page.jsx` (new file)
- Fetches figure data from `/share/fig/{shareToken}`
- Displays: original image, explanation text, source attribution, reading level, CTA button
- Visual distinction between original and explanation sections
- Mobile responsive
- Error state for expired/revoked/invalid tokens

**File:** `frontend/src/App.jsx`
- Add route: `/share/fig/:shareToken`

**Verification:** `npm run build` passes

### Step 6: ChartCard Share Button
**File:** `frontend/src/components/chart-card.jsx`
- Add share icon button per chart
- On click: calls POST endpoint, opens ShareDialog with figure URL

**File:** `frontend/src/pages/result-page.jsx`
- Update ShareDialog to handle both document and figure URLs

**Verification:** Manual test: share button generates correct URL

### Step 7: Visibility Integration
**File:** `internal/services/share.go`
- `GetSharedFigure` checks document visibility before serving
- Private docs → 404 on share page
- Unlisted docs → share page works, not in dashboard
- Public docs → share page works, in dashboard

**Verification:** Unit tests for each visibility case

## Verification Commands
```bash
go test ./...           # All Go tests pass
go vet ./...            # No vet errors
cd frontend && npm run build  # Frontend builds
gofmt -l internal/ cmd/  # No format issues
```

## Risks
- Document visibility column may not exist yet — need to check if it was added in prior chunks
- Existing ShareDialog logic needs careful merge to avoid breaking document-level sharing
- Share page must work without auth (public route)
- Chart image serving must work for both authenticated and unauthenticated access

## Open Questions
- Does documents table already have a visibility column? If not, need to add it.
- Should the share page include the full simplified text or just the figure-specific explanation?
