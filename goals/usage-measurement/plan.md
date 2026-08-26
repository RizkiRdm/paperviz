# Plan — Chunk 6.1 Usage Measurement

## Solution Approach

Add analytics capabilities to PaperViz by:
1. Adding `processing_time_ms` to documents table
2. Creating `analytics_events` table for comparison/share-referral events
3. Building a new analytics repository with aggregate queries
4. Adding `GET /api/analytics` handler
5. Instrumenting pipeline to capture processing time

## Steps

### Step 1: Migration — add columns and table
**Files:** `migrations/012_usage_analytics.sql`
- `ALTER TABLE documents ADD COLUMN processing_time_ms INTEGER`
- `CREATE TABLE analytics_events (id TEXT PRIMARY KEY, event_type TEXT NOT NULL, entity_id TEXT, metadata TEXT, created_at INTEGER NOT NULL)`
- `CREATE INDEX idx_analytics_events_type ON analytics_events(event_type)`
- `CREATE INDEX idx_analytics_events_created ON analytics_events(created_at)`

**Verification:** `sqlite3 paperviz.db ".schema documents"` shows new column; `.schema analytics_events` shows table

### Step 2: Update Document struct and repo
**Files:** `internal/repository/types.go`, `internal/repository/documents.go`
- Add `ProcessingTimeMs *int` to `Document` struct
- Update `Insert` to include new column
- Update `Get` to scan new column
- Add `SetProcessingTime(id string, ms int) error` method

**Verification:** `go build ./...` passes

### Step 3: Create analytics repository
**Files:** `internal/repository/analytics.go`
- `type AnalyticsRepo struct { db dbExecutor }`
- `func (r *AnalyticsRepo) GetSummary() (*AnalyticsSummary, error)` — runs all aggregate queries
- Struct: `AnalyticsSummary` with all metric fields

**Verification:** `go build ./...` passes

### Step 4: Instrument pipeline to capture processing time
**Files:** `internal/services/intake.go`
- Record start time at pipeline entry
- On completion (success or failure), call `docRepo.SetProcessingTime(id, elapsedMs)`

**Verification:** `go build ./...` passes

### Step 5: Track comparison events
**Files:** `internal/handlers/documents.go`
- In `Compare` handler, after successful comparison, insert `analytics_events` row with `event_type='comparison'`

**Verification:** `go build ./...` passes

### Step 6: Track share-referral events
**Files:** `internal/handlers/share.go`
- In `TrackReferral` handler, also insert `analytics_events` row with `event_type='share_referral'`

**Verification:** `go build ./...` passes

### Step 7: Create analytics handler
**Files:** `internal/handlers/analytics.go`
- `type AnalyticsHandler struct { db *sql.DB }`
- `func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request)`
- Calls `AnalyticsRepo.GetSummary()`, returns JSON

**Verification:** `go build ./...` passes

### Step 8: Wire router
**Files:** `internal/handlers/router.go`
- Add `r.With(authMiddleware.RequireAuth).Get("/analytics", analyticsHandler.GetSummary)`

**Verification:** `go build ./...` passes

### Step 9: Add migration to main.go
**Files:** `cmd/server/main.go`
- Add `012: readMigration("migrations/012_usage_analytics.sql")` to migrations map

**Verification:** Server starts, migration applies

### Step 10: Run tests
**Verification:** `go test ./...` passes

## Risks
- Aggregate queries on large document sets could be slow — acceptable for MVP with low volume
- `processing_time_ms` nullable — backfill not needed for existing documents
- No auth on analytics endpoint means any logged-in user sees all metrics — acceptable for solo-dev MVP
