# PaperViz Full E2E + Security + Performance + MCP Audit Report

**Date:** 2026-09-17 01:00 WIB
**Test Environment:** localhost:6767 (Go server, SQLite, Vite-built frontend)
**Test Paper:** Attention Is All You Need (arXiv:1706.03762v7)
**Tester Persona A:** Academic user (upload, paste, share, export)
**Tester Persona B:** Code agent via MCP (Claude Code connecting to PaperViz MCP server)
**Server Build:** HEAD commit `ed0559d` (audit: P50-P61 final audits pass), WIP uncommitted changes

---

## Executive Summary

| Category | Status |
|----------|--------|
| Critical Bugs Found | 2 |
| High Bugs Found | 4 |
| Medium Bugs Found | 3 |
| Low / Informational | 1 |
| Security Vulnerabilities | 2 (1 CRITICAL, 1 HIGH) |
| MCP Server Status | BROKEN (WIP uncommitted code panics at startup) |
| Frontend Status | PARTIAL (CSP blocks fonts, auth flow broken by backend bugs) |
| Pipeline Status | WORKS for pasted text; PDF image extraction broken |

---

## Bug Inventory

### BUG-01: Auth Session Self-Destruction [CRITICAL]
**File:** `internal/handlers/auth.go:146-157`
**Severity:** CRITICAL — blocks all authenticated user flows
**Symptom:** Every login/signup creates a session, then immediately deletes ALL sessions for that user (including the one just created). Result: no user can stay logged in.

**Root Cause:** In the Login handler, `createSessionAndSetCookie()` is called first (line 147), then `DeleteByUserID()` runs on line 154 — which deletes ALL sessions for that user, including the one just created.

```go
// Line 147: session created + cookie set
if err := h.createSessionAndSetCookie(w, user.ID); err != nil { ... }

// Line 154: THIS DELETES THE SESSION JUST CREATED
if err := repository.NewSessionRepo(h.db).DeleteByUserID(user.ID); err != nil {
    slog.Error("delete old sessions failed", "error", err)
    // Non-fatal: new session is already set on the cookie
}
```

**Reproduction:**
```bash
# Login → 200 OK
curl -c cookies.txt -X POST http://localhost:6767/api/auth/login \
  -d '{"email":"user@test.com","password":"pass"}'  # → 200

# Immediately check auth → 401 (session deleted)
curl -b cookies.txt http://localhost:6767/api/auth/me  # → 401

# Verify DB: sessions table empty
sqlite3 paperviz.db "SELECT count(*) FROM sessions;"  # → 0
```

**Impact:** 100% of authed users affected. All protected pages (account, agents, saved papers) inaccessible.

**Fix:** Swap the order — delete old sessions BEFORE creating the new one. Or scope `DeleteByUserID` to exclude the current session token.

---

### BUG-02: Missing Migration 017 Registration [HIGH]
**File:** `cmd/server/main.go:33-50`, `cmd/mcp/main.go:26-44`
**Severity:** HIGH — blocks OAuth users, causes 500 on `/api/auth/me`
**Symptom:** Migration file `017_oauth_columns.sql` exists on disk but is NOT registered in `loadMigrations()` in either entrypoint. DB schema at v16 lacks `oauth_provider` column → every authed user fetch 500s.

**Root Cause:** Both `cmd/server/main.go` and `cmd/mcp/main.go` have hardcoded migration maps that stop at v16. Migration 017 (OAuth columns) was added later but never registered.

**Reproduction:**
```bash
# After fresh DB init, server starts but /api/auth/me always 500s
curl -b valid-session.txt http://localhost:6767/api/auth/me
# → 500 {"error":"internal_error"}
# Server log: "no such column: oauth_provider"
```

**Impact:** All non-OAuth users broken (signup creates user but can never fetch it).

**Fix:** Add `"017_oauth_columns.sql"` to both `loadMigrations()` maps in `cmd/server/main.go` and `cmd/mcp/main.go`.

---

### BUG-03: GetByID NULL Scan on OAuth Columns [HIGH]
**File:** `internal/repository/users.go:53-62`
**Severity:** HIGH — compounds BUG-02
**Symptom:** Even after manually applying migration 017, non-OAuth users (where `oauth_provider` is NULL) cause `GetByID` to crash with "converting NULL to string is unsupported".

**Root Cause:** `GetByID` scans `oauth_provider` and `oauth_id` into `string` fields (not `*string` or using COALESCE). Non-OAuth users have NULL in these columns.

**Reproduction:**
```bash
# After manual migration 017 apply + BUG-02 workaround
curl -b valid-session.txt http://localhost:6767/api/auth/me
# → 500 {"error":"internal_error"}
# Server log: "sql: Scan error on column index 5, name \"oauth_id\": converting NULL to string is unsupported"
```

**Fix:** Either COALESCE in SQL: `COALESCE(oauth_provider, '')`, or change User struct fields to `*string`.

---

### BUG-04: Document IDOR — No Ownership Verification [CRITICAL]
**File:** `internal/handlers/documents.go` (all document-scoped endpoints)
**Severity:** CRITICAL — any authenticated user can read/modify/delete any other user's documents
**Symptom:** User B can access, annotate, and DELETE User A's documents by calling endpoints with User A's document ID.

**Reproduction:**
```bash
# User A creates doc via paste flow
# User B signs up, then:
curl -b userB-cookies.txt "http://localhost:6767/api/documents/<userA-docId>"
# → 200 OK with full chapter data

curl -b userB-cookies.txt -X POST "http://localhost:6767/api/documents/<userA-docId>/annotations" \
  -d '{"target_type":"paper","target_id":"<userA-docId>","content":"hacked"}'
# → 201 Created

curl -b userB-cookies.txt -X DELETE "http://localhost:6767/api/documents/<userA-docId>"
# → 200 {"status":"deleted"}
```

**Affected Endpoints:**
- `GET /api/documents/:id` — full content exposed
- `GET /api/documents/:id/claims` — claims exposed
- `GET /api/documents/:id/research-map` — accessible
- `POST /api/documents/:id/annotations` — can create annotations on other users' docs
- `DELETE /api/documents/:id` — can delete other users' documents

**Impact:** Complete data breach. Any authenticated user can read, modify, or destroy any other user's work.

**Fix:** Every document-scoped handler must verify `doc.UserID == session.UserID` before proceeding. Apply the same pattern used for collections (AGENTS.md D3).

---

### BUG-05: MCP Server Panics on WIP Schema Tags [MEDIUM]
**File:** `internal/mcp/schemas.go` (untracked, WIP)
**Severity:** MEDIUM — blocks MCP development only (not shipped)
**Symptom:** MCP binary crashes at startup with panic: `jsonschema:"description=..."` tags have spaces after `=` which the go-sdk v1.7.0 rejects.

**Root Cause:** `internal/mcp/schemas.go` (untracked WIP) uses `jsonschema:"description=Paper text to ingest..."` — go-sdk parser interprets `description=Paper` as a tag key starting with uppercase, which is invalid.

**Reproduction:**
```bash
go build ./cmd/mcp && ./mcp
# panic: AddTool: tool "ingest_document": input schema: ForType: tag must not begin with 'WORD='
```

**Impact:** MCP binary cannot start from uncommitted code. HEAD (committed) version does not have this file and works.

**Fix:** Remove `jsonschema:"description=..."` tags entirely (they're optional), or use proper jsonschema format.

---

### BUG-06: MCP ingest_document Never Processes Documents [HIGH]
**File:** `internal/mcp/tools.go:59-96`
**Severity:** HIGH — MCP ingest is ingest-only, no pipeline
**Symptom:** `ingest_document` inserts a processing row but never starts the pipeline. Documents stay "processing" forever. `get_document` returns status=processing indefinitely. Evidence/figures always empty.

**Root Cause:** `handleIngestDocument` calls `ValidateAndInsert` (inserts processing row) but never calls `RunPipelineAndPersist`. The tool description says "No LLM, no summarization" — but that means the document is never processed at all.

**Reproduction:**
```bash
# MCP ingest
echo '{"method":"tools/call","params":{"name":"ingest_document","arguments":{"text":"Attention paper text..."}}}' | ./mcp
# → {"document_id":"xxx","status":"processing"}

# Poll get_document forever
# → always {"status":"processing"}
# DB: status stays "processing", chapters=0, charts=0, evidence=0
```

**Impact:** MCP workflow is broken. Agent calls ingest → gets doc_id → polls get_document → stuck forever at "processing". The product promise "polling via get_document" is false for real use.

**Fix:** Either add pipeline launch to the MCP handler, or document this as design (agents must trigger processing via a separate mechanism).

---

### BUG-07: PDF Image Extraction Broken [MEDIUM]
**File:** `internal/external/` (pdfcpu interaction)
**Severity:** MEDIUM — PDF uploads complete but charts are lost
**Symptom:** PDF upload pipeline logs error: `extract images: strconv.Atoi: parsing "all": invalid syntax`. Pipeline completes without image charts.

**Root Cause:** pdfcpu's image extraction returns a page string "all" which `strconv.Atoi` cannot parse. This is a version compatibility issue between pdfcpu and the extraction code.

**Server Log:**
```json
{"error":"extract images: strconv.Atoi: parsing \"all\": invalid syntax","message":"chart image extraction failed","severity":"error"}
```

**Impact:** PDF uploads lose all embedded chart images. The "image fallback" chart path never executes.

**Fix:** Handle "all" page value in image extraction, or pass specific page numbers instead.

---

### BUG-08: Evidence Extraction Yields Zero Rows for Pasted Text [MEDIUM]
**File:** `internal/services/charts_llm.go:207-286`, `internal/services/dataset_build.go:19-24`
**Severity:** MEDIUM — pasted-text documents get no evidence, no charts
**Symptom:** `ExtractNumericEvidence` works in isolation (returns items), but `BuildCandidateDatasets` skips all items because their `Entity` field is empty (BLEU score sentences: "achieved BLEU score of 28.4" → entity="", not "the Transformer"). Log: "no evidence extracted" × 7 chapters.

**Root Cause chain:**
1. `extractEntityValue` regex requires `achieved|obtained|reported|had|reached|scored` verb + entity + value. "BLEU score of 28.4" matches standalone but doesn't capture entity.
2. `groupEvidence` in `dataset_build.go:22` skips items where `entity == ""`.
3. Result: zero datasets → zero charts → zero evidence rows for pasted-text docs.

**Evidence:** DB shows `charts=0, evidence=0` for pasted doc despite chapter excerpts containing "achieved a BLEU score of 28.4".

**Impact:** Pasted-text documents get no charts and no evidence data. The evidence verification still runs (via Gemini) but the structured evidence table is empty.

**Fix:** Allow entity-less evidence items to form "aggregate" datasets, or improve entity extraction for metric+value patterns.

---

### BUG-09: Chapter Index Misalignment in Chart Generation [LOW]
**File:** `internal/services/charts_llm.go:274-281`
**Severity:** LOW — cosmetic, charts link to wrong chapters
**Symptom:** `Chart.ChapterIndex = nextOrder` uses running chart counter (displayOrder parameter) instead of the actual chapter index. Charts may link to wrong chapters or none.

**Evidence:** Line 279: `ChapterIndex: nextOrder` where `nextOrder` is the chart display counter, not the chapter's index in the chapters array.

**Impact:** Charts display under wrong chapter headings.

**Fix:** Pass the actual chapter index to `GenerateChapterCharts` instead of using the running chart counter.

---

## Security Audit

### Security Headers ✅ GOOD
| Header | Value | Status |
|--------|-------|--------|
| Content-Security-Policy | `default-src 'self'; font-src 'self' https://fonts.googleapis.com...` | ✅ |
| X-Content-Type-Options | `nosniff` | ✅ |
| X-Frame-Options | `DENY` | ✅ |
| X-XSS-Protection | `1; mode=block` | ✅ |
| Cache-Control | `no-store` (on API) | ✅ |

### Cookie Flags ✅ GOOD
All session cookies use: `HttpOnly=true, Secure=true, SameSite=Lax`

### CSRF Protection ⚠️ PARTIAL
- OAuth state parameter: random nonce in httpOnly cookie (fixed from hardcoded string)
- But no CSRF token on state-changing POST endpoints

### Rate Limiting ✅ GOOD
- `POST /api/documents`: 1 req/30s, burst 2 (IP-based)
- `POST /api/auth/signup`: 5 req/60s, burst 3
- `POST /api/auth/login`: 5 req/60s, burst 3
- MCP tools: per-key rate limits (analyze: 5/min, read: 30/min)

### SQL Injection ✅ RESISTANT
Parameterized queries throughout. `chi` router + `database/sql` with `?` placeholders.

### XSS ✅ GOOD
React auto-escapes by default. CSP `script-src 'self'` blocks inline scripts.

### CRITICAL: IDOR (see BUG-04)
Any authenticated user can access any document by ID. No ownership verification on document-scoped endpoints.

---

## MCP Audit (Persona B: Code Agent)

### Server Status
- **HEAD (committed):** Builds and runs, 5 tools registered
- **WIP (uncommitted):** Panics at startup (BUG-05)
- **Tested with:** Patched build (removed jsonschema tags for audit)

### Tool Verification
| Tool | Status | Notes |
|------|--------|-------|
| `ingest_document` | ✅ Deterministic | Inserts processing row, returns doc_id |
| `search_documents` | ✅ Works | Returns matching titles |
| `get_document` | ⚠️ Partial | Metadata only; sections/evidence/figures depend on pipeline (which never runs — BUG-06) |
| `get_figures` | ✅ Works | Returns empty (no charts generated) |
| `get_evidence` | ✅ Works | Returns empty (no evidence rows) |

### Rate Limiting
- `ingest_document`: 5/min per API key → correctly enforced (tests 5/7 = RATELIMITED on 5th+)
- `search_documents`/`get_document`: 30/min per API key

### Error Handling
- Invalid document_id → `{"error":"document not found"}`
- Empty text → `{"error":"text is required"}`
- Invalid reading_level → `{"error":"invalid reading_level..."}`
- Size limit (500KB) → `{"error":"size_limit_exceeded"}`

### Agent Workflow Test
```
Agent calls ingest_document("Attention paper text...") →
  Returns: {"document_id":"xxx","status":"processing","title":"..."}

Agent calls get_document("xxx") →
  Returns: {"status":"processing"} forever (pipeline never starts)

Agent calls search_documents("attention") →
  Returns: {"documents":[]} (title doesn't match "attention")

Agent calls get_figures("xxx") →
  Returns: {"figures":[]} (empty)

Agent calls get_evidence("xxx") →
  Returns: {"evidence":[],"claims":[]} (empty)
```

**Verdict:** MCP surface is non-functional for real agent workflows. Ingest works but processing never happens.

---

## Frontend Audit

### Route Coverage
| Route | Status | Notes |
|-------|--------|-------|
| `/` (Landing) | ✅ Loads | Paste tab works, upload tab works |
| `/login` | ✅ Loads | Auth broken by BUG-01/02/03 |
| `/signup` | ✅ Loads | Signup 201 OK but user fetch broken |
| `/account` | ⚠️ Loads | Redirects to /login (auth broken) |
| `/agents` | ⚠️ Loads | Shows MCP config, "Sign in to get API key" |
| `/upload` | → `/` | Redirect (correct) |
| `/dashboard` | → `/account` | Redirect (correct) |
| `/404` | ✅ Loads | Not-found page renders |

### Console Errors
1. **CSP font violation** (every page): `fontshare.com` font blocked by CSP `font-src` directive — `api.fontshare.com` allowed but `cdn.fontshare.com` not listed
2. **401 on /api/auth/me** (expected for unauthed users)
3. **500 on /api/auth/me** (BUG-03 — NULL scan)

### Bundle Size
- HTML: ~1KB
- JS: Served from `assets/index-BMpL5jRW.js` (Vite build)
- CSS: Served from `assets/index-DVWWgPI6.css`

### Accessibility
- All form inputs have labels
- Tab roles present on upload component
- Button text descriptive

---

## Performance

| Metric | Value |
|--------|-------|
| Health check response | 1.2ms |
| Landing page TTFB | ~3ms |
| Pipeline (paste, ~40KB) | 13-19s (Gemini calls) |
| Gemini API calls per doc | 4-7 calls (simplify + verify + chapters + chart plans) |
| Pipeline stages | simplify → verify → chapters → charts → image fallback |

---

## Recommended Fix Priority

| Priority | Bug | Effort |
|----------|-----|--------|
| **P0** | BUG-04: Document IDOR | Medium — add ownership check to all document handlers |
| **P0** | BUG-01: Session self-deletion | Low — swap DeleteByUserID order in Login handler |
| **P1** | BUG-02: Migration 017 not registered | Low — add to loadMigrations maps |
| **P1** | BUG-03: NULL scan on OAuth columns | Low — COALESCE in SQL or `*string` in struct |
| **P1** | BUG-06: MCP never processes | Medium — add pipeline launch or document design |
| **P2** | BUG-05: MCP WIP schema panic | Low — remove jsonschema tags |
| **P2** | BUG-07: PDF image extraction | Medium — fix strconv.Atoi for "all" |
| **P2** | BUG-08: Evidence extraction zero | Medium — fix entity extraction or allow entity-less |
| **P3** | BUG-09: Chapter index misalignment | Low — pass chapter index instead of chart counter |

---

*Report generated by Sisyphus E2E audit, 2026-09-17.*
