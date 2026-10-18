# Master Refactor Audit — Wave H (P50–P61)

> Plan: `.omo/plans/paperviz_master_refactor_plan.md`. Branch `refactor/master`,
> waves A–G landed through `8879d63`; docs pass `1377f8e`. This file is the
> evidence record for the plan's final audit gate. One file on purpose — the
> plan asks for an ownership table and per-phase audit results, and the repo
> policy is "no unnecessary docs."

## P50 — UX consistency: PASS

Docs pass (`1377f8e`) synced all prose to shipped behavior. Frontend edits across
waves reused existing DESIGN.md classes/tokens only (no invented values), ARIA
tab patterns preserved, terminology (Understanding/Evidence/Figures/Source)
consistent across `components/result/*`.

## P51 — Secondary dominance: PASS

Nav evidence (grep of `Link`/`Navigate` across `frontend/src`):
- Single canonical entry: `/upload` → `/` redirect (`App.jsx:44`); `/dashboard` → `/account` (`App.jsx:45`).
- No global nav bar; landing page carries the 2 primary CTAs (Add to Claude Code / Sign in).
- `/account`, `/agents` reachable only from small header/footer links — secondary weight, primary flow `Input → Result` untouched.

## P52 — Test architecture: PASS

Tests organized by business boundary (11 packages): `internal/handlers` (HTTP contract + ownership), `internal/services` (pipeline/evidence/charts/grounding), `internal/repository`, `internal/mcp` (contract tests per tool), `internal/app/*` (documents/credentials/users/analytics), `internal/testutil` (shared seeds). No test moves required.

## P53 — Duplication audit: PASS

- HTTP transport = single funnel: `lib/auth.js` `authFetch` wrapper + `lib/api.js` (grep `\bfetch\(` outside api.js: only the auth wrapper itself and one analytics beacon).
- Result page state = single source (`hooks/useResultState.js`); no per-page poll/error copies.
- Add-to-collection UI exists once (`result-actions.jsx`) after dead `collections-panel.jsx` deletion.

## P54 — File-size gate (>250 LOC): PASS (all justified)

Rule: split real god-files (multiple concerns), not arbitrary 200-LOC splits.

| File | LOC | Verdict |
|---|---|---|
| `internal/mcp/tools.go` | 547 | single concern (5 tool handlers + schemas); already the P26–P34 target, split would fragment one contract |
| `repository/documents.go` | 347 | persistence-only CRUD for the doc aggregate |
| `services/import.go` | 344 | one domain (DOI/URL fetch + SSRF rules) |
| `external/llm.go` | 337 | provider adapter (goai wrapper) — infra, exempt from layering split |
| `services/share.go` | 317 | one domain (share/visibility/expiry) |
| `handlers/collections.go` | 315 | transport only, thin per handler |
| `services/intake.go` | 305 | one domain (ingestion + pipeline entry) |
| `services/charts_llm.go` | 286 | LLM plan mechanics only (P15 separation already applied) |
| `services/comparison.go` | 283 | one domain (compare, in-app only) |
| `services/grounding.go` | 277 | one domain (10 deterministic rules) |
| `services/evidence_extract.go` | 260 | one domain (regex extraction) |
| `frontend/src/lib/api.js` | 346 | single API surface (deliberate, see P53) |
| `frontend/src/components/chart-card.jsx` | 275 | single component with provenance/grounding display |

No file violates "multiple unrelated concerns." Nothing further to split.

## P55 — Dependency direction: PASS

`graphify path` results: `mcp → handlers` NONE · `mcp → gemini/LLM` NONE ·
`handlers → repository` NONE (confirmed by import grep: zero `paperviz/internal/repository` imports in `internal/handlers`; prose comments only).
Enforced shape: `UI → API → handlers → app/services → repo/infra`; `User Model → MCP → data layer`.

## P56 — Business-logic ownership table (one owner per rule)

| Business rule | Owner |
|---|---|
| Document create/tx boundary, read-model aggregation | `internal/app/documents` (`service.go`, `readmodel.go`) |
| Source-type policy (pdf/pasted_text/doi/url capabilities) | `internal/services/source_policy.go` |
| Pipeline stage orchestration + user-facing stages | `internal/services/intake.go` (`OnStage`), labels in `hooks/useResultState.js` |
| Simplify / claim-diff verification | `internal/services/simplify.go`, `verification.go` |
| Evidence extraction (numeric patterns) | `internal/services/evidence_extract.go` |
| Dataset building | `internal/services/dataset_build.go` |
| Chart validity + grounding (10 rules → DO NOT RENDER) | `internal/services/grounding.go` (`ValidateGrounding`) |
| Chart failure taxonomy | `internal/services/charts.go` (`ChartFailureCategory`) |
| Provider policy (retries, rate limits, timeouts) | goai transport + `internal/infra/ratelimit`; HTTP auth rate limit `internal/handlers/ratelimit.go` |
| Error model (apperr kinds → HTTP mapping) | `internal/apperr/errors.go` + `writeDocumentAccessError` |
| Document access (owner/anonymous rules) | `internal/services` `AuthorizeDocumentAccess` + `handlers.RequireDocAccess` (regression test `TestSectionRoutesRespectOwnershipOverRouter`) |
| Collection / annotation ownership | `internal/services/collections.go`, `annotations.go` |
| Share visibility, 7-day expiry, funnel counters | `internal/services/share.go` |
| Identity | Clerk (`users.id`); profile/API-key digests `internal/app/users`, `internal/app/credentials` |
| MCP surface (exactly 5 tools, no reasoning) | `internal/mcp/tools.go` + contract tests + `scripts/mcp-smoke.sh` |

## P57 — E2E flow: PASS

Go-CI runs the `e2e` job green on every push (e.g. run `39144057154` on `8879d63`, success 3m23s), covering Input→ingestion→result→navigation specs.

## P58 — UX smoke: covered by e2e; manual browser pass deferred

Automated nav/result/share specs run in CI. A hands-on visual pass with screenshots is left to the owner (can be run via `agent-browser` on demand) — flagged, not blocking.

## P59 — MCP smoke: PASS

- `scripts/mcp-smoke.sh` (new): builds `cmd/mcp`, runs real stdio JSON-RPC with a **fake `GEMINI_API_KEY`** (proves LLM-independence), exercises all 5 tools: `ingest_document` → `search_documents` → `get_document`/`get_evidence`/`get_figures` (not-found paths return structured errors). Result: `MCP smoke OK: 5 tools exercised over stdio`.
- `internal/mcp/mcp_contract_test.go` (397 LOC) covers every tool at handler level: handshake/list exact-5, valid/empty/invalid/not-found per tool.
- Tool count: `grep -c 'Name:' internal/mcp/tools.go` = 5.

## P60 — Architecture regression: PASS

- God-files: only pre-existing single-concern files >250 LOC (P54 table); no new ones created by waves.
- Handler→repo coupling: zero (P55).
- Hidden LLM deps in MCP: none (P28 removed dead gemini field; P59 fake-key boot proves it).
- Unneeded infra: none added (no ORM/queue/broker/microservice; monolith preserved).
- Scope creep: dead code deleted instead of carried (`collections-panel.jsx`, `useCollections.js`, `INSTALLATION_AND_SETUP.md`).

## P61 — Final simplification: PASS

- `lib/mappers/` layer (P43) deliberately NOT created — API/DTO payloads already render directly in JSX (grep found only two legitimate `JSON.parse` sites: chart_data + local state); adding a transform layer would be code nobody asked for.
- Refactor additions all justified: app services, ownership guard + regression test, state machine, smoke script, testutil seeds.
- Gates (final): `go build` 3 binaries · `go test ./...` all pkgs · `go vet` clean · `gofmt -l .` empty · MCP handshake 5 tools · `npm run lint` clean · frontend build clean.
