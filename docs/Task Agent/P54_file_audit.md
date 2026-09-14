# P50-P61 Final Audit — PaperViz Master Refactor

> Date: 2026-09-15
> Baseline: 445 tests passed, `go vet` clean, `go build ./cmd/mcp` clean, `go build ./cmd/server` clean

---

## P54: File Size Audit

### GO Files (non-test, sorted by LOC)

| LOC | File | Status |
|-----|------|--------|
| 1008 | `internal/handlers/documents.go` | GOD-FILE — needs split (P08 planned) |
| 547 | `internal/mcp/tools.go` | HIGH — 5 tools, split by tool would fragment |
| 431 | `internal/handlers/auth.go` | HIGH — 6 endpoints + OAuth + session |
| 399 | `internal/external/gemini.go` | ACCEPTABLE — single Gemini adapter, pure infra |
| 344 | `internal/services/import.go` | HIGH — DOI/URL/fetcher + SSRF guard |
| 323 | `internal/repository/documents.go` | HIGH — many query methods for doc aggregate |
| 317 | `internal/services/share.go` | HIGH — share + referral + public view logic |
| 315 | `internal/handlers/collections.go` | HIGH — 7 CRUD endpoints |
| 286 | `internal/services/charts_llm.go` | HIGH — LLM plan + chart generation |
| 283 | `internal/services/comparison.go` | HIGH — comparison + ponytail ceilings |
| 277 | `internal/services/grounding.go` | HIGH — 10 validation rules |
| 264 | `internal/services/intake.go` | HIGH — pipeline orchestration |
| 260 | `internal/services/evidence_extract.go` | HIGH — 5 regex patterns |
| <250 | All other .go files | OK |

### JSX Files (sorted by LOC)

| LOC | File | Status |
|-----|------|--------|
| 419 | `frontend/src/components/annotation-panel.jsx` | HIGH — CRUD + error handling + UI |
| 397 | `frontend/src/components/collections-panel.jsx` | HIGH — CRUD + error handling + UI |
| 300 | `frontend/src/pages/result-page.jsx` | HIGH — but was 743, now 297 (split done) |
| 257 | `frontend/src/components/chart-card.jsx` | BORDERLINE — provenance + grounding + share |
| 234 | `frontend/src/components/data-chart.jsx` | OK — Recharts wrapper + grounding badge |
| 232 | `frontend/src/components/result/understanding-section.jsx` | OK — simplified text + controls |
| 231 | `frontend/src/pages/share-paper-page.jsx` | OK — public share view |
| <200 | All other .jsx files | OK |

### Files Over 250 LOC Summary

- **GO:** 13 files exceed 250 LOC. Top offender `documents.go` at 1008 LOC (god-file, P08 split target).
- **JSX:** 4 files exceed 250 LOC. `annotation-panel.jsx` and `collections-panel.jsx` are largest.
- **Action required:** P08 split of `documents.go` is the critical path. Others are justified by single-responsibility (auth has 6 endpoints, gemini is one adapter).

---

## P50: UX Consistency Audit

### Terminology Check
- ✅ "Understand" used consistently for primary action (not "analyze", "summarize", "explain")
- ✅ "Evidence" used for claim data (not "proof", "verification")
- ✅ "Figures" used for charts (not "charts", "graphs", "visualizations")
- ✅ "Source" used for original text section
- ✅ Stage labels: reading doc → extracting structure → preparing evidence → rebuilding figures → completing

### Nav Patterns
- ✅ Primary nav: Logo + Upload (primary) + Sign In (secondary)
- ✅ Account nav: API Key + Usage + Subscription
- ✅ No orphaned nav items, no dead routes
- ✅ `/dashboard` → `/account` redirect consistent

### Button Behavior
- ✅ Primary CTA: `bg-black text-white rounded-full` (Filled Dark CTA from DESIGN.md)
- ✅ Secondary: outlined style with `#e5e5e5` border
- ✅ Ghost buttons for nav items
- ✅ All buttons have loading/disabled states

### Hierarchy
- ✅ Satoshi for display headings only (36px+)
- ✅ Inter for body, labels, UI (all sizes ≤30px)
- ✅ Electric Blue (#2563eb) for links/active states only
- ✅ Deep Sapphire (#1e40af) for primary action fills only
- ✅ 1px #e5e5e5 borders as primary container definition

### Loading Patterns
- ✅ Skeleton loaders for result sections
- ✅ Processing stages with 5-step progress
- ✅ Inline error + Retry pattern consistent (12.1 standard)
- ✅ No spinners without context

### DESIGN.md Token Compliance
- ✅ All colors from Dub palette, no invented colors
- ✅ Border radius follows vocabulary (9999px, 16px, 12px, 8px, 6px)
- ✅ Spacing follows 4px base unit
- ✅ Shadows minimal, borders preferred

---

## P55: Dependency Direction Audit

### Required Direction
```
UI → API → App → Domain → Repo/Infra
User Model → MCP → App/Data layer
```

### Audit Results
- ✅ **UI → API:** Frontend calls `/api/*` endpoints, no direct repo access
- ✅ **API → App:** Handlers call `app/documents` service (P07)
- ✅ **App → Domain:** Services call domain functions (charts, evidence, grounding)
- ✅ **Domain → Repo/Infra:** Services call repository for persistence, external for LLM
- ✅ **MCP → App/Data:** `internal/mcp/tools.go` imports `repository` (data access), no handler imports
- ⚠️ **Handler → Repo coupling:** 36 instances in `documents.go` (New.*Repo calls). Also in `auth.go` (11), `analytics.go` (1), `middleware.go` (2). P09 target.
- ⚠️ **Services → Repo coupling:** `annotations.go` creates `NewAnnotationRepo` inline (5 instances). Should accept repo via DI.

### Reverse Dependency Check
- ✅ No `internal/handlers` imports in `internal/mcp/`
- ✅ No `internal/mcp` imports in `internal/handlers/`
- ✅ No `internal/mcp` imports in `internal/app/`
- ✅ MCP depends on `repository` only (data layer) — correct direction

---

## P56: Business Logic Ownership Audit

| Business Rule | Owner | Location | Notes |
|---------------|-------|----------|-------|
| Ingestion (PDF/DOI/URL/paste) | `services/import.go` + `services/intake.go` | Pipeline entry | SSRF guard in import.go |
| Source-type policy | `services/source_policy.go` | Policy table | pdf→text+image, paste→text-only |
| Chart validation | `services/grounding.go` | 10 deterministic rules | No LLM override |
| Evidence grounding | `services/grounding.go` | Same owner as chart validation | Unified grounding service |
| Evidence extraction | `services/evidence_extract.go` | 5 regex patterns | Deterministic |
| Dataset building | `services/dataset_build.go` | Group by metric+unit | Pure transform |
| Chart LLM planning | `services/charts_llm.go` | Gemini plan call | LLM interprets, never invents |
| Simplification | `services/simplification.go` | Gemini call | Text rewrite only |
| Claim-diff verification | `services/verification.go` | DiffClaims | 2-call Gemini |
| Search | `services/documents.go` | FTS/ILIKE | Document title+chapters |
| MCP behavior | `mcp/tools.go` + `mcp/server.go` | 5 tools, stdio | No reasoning, data only |
| Provider policy | `services/source_policy.go` | Source type rules | Guard in service, not handler |
| Collection ownership | `services/collections.go` | Per-user enforcement | ErrForbidden on mismatch |
| Annotation ownership | `services/annotations.go` | Per-user enforcement | ErrForbidden on mismatch |
| Share/referral | `services/share.go` | Token + counters | 7-day expiry |

### One-Owner Verification
- ✅ Each rule has single owner file
- ⚠️ Grounding owns both chart validation AND evidence grounding — acceptable, same domain
- ⚠️ `annotations.go` creates repo inline instead of accepting via DI — minor ownership smell

---

## P60: Architecture Regression Check

### God-Files (check for new ones)
- `documents.go` 1008 LOC — **EXISTING** god-file, P08 target. Not new.
- `tools.go` 547 LOC — **EXISTING**, justified by 5 tool implementations
- No NEW god-files introduced during refactor

### Handler → Repo Coupling
- 36 instances in `documents.go` — **EXISTING** (P09 target, not yet executed)
- 14 instances in `auth.go` — **EXISTING** (session/user repo creation)
- 1 instance in `analytics.go` — **EXISTING**
- No NEW coupling introduced

### Duplicate Rules
- ✅ No duplicated business logic detected
- ✅ Each domain rule has single owner (see P56)
- ✅ Evidence extraction: one location (`evidence_extract.go`)
- ✅ Grounding: one location (`grounding.go`)

### Hidden LLM Dependencies
- ✅ `mcp/tools.go` has NO Gemini imports (data only)
- ✅ `mcp/server.go` holds GeminiClient but only passes to tools for data retrieval
- ✅ No hidden Gemini calls in pipeline or ingestion

### Accidental MCP Reasoning
- ✅ MCP tools return structured data only
- ✅ No summarization/interpretation in MCP path
- ✅ `ingest_document` is LLM-independent

### Unneeded Abstraction
- ✅ No premature interfaces in handler layer
- ✅ No wrapper types adding no value
- ✅ Repository types are concrete, not over-abstracted

### Scope Creep
- ✅ No new features added during refactor
- ✅ All changes within P01-P61 plan scope
- ✅ No new dependencies introduced

---

## P61: Final Simplification

### What Was Added During Refactor
1. `internal/app/documents/{service.go,readmodel.go}` — **NEEDED** (P07 app layer)
2. `internal/handlers/documents_create.go` — **NEEDED** (P08 split)
3. `frontend/src/components/result/*.jsx` — **NEEDED** (P04 result decomposition)
4. `docs/Task Agent/P01_inventory.md` — **NEEDED** (audit evidence)

### What Could Be Stripped
- None identified. All additions serve clear architectural purpose.

### Cleverness vs Clarity
- ✅ Read-model aggregation is straightforward (no hidden magic)
- ✅ Service interface is simple (3 methods: Create, Get, List)
- ✅ No dynamic dispatch, no reflection, no code generation

### Recommendation
- **Keep all additions.** No unnecessary code detected.
- **Next priority:** Execute P09 (kill handler→repo coupling) and P10 (read-model aggregation) to complete the god-file decomposition.

---

## Summary

| Audit | Status | Action |
|-------|--------|--------|
| P50 UX Consistency | ✅ PASS | No changes needed |
| P54 File Size | ⚠️ 13 GO + 4 JSX >250 LOC | P08 split documents.go (critical) |
| P55 Dep Direction | ⚠️ 36 handler→repo couplings | P09 kill coupling |
| P56 Ownership | ✅ PASS | One owner per rule verified |
| P60 Arch Regression | ✅ PASS | No new violations |
| P61 Simplification | ✅ PASS | All additions justified |

**Next action:** Execute P09 (handler→repo decoupling) + P10 (read-model aggregation) to resolve the two ⚠️ items.
