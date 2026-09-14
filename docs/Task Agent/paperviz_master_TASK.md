# PaperViz Refactor — Compressed Master Task

MISSION: refactor end-to-end, no rewrite. Stay monolith. Fix UX/nav/frontend/backend/pipeline/charts/MCP/code-org.

RULES (apply every phase):
- audit before edit
- preserve behavior unless phase says change it
- no microservices/queues/event-bus/speculative infra
- simple explicit code, ~200 LOC/file max, decompose god-files
- kill dead code/dup logic/noise comments
- business logic lives in app/domain services, NOT handlers/routers/UI
- MCP = data/tools only, NEVER a reasoning layer, NEVER silently calls app's Gemini/LLM
- model behind MCP = user's own model, not ours
- validate before next phase, don't stop at surface fixes, no scope creep

DEPS: UI→API→App→Domain/Services→Repo/Infra. User Model→MCP→App/Data layer. No reverse deps. No handler owns business logic. No MCP reasoning.

MCP FINAL SURFACE (exactly 5 tools, no more):
ingest_document, search_documents, get_document, get_evidence, get_figures
(delete: analyze_paper, get_summary, get_claims, compare_papers — compare logic stays in-app only, not MCP)

---

## PHASES (execute top-down, no manual dep coordination)

P01 Inventory: map whole repo (frontend/backend/handlers/services/repos/adapters/pipeline/charts/evidence/mcp/routing/tests/god-files/dup logic/dep violations).

P02 Define canonical flow: Input→processing→understanding→evidence→figures→source→mgmt. Rank screens primary/secondary. No redesign yet.

P03 Fix routing/IA: dead routes, bad redirects, dup entry points, unneeded top pages. Make primary journey obvious.

P04 Rebuild primary UX: Upload/Paste/Import→Processing→Result. Result page must show: understanding, evidence, source, mgmt.

P05 Clean ingestion UX: unify PDF/paste/other input, make source type explicit, cut complexity.

P06 Processing UX states: reading doc / extracting structure / preparing evidence / rebuilding figures / completing. Hide impl details.

### Backend/Domain
P07 Doc domain boundary: app services own logic (create/get/list/metadata/sections/evidence/figures/read-models). Handlers = transport only.

P08 Split documents.go: no handler owns repo orchestration/aggregation/business decisions/big transforms.

P09 Kill handler→repo direct coupling everywhere app service should own it.

P10 Clean doc read-models: aggregation (doc+chapters+charts+evidence+claims) moves out of handlers into read-model layer.

P11 One canonical pipeline: clear stages — ingestion/extraction/transformation/verification/figures/persistence. No dup parsing.

P12 Split pipeline.go: orchestration coordinates, doesn't hold business rules.

P13 Source-type rules: explicit PDF vs paste vs future-source behavior. No PDF-only ops leaking to non-PDF.

P14 Remove time.Sleep/arbitrary delays from business logic → move rate-limit/retry to infra/provider layer.

### Charts/Evidence
P15 Split charts.go: separate rules/extraction/parsing/annotation/provenance/persistence/LLM-mechanics. No god-service.

P16 Chart policy vs LLM mechanics: domain defines valid-chart/type/provenance/grounding; LLM call = infra concern.

P17 Chart grounding states: verified/partial/unsupported/failed. Never imply stronger evidence than source has.

P18 Split types.go: no mixed-concern dumping ground (extraction/charts/verification/comparison/evidence/pipeline types separated).

P19 Repo layer cleanup: kill dup queries, business logic, over-abstraction, unclear tx boundaries. Repos = persistence only.

P20 Doc aggregate consistency: audit doc/sections/evidence/claims/figures/tables ownership+lifecycle.

P21 Error model: separate validation/not-found/permission/processing/provider/internal errors. No infra errors leaking to UX.

P22 Config/policy cleanup: move provider config/limits/flags/timeout-retry out of random services into policy layer.

P23 Split router.go: router = route registration + middleware + DI wiring only. Not a service.

P24 External adapter cleanup (esp. Gemini): clean interfaces, no provider leakage into domain.

### MCP
P25 MCP boundary invariant: User Model→MCP→Data/Tools (NOT →app Gemini→result). Reasoning stays with user's model.

P26 Lock MCP surface to exactly 5 tools (see above). No expansion w/o strong reason.

P27 Replace analyze_paper→ingest_document. Allowed: receive/parse/extract structure+tables+figures/id evidence candidates/persist/return id+status. FORBIDDEN: summarization/reasoning/synthesis/hidden Gemini/interpretation.
  Output: {document_id, title, status, metadata}

P28 Verify MCP ingest path is LLM-independent end to end — trace full call path, split deterministic extraction from any reasoning call.

P29 Add search_documents. In: {query, limit}. Out: {results:[{document_id,title,relevance,matched_sections:[{chapter_id,title}]}]}

P30 Replace get_summary→get_document. Support selective retrieval via include:["metadata","sections",...]. No pre-generated interpretation.

P31 Merge get_claims into get_evidence. Returns claims/source text/page/section/figure+table refs/provenance. PaperViz supplies evidence only; user model judges support/contradiction/sufficiency.

P32 Keep+refactor get_figures: structured data only (figure id/type/chart data/source page+text/chapter/provenance/grounding). No AI-generated explanations.

P33 Delete from MCP: analyze_paper, get_summary, get_claims, compare_papers. (compare_papers: app-level Compare feature can stay, just not exposed via MCP.)

P34 Define explicit output schemas per MCP tool: deterministic/machine-readable/minimal/composable/stable. Avoid giant payloads — support selective retrieval.

P35 Verify MCP dep direction: MCP→App/Data→Repo/Infra only. FORBIDDEN: MCP→Handler→Repo, MCP→Gemini→result.

P36 MCP contract tests per tool: valid/invalid input, not-found, empty result, malformed doc, deterministic output, schema correctness, confirm no prohibited LLM calls triggered.

### Frontend
P37 Split result-page.jsx into: header/understanding/evidence/figures/source/actions components.

P38 Result state machine: loading/partial/ready/error/retryable-error/empty. No dup state logic across components.

P39 Chart UX: show what was extracted, source, grounding status, source location, failure states. Never imply auto-trust.

P40 Dashboard refactor (stays secondary): recent docs/search/collections/resume-work. Not the product center.

P41 Split oversized annotation/collection components; business logic→hooks/services, keep presentation thin.

P42 Centralize fetch/loading/error/retry/transform logic in hooks — no dup API semantics per page.

P43 DTO/transform cleanup: keep transport DTOs separate from presentation models, no deep ad-hoc transforms in JSX.

### Integration
P44 Auth UX+arch alignment across routes/API calls/redirects/protected resources/sharing.

P45 Dashboard/Collection/Compare flow repair — secondary features must not corrupt primary flow. Compare stays app-only, not MCP.

P46 Share flow audit: creation/public view/permissions/invalid links/expired-deleted resources.

P47 Empty/loading/error/success audit across every major screen — no undefined states.

P48 Recovery UX for: failed ingestion/processing, missing source data, failed chart extraction, network errors.

P49 Accessibility/interaction audit: keyboard, buttons, focus, labels, dialogs, responsive, destructive actions.

### Final
P50 UX consistency pass: terminology/nav patterns/button behavior/hierarchy/loading patterns.

P51 Verify secondary features don't dominate — product = "understand papers with evidence," not "giant research workspace."

P52 Test architecture repair: organize by business boundary not implementation accident. Priority: doc flow/pipeline/evidence/charts/MCP/critical UX-API.

P53 Duplication audit (transforms/validation/repo calls/API logic/error handling/business rules) — consolidate only true duplicates.

P54 File size audit: enforce ~200 LOC/file target, decompose real god-files (not arbitrary splitting).

P55 Dependency direction audit (see DEPS above) — no reverse deps, no MCP reasoning, no handler-owned orchestration.

P56 Business logic ownership audit: one clear owner per rule — esp. ingestion, source-type behavior, chart validation, evidence grounding, search, MCP behavior, provider policy.

P57 E2E flow validation: Input→ingestion→processing→doc→evidence→figures→result→mgmt. Also: User Model→MCP→PaperViz data→user reasoning.

P58 UX smoke test as real user: entry/ingest/processing/result/chart/evidence/source/nav/recovery/secondary flows.

P59 MCP smoke test — verify all 5 tools independently: structured output, deterministic behavior, composable retrieval, no hidden reasoning, no prohibited LLM calls.

P60 Architecture regression check: scan for new god-files, new handler→repo coupling, dup rules, hidden LLM deps, accidental MCP reasoning, unneeded abstraction, scope creep.

P61 Final simplification: strip anything added during refactor that isn't necessary. Prefer less code + clear ownership over cleverness.

---

## DONE WHEN
Primary UX clear · nav coherent · backend responsibilities separated · business logic has 1 owner each · god-files decomposed · pipeline consistent · chart/evidence grounding explicit · MCP = exactly 5 tools, no reasoning, no silent Gemini calls · compare reasoning stays with user's model · frontend state cleaner · secondary features contained · tests/smoke pass · no major arch violations · no unneeded infra added.

EXECUTION: strict P01→P61 top-down. Orchestrator owns all dep sequencing — human doesn't pick next task/blocking order/file order. Don't declare done until final audit (P60-61) passes, not just visible fixes.
