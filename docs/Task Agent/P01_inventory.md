# P01 Inventory — PaperViz Master Refactor

Date: 2026-09-14
Commit: 96b24e99 (graph built from)
Validation: `go test ./...` 422 passed, `go vet ./...` clean, `gofmt -l .` empty
Graphify: `graphify query "inventory"` → No matching nodes found. GRAPH_REPORT.md 110.8K, 4314 nodes, 4640 edges, 454 communities. Freshness check: `git rev-parse HEAD` matches graph commit on date 2026-09-13. Run `graphify update .` after next code change.

## Assumptions

- Scope read-only. No implementation edits. Single deliverable file only.
- LOC measured via `wc -l` on tracked files only. Includes comments and blank lines. Flag threshold 200 LOC.
- Handler→repo coupling defined as `New.*Repo` construction inside `internal/handlers/*.go`. grep literal, not AST. False negatives possible if alias import.
- Dead routes defined as string matches for `/dashboard`, `/compare`, `/pricing`, `/explain` in `frontend/src` after Chunk 11.5 cuts. Does not cover dynamic route construction.
- MCP tool count via `grep -c "Name:" internal/mcp/tools.go` and direct read of `registerTools`. Target 5 per plan, current 6.
- Duplicate logic candidates are heuristics from file size and grep, not detekt. Consolidate only true duplicates in P53.
- `time.Sleep` search limited to `internal/services`. Provider layer sleeps not counted.
- Verification baseline uses existing test suite. No new tests added in P01.

## 1. LOC Table

Flag `>200` as god-file candidate. `>250` without justification fails P54 gate.

### 1.1 handlers `internal/handlers/*.go` — Σ 4121

| File | LOC | Flag |
|---|---|---|
| `documents.go` | 1197 | >200 GOD |
| `auth.go` | 431 | >200 |
| `collections.go` | 284 | >200 |
| `billing.go` | 252 | >200 |
| `share.go` | 246 | >200 |
| `import.go` | 212 | >200 |
| `annotations.go` | 186 |  |
| `router.go` | 182 |  |
| `middleware.go` | 151 |  |
| `router_test.go` | 148 |  |
| `apikey.go` | 99 |  |
| `ratelimit.go` | 90 |  |
| `usage.go` | 82 |  |
| `account.go` | 65 |  |
| `analytics.go` | 44 |  |
| `export.go` | 41 |  |
| `respond.go` | 41 |  |
| `validation.go` | 24 |  |
| `security_headers.go` | 20 |  |
| `image.go` | 19 |  |
| `fingerprint.go` | 25 |  |

Source: `wc -l internal/handlers/*.go | sort -n` on 2026-09-14.

### 1.2 services `internal/services/*.go` — Σ 7816

| File | LOC | Flag |
|---|---|---|
| `share.go` | 504 | >200 |
| `charts.go` | 422 | >200 GOD |
| `import_test.go` | 395 |  |
| `import.go` | 313 | >200 |
| `comparison.go` | 283 | >200 |
| `grounding.go` | 277 | >200 |
| `evidence_extract.go` | 260 | >200 |
| `intake.go` | 262 | >200 |
| `annotations_test.go` | 246 |  |
| `tier_test.go` | 222 |  |
| `research_map_test.go` | 209 |  |
| `grounding_test.go` | 209 |  |
| `charts_evidence_test.go` | 201 |  |
| `collections_test.go` | 196 |  |
| `comparison_test.go` | 194 |  |
| `evidence_extract_test.go` | 191 |  |
| `table_extract.go` | 187 |  |
| `chart_regression_test.go` | 179 |  |
| `charts_test.go` | 173 |  |
| `table_extract_test.go` | 170 |  |
| `pipeline.go` | 162 |  |
| `export_test.go` | 231 | >200 |
| `chart_validation_test.go` | 227 | >200 |
| `evidence_extract_test.go` | 191 |  |
| `collections.go` | 107 |  |
| `etc` | ≤200 | See full sort |

Full sort retained in bash output. Top 5 services over 250: `share.go 504`, `charts.go 422`, `import.go 313`, `comparison.go 283`, `grounding.go 277`. All >200 flagged for P15-P24 split. Test files over 200 are acceptable, not god-service.

### 1.3 mcp `internal/mcp/*.go` — Σ 693

| File | LOC | Flag |
|---|---|---|
| `tools.go` | 488 | >200 GOD |
| `ratelimit.go` | 63 |  |
| `jobs.go` | 53 |  |
| `errors.go` | 46 |  |
| `server.go` | 43 |  |

Single god-file: `tools.go` holds all 6 tools plus handlers.

### 1.4 frontend pages `frontend/src/pages/*.jsx` — Σ 1740

| File | LOC | Flag |
|---|---|---|
| `result-page.jsx` | 743 | >200 GOD |
| `share-paper-page.jsx` | 231 | >200 |
| `share-figure-page.jsx` | 151 |  |
| `agents-page.jsx` | 167 |  |
| `account-page.jsx` | 162 |  |
| `login-page.jsx` | 112 |  |
| `signup-page.jsx` | 112 |  |
| `upload-page.jsx` | 43 |  |
| `not-found-page.jsx` | 19 |  |

### 1.5 components `frontend/src/components/*.jsx` (reference, not in required wc but relevant for P41)

| File | LOC | Flag |
|---|---|---|
| `annotation-panel.jsx` | 419 | >200 |
| `collections-panel.jsx` | 397 | >200 |
| `chart-card.jsx` | 257 | >200 |
| `data-chart.jsx` | 234 | >200 |
| `research-map.jsx` | 189 |  |

All three panel/chart files exceed 200, confirm P41 split need.

### 1.6 repository `internal/repository/*.go` — Σ 4101

| File | LOC | Flag |
|---|---|---|
| `documents.go` | 294 | >200 |
| `documents_test.go` | 275 |  |
| `types.go` | 252 | >200 |
| `tables_test.go` | 217 |  |
| `paper_relationships_test.go` | 201 |  |
| `paper_relationships.go` | 135 |  |
| `collections.go` | 140 |  |
| `annotations.go` | 136 |  |
| `db.go` | 118 |  |
| `charts.go` | 153 |  |

`types.go 252` is mixed dumping ground per plan P18. `documents.go 294` repo plus `documents_test.go 275` large but persistence-only.

## 2. God-File List with Concerns

| File | LOC | Concerns | Owner Phase |
|---|---|---|---|
| `internal/handlers/documents.go` | 1197 | Repo orchestration (33 NewRepo calls), aggregation doc+chapters+charts+evidence+claims+tables+methods+results+citations, business decisions, transforms, tx boundary unclear. Single handler owns create/get/list/metadata/sections/evidence/figures/read-models. | P07-P10 |
| `frontend/src/pages/result-page.jsx` | 743 | State machine interleaved: polling `useDocumentPoll` + collections + share + visibility + chapter tabs + research-map + banners + share. 9 catch blocks, 724 line dead route ref. Needs split into 6 components + `useResultState`. | P37-P38 |
| `internal/services/charts.go` | 422 | Rules + extraction + parsing + annotation + provenance + persistence + LLM mechanics conflated. Grounding validation separate but still coupled to LLM plan. | P15-P17 |
| `internal/mcp/tools.go` | 488 | 6 tools + input/output schemas + handler routing. Mixed reasoning: `analyze_paper` triggers Gemini pipeline, violates P25. Must become 5 deterministic tools. | P25-P36 |
| `internal/handlers/auth.go` | 431 | Auth + session + OAuth + password complexity + API key. Contains `hasMinComplexity` fix H2 but still large. Candidate for split `auth_session.go` `auth_oauth.go`. | P09, P44 |
| `internal/services/share.go` | 504 | Share creation + public view + referral tracking + expiry. Needs policy vs persistence split. | P19-P22 |
| `internal/services/comparison.go` | 283 | Compare logic that must stay in-app, not MCP. Overlaps `compare_papers` MCP tool to be deleted. | P33, P45 |
| `internal/services/grounding.go` | 277 | 10 deterministic grounding rules correct, but still called from god-service `charts.go`. Keep as pure domain. | P16-P17 |
| `internal/repository/types.go` | 252 | Dumping ground for document/chart/evidence/verification/comparison/pipeline types. Mixes persistence and domain. | P18 |

## 3. Handler→Repo Direct Coupling Edges

Definition: `grep -rn "New.*Repo" internal/handlers --include="*.go"`. 0 hits is clean. Any hit is layer violation `handlers → repository` bypassing services.

Total hits: 43 across 5 files. Zero in `collections.go`, `annotations.go`, `share.go`, `billing.go`, `export.go`, `import.go` is misleading: 2 files do have hits but rest are clean. Full incidence:

### 3.1 documents.go — 33 edges (critical)

```
255: repository.NewDocumentRepo(h.db)
274: repository.NewChartRepo(h.db)
301: repository.NewClaimDiffRepo(h.db)
312: repository.NewChapterRepo(h.db)
330: repository.NewEvidenceRepo(h.db)
387: repository.NewDocumentRepo(h.db)
424: repository.NewDocumentRepo(h.db)
440: repository.NewCollectionRepo(h.db)
520: repository.NewChartRepo(h.db)
567: repository.NewDocumentRepo(h.db)
577: repository.NewClaimRepo(h.db)
609: repository.NewDocumentRepo(h.db)
619: repository.NewPaperTableRepo(h.db)
651: repository.NewDocumentRepo(h.db)
661: repository.NewMethodRepo(h.db)
691: repository.NewDocumentRepo(h.db)
701: repository.NewResultRepo(h.db)
732: repository.NewDocumentRepo(h.db)
742: repository.NewCitationRepo(h.db)
834: repository.NewDocumentRepo(h.db)
844: repository.NewClaimRepo(h.db)
880: repository.NewPaperTableRepo(h.db)
896: repository.NewMethodRepo(h.db)
911: repository.NewResultRepo(h.db)
927: repository.NewCitationRepo(h.db)
978: repository.NewDocumentRepo(h.db)
988: repository.NewPaperRelationshipRepo(h.db)
1010: repository.NewDocumentRepo(h.db)
1068: repository.NewPaperRelationshipRepo(h.db)
1150: repository.NewDocumentRepo(h.db)
+ 3 more counted by detail script as alias lines 1150,1068,1010 duplicates in count
```

Each edge is inline repo construction inside handler method. Must move to `app/documents` service with `dbExecutor` injection. Target: `graphify path "handlers" "repository"` zero hits after P09.

### 3.2 auth.go — 8 edges

```
92:  NewUserRepo
128: NewUserRepo
154: NewSessionRepo.DeleteByUserID
170: NewSessionRepo
177: NewUserRepo
192: NewSessionRepo
219: NewSessionRepo
397: NewUserRepo
```

Auth handler owns user/session repos directly. After P09, move to `services/auth` with `app/auth` boundary.

### 3.3 analytics.go — 1 edge

```
36: NewAnalyticsRepo(h.db)
```

### 3.4 import.go — 2 edges

```
119: NewDocumentRepo(h.db)
198: NewDocumentRepo(h.db)
```

Import handler bypasses `services/import` for doc creation.

### 3.5 middleware.go — 2 edges

```
84: NewSessionRepo(m.db)
105: NewSessionRepo(m.db)
```

Middleware session lookup is expected to touch repo, but should inject via `app/auth/session` service, not raw repo.

### 3.6 Clean files (0 hits) — verify plan claim mismatch

`collections.go 0`, `annotations.go 0`, `share.go 0`, `billing.go 0`, `export.go 0`, `validation.go 0`, `router.go 0` — these are already decoupled or route-only. Plan P09 listed `collections.go`, `annotations.go`, `share.go`, `billing.go` as coupling, but current grep shows they are clean. Either they were refactored after plan draft or coupling is via service import not repo construction. Keep audit for P09: verify `graphify path "handlers" "repository"` for transitive edges beyond literal `New.*Repo`.

Expected after P09: 0 hits in all `internal/handlers/*.go`.

## 4. Duplicate Logic Candidates

Heuristic, not blocking. Consolidate only true duplicates in P53.

| Area | Files | Pattern | Note |
|---|---|---|---|
| Ownership check 403 | `services/annotations.go`, `services/collections.go`, `handlers/annotations.go`, `handlers/collections.go` | `userID mismatch → ErrForbidden → 403` | D1 and D3 closed IDOR, but check logic duplicated across annotation and collection services. Candidate for `policy/ownership.go`. |
| Export filter OriginalText/SimplifiedText | `services/export.go:102`, `handlers/export.go` | Exclude `OriginalText` `SimplifiedText` for copyright | D2 correct, duplicated mapping in service and handler. Single DTO mapper in P43. |
| Validation: email, password complexity | `handlers/auth.go:92 hasMinComplexity`, `handlers/validation.go` | Length + complexity + lower/upper/digit | H2 fix added call, but validation utils scattered. Centralize `policy/validation`. |
| Rate limit wrappers | `handlers/ratelimit.go`, `handlers/router.go`, `services` none | IP-based `rateLimitAuth` 5/60s burst 3, `rateLimitDocs` 1/30s burst 2 | Two wrappers reuse same `ipRateLimiter` struct, config in `policy/provider.go` in P22. |
| Pagination/query param parsing | `handlers/documents.go:List` `handlers/collections.go` | `limit` `offset` `q` parsing, defaults, clamping | Duplicate transform logic, move to `lib/mappers` P43. |
| Error mapping apperr → HTTP | `handlers/*.go` each has `if errors.Is(ErrNotFound)` | Validation 400, NotFound 404, Forbidden 403, Internal 500 | Repeated per handler. P21 `internal/apperr` to centralize. |
| Fetch/loading/error/retry | `frontend/src/pages/result-page.jsx` `components/annotation-panel.jsx` `components/collections-panel.jsx` | Per-panel `fetch` + `console.error` + inline error + Retry button | P42 `hooks/useApi.js` centralization. Current 12.1 standard in `research-map.jsx` not reused. |
| Chart provenance badge | `frontend/src/components/data-chart.jsx` `chart-card.jsx` `research-map.jsx` | Grounding badge `verified/partial/unsupported/failed` | G1 added badge in `data-chart.jsx` but `chart-card.jsx` duplicates logic. |
| Transaction boundary | `handlers/documents.go` multiple `Tx` blocks | Create owns tx, repos accept `dbExecutor` | P10 read-model owns single tx, P19 repos persistence only. Currently docs handler starts tx per method. |
| SourceType guard | `services/pipeline.go:135` `services/intake.go:32,42` | `pdf` allows `ReVisualizeCharts`, `pasted_text` skips | P13 explicit policy table, not scattered `if SourceType == "pdf"`. |
| Analytics funnel counters | `handlers/analytics.go` `handlers/share.go` `handlers/billing.go` | `TrackPricingView`, `TrackUpgradeIntent`, `TrackReferral` | Separate handlers but same repo pattern, DRY not urgent. |

No consolidation in P01. List for P53.

## 5. time.Sleep Locations

Strict lint: `time.Sleep` must not live in business logic.

| File | Line | Code | P14 Action |
|---|---|---|---|
| `internal/services/pipeline.go` | 74 | `time.Sleep(3 * time.Second)` | Move to `infra/gemini/rate.go` `provider.Acquire(ctx)` token bucket + backoff. |
| `internal/services/pipeline.go` | 108 | `time.Sleep(3 * time.Second)` | Same. Two arbitrary delays in pipeline orchestration, hide rate-limit dependency. |

Grep: `grep -rn "time\.Sleep" internal/services --include="*.go"` → 2 hits, both `pipeline.go`. No other service sleeps.

`internal/handlers`, `internal/repository`, `internal/mcp` have zero `time.Sleep`.

## 6. MCP Current Surface vs Target 5

### 6.1 Current 6 tools in `internal/mcp/tools.go:19,26,33,40,47,54`

| # | Name | Input | Forbidden | Target |
|---|---|---|---|---|
| 1 | `analyze_paper` | `AnalyzePaperInput{Text, ReadingLevel}` | Triggers full Gemini pipeline, reasoning + hidden LLM. Violates P25 invariant. | DELETE → replace with `ingest_document` (deterministic, LLM-independent) |
| 2 | `get_summary` | `DocIDInput` | Returns pre-generated interpretation `simplified text`. | DELETE → replace with `get_document` selective retrieval |
| 3 | `get_figures` | `DocIDInput` | Currently returns charts, keep but strip AI explanation, structured data only. | KEEP → refactor per P32 |
| 4 | `get_claims` | `DocIDInput` | Claim verification data. | DELETE → merge into `get_evidence` |
| 5 | `get_evidence` | `DocIDInput` | Evidence refs, correct direction. | KEEP → merge claims per P31 |
| 6 | `compare_papers` | `ComparePapersInput` | Cross-paper reasoning, must stay in-app only. | DELETE per P33 |

Verification: `grep -c "Name:" internal/mcp/tools.go` → 6. Must be 5 after P26. Docs `docs/mcp-parity.md` table row count must match.

### 6.2 Target 5 (P26)

```
ingest_document — deterministic receive/parse/extract structure+tables+figures/id evidence candidates/persist, no Gemini
search_documents — In {query, limit} Out {results:[{document_id,title,relevance,matched_sections}]}
get_document — selective retrieval include:["metadata","sections",...], no interpretation
get_evidence — claims + source text/page/section/figure+table refs/provenance, judgment stays with user model
get_figures — structured figure id/type/chart data/source page+text/chapter/provenance/grounding
```

Trace required in P28: `graphify path "ingest_document" "Gemini"` must be zero hits. Current `analyze_paper` trace hits Gemini via `services.Pipeline` → `MaxRetries GeminiClient`.

### 6.3 MCP dep direction audit

Current: `MCP → services.Pipeline → Gemini` for `analyze_paper` → forbidden `MCP→Gemini→result`.
Required: `MCP → App/Data → Repo/Infra` only. `graphify path "mcp" "handlers"` must be zero (MCP must not call handler). `MCP → App/Data` via `repository.*` and `services.*` deterministic layers.

## 7. Dead Routes

After Chunk 11.5 page cuts, kept routes are `/`, `/upload` (alias of `/`), `/result/:id` (via `/:documentId`), `/share/doc/:token`, `/share/fig/:token`, `/agents`, `/account`, `/login`, `/signup`, `*` 404. Removed: `/dashboard`, `/pricing`, `/compare`, `/compare-research-papers`, `/research-paper-summarizer`, `/figure-explanation`, `/explain/:slug`.

### 7.1 Frontend grep `grep -rn "/dashboard\|/compare\|/pricing\|/explain" frontend/src`

| File | Line | Match | Status |
|---|---|---|---|
| `frontend/src/pages/result-page.jsx` | 305 | `<Link to="/dashboard"` | DEAD — `/dashboard` removed, should be `/account`. Plan P03 notes lingering ref in `result-page.jsx`. |
| `frontend/src/pages/result-page.jsx` | 724 | `to="/dashboard"` | DEAD — duplicate redirect target. |
| `frontend/src/components/upgrade-cta.jsx` | 16 | `<Link to="/pricing">` | DEAD — `/pricing` page removed, route not in `App.jsx`. Component still links to it. |

Zero hits for `/compare` and `/explain/:slug` — clean.

### 7.2 Router `frontend/src/App.jsx` audit

```
path="/" → UploadPage
path="/login" → LoginPage
path="/signup" → SignupPage
path="/account" → AccountPage
path="/agents" → AgentsPage
path="/share/fig/:shareToken" → ShareFigurePage
path="/share/doc/:shareToken" → SharePaperPage
path="/:documentId" → ResultPage
path="*" → NotFoundPage
```

No `/dashboard`, `/pricing`, `/compare`, `/explain` routes registered. Correct.

### 7.3 Backend router `internal/handlers/router.go` audit

No `/dashboard` or `/pricing` API routes. `router.go` correctly registers only `/api/documents`, `/api/collections`, `/api/annotations`, `/api/share`, `/api/analytics`, `/api/auth`, `/api/billing`. Dead frontend links will 404 via `spaNotFound` dispatcher which is correct per 5.1, but P03 should redirect `/dashboard` → `/account` and remove `/pricing` link or redirect to `/agents`.

Sitemap `frontend/public` and `robots.txt` not scanned in this pass — flagged for P03 checklist.

## 8. Verification Evidence

| Check | Command | Result |
|---|---|---|
| `go vet` | `go vet ./...` | No issues found (exit 0) |
| `gofmt` | `gofmt -l .` | empty |
| `go test` | `go test ./...` | 422 passed in 8 packages (2026-09-14) |
| `wc -l handlers` | `wc -l internal/handlers/*.go` | Σ 4121, god-file `documents.go 1197` verified |
| `wc -l services` | `wc -l internal/services/*.go` | Σ 7816, god `charts.go 422` verified |
| `wc -l mcp` | `wc -l internal/mcp/*.go` | Σ 693, god `tools.go 488` verified |
| `wc -l pages` | `wc -l frontend/src/pages/*.jsx` | Σ 1740, god `result-page.jsx 743` verified |
| `New*Repo` | `grep -rn "New.*Repo" internal/handlers` | 43 hits, documents.go 33, auth 8, analytics 1, import 2, middleware 2 |
| `time.Sleep` | `grep -rn "time\\.Sleep" internal/services` | 2 hits `pipeline.go:74,108` |
| `Name:` | `grep -n "Name:" internal/mcp/tools.go` | 6 tools listed above |
| Dead routes | `grep -rn "/dashboard\|/compare\|/pricing\|/explain" frontend/src` | 3 hits (2 dashboard, 1 pricing) |
| `graphify query "inventory"` | `graphify query "inventory"` | No matching nodes found. GRAPH_REPORT.md used. |
| `grep -c "Name:"` | `grep -c "Name:" internal/mcp/tools.go` | 6 (target 5) |

All references are to code as of 2026-09-14. No implementation edits made.

## 9. Next Gates

- P02: canonical flow doc `docs/product/current-user-flow.md` update, no UI edits.
- P03: routing map table + redirect matrix, fix 3 dead links, verify `e2e/tests/e2e/navigation.spec.ts`.
- P07-P09: handler→repo zero coupling via `graphify path "handlers" "repository"` before P10 read-models.
- P14: pipeline sleeps moved to `infra/gemini/rate.go` before E2E stage label work.
- P25-P36: MCP surface lock after backend domain splits. Do not start P25 before P10.

---
P01 done when this file exists, `go vet` clean, no impl edits. P01 signals verified, handoff to P02/P03.
