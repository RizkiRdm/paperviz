# Chart Engine Rework — Execution Plan

## Dependency Graph

```
C1 (done) → C2 → C3 → C5 → C6 → C7 → C8 → C9 → C11 → C13 → C14 → C15 → C16 → C17
                  ↓                  ↓         ↓         ↓
                  C4 ──────────────→           C10       C12
```

## Parallel Opportunities

- C3 + C4 can run in parallel (both feed C5)
- C9 + C10 can run in parallel (both after C8)
- C11 + C12 can run in parallel (both after C10)

## Phase 1: Evidence Foundation (C2, C3, C4)

### Task 1: C2 — Numeric Evidence Model
**Files:** `internal/models/evidence.go`, `internal/models/evidence_test.go`
**Work:**
- Define `NumericEvidence` struct with required fields
- Define `EvidenceSource` struct (page, text)
- Add validation: every item has source, generated numbers not evidence
- Table-driven tests: valid evidence, missing source, invalid value

**Skills:** `golang-patterns`, `golang-testing`, `karpathy-guidelines`
**Verify:** `go test ./internal/models/...`

### Task 2: C3 — Extract Numeric Evidence from Text
**Files:** `internal/services/evidence_extract.go`, `internal/services/evidence_extract_test.go`
**Work:**
- Parse text for numeric claims with context
- Preserve source text and page reference
- Ambiguous claims stay ambiguous (no baseline assumption)
- Table-driven tests: clear numbers, ambiguous statements, missing context

**Skills:** `golang-patterns`, `golang-testing`, `error-handling`
**Verify:** `go test ./internal/services/...`

### Task 3: C4 — Extract Structured Data from Tables
**Files:** `internal/services/table_extract.go`, `internal/services/table_extract_test.go`
**Work:**
- Reuse existing table extraction from `internal/services/charts.go`
- Add provenance tracking (page, headers, units)
- Missing cells remain missing, not zero
- Table-driven tests: complete table, missing cells, unit preservation

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/services/...`

## Phase 2: Dataset & Generation (C5, C6, C7)

### Task 4: C5 — Build Candidate Dataset Layer
**Files:** `internal/models/dataset.go`, `internal/services/dataset_build.go`, `*_test.go`
**Work:**
- Define `CandidateDataset` struct (id, title, dimensions, unit, points)
- Build from NumericEvidence + TableData
- One chapter → 0..N datasets
- Table-driven tests: single dataset, multiple datasets, empty chapter

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/services/...`

### Task 5: C6 — Change Chart Generation to Use Evidence
**Files:** `internal/services/charts.go` (modify), `*_test.go`
**Work:**
- Replace chapter→LLM→labels/values with chapter→evidence→dataset→LLM→plan
- LLM decides: useful? dataset? type? title? axes? takeaway?
- LLM must NOT generate numeric values
- No dataset → no chart

**Skills:** `golang-patterns`, `golang-testing`, `error-handling`
**Verify:** `go test ./internal/services/...`

### Task 6: C7 — Create Chart Specification
**Files:** `internal/models/chart_spec.go`, `*_test.go`
**Work:**
- Type-specific schemas: BarSpec, LineSpec, ScatterSpec, PieSpec
- Scatter requires numeric X/Y
- Pie only for parts-of-whole
- Table-driven tests: each chart type, invalid scatter, invalid pie

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/models/...`

## Phase 3: Safety (C8, C9)

### Task 7: C8 — Implement Deterministic Grounding Validator
**Files:** `internal/services/grounding.go`, `internal/services/grounding_test.go`
**Work:**
- 10 validation rules from spec
- Status: verified/partial/unsupported
- unsupported → reject
- Deterministic, no LLM dependency
- Table-driven tests: all 10 rules, each status

**Skills:** `golang-patterns`, `golang-testing`, `error-handling`
**Verify:** `go test ./internal/services/...`

### Task 8: C9 — Fix Missing Value Behavior
**Files:** `frontend/src/components/data-chart.jsx` (verify), `*_test.go`
**Work:**
- Verify TASK-2 fix is complete
- Backend rejects invalid dimensions
- Frontend does not invent numeric values
- Table-driven tests: missing required, missing optional

**Skills:** `impeccable`, `vercel-react-best-practices`, `golang-testing`
**Verify:** `go test ./... && npm run build`

## Phase 4: Multi-Chart & Fallback (C10, C11, C12)

### Task 9: C10 — Remove One-Chart-Per-Chapter Limitation
**Files:** `internal/services/charts.go` (modify), `*_test.go`
**Work:**
- Chapter → evidence → datasets → 0..N charts
- Remove arbitrary limit
- Table-driven tests: zero charts, multiple charts, single chart

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/services/...`

### Task 10: C11 — Keep Image Extraction as Fallback
**Files:** `internal/services/charts.go` (verify), `*_test.go`
**Work:**
- Verify existing image fallback intact
- Image-derived numbers must be grounded
- Unsupported image data → omit chart
- Table-driven tests: image success, image failure, unsupported

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/services/...`

### Task 11: C12 — Add Chart Provenance
**Files:** `internal/models/chart_spec.go` (extend), `internal/services/charts.go` (extend), `*_test.go`
**Work:**
- Add provenance fields: paper_id, page, evidence_ids, source_method, dataset_id, grounding_status
- Survives API serialization
- Table-driven tests: provenance present, serialization roundtrip

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./internal/...`

## Phase 5: Frontend & Observability (C13, C14)

### Task 12: C13 — Improve Frontend Chart Safety
**Files:** `frontend/src/components/data-chart.jsx`, `frontend/src/components/chart-card.jsx`
**Work:**
- Render validated spec only
- No inferring/repairing/inventing
- Show grounding status, source page, limitations
- Tests: render valid spec, reject invalid, show status

**Skills:** `impeccable`, `vercel-react-best-practices`
**Verify:** `npm run build`

### Task 13: C14 — Add Chart Failure Categories
**Files:** `internal/services/charts.go`, `internal/models/chart_spec.go`
**Work:**
- Define failure categories: EXTRACTION_ERROR, DATASET_ERROR, etc.
- Log: paper_id, page, stage, category, message, fallback_used
- Table-driven tests: each category logged correctly

**Skills:** `golang-patterns`, `golang-testing`, `error-handling`
**Verify:** `go test ./internal/services/...`

## Phase 6: Validation & Cleanup (C15, C16, C17)

### Task 14: C15 — Add Regression Tests
**Files:** `internal/services/chart_regression_test.go`
**Work:**
- 8 test cases from spec
- Text comparison, table, ambiguous, missing, scatter, pie, unsupported, image fallback
- All must pass

**Skills:** `golang-testing`
**Verify:** `go test ./internal/services/... -run TestChartRegression`

### Task 15: C16 — Validate with Real Academic Papers
**Files:** `testdata/papers/`, `internal/services/chart_validation_test.go`
**Work:**
- Create test corpus (5-7 papers)
- Measure: grounded/total charts
- Document results

**Skills:** `golang-testing`
**Verify:** `go test ./internal/services/... -run TestRealPaperValidation`

### Task 16: C17 — Final Chart Pipeline Cleanup
**Files:** `internal/services/charts.go`, remove old code
**Work:**
- Remove old prompts that generate numeric values
- Remove unused chart parsing code
- Keep image fallback
- Keep backward compat
- Update comments
- Run all tests

**Skills:** `golang-patterns`, `golang-testing`
**Verify:** `go test ./...`

## Post-Execution

### Task 17: Update Project State
**Files:** `docs/PROJECT_STATE.md`, `AGENTS.md`
**Work:**
- Update "Current Focus" section
- Add shipped chunks to "Shipped / Stable"
- Update "Known Gaps" if any remain
- Update "Where Things Live" with new files

**Skills:** `cavecrew-builder`
**Verify:** Files updated, commit pushed
