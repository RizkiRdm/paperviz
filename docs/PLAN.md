# PLAN.md — PaperViz

Constraint context: solo-dev, ~1 hour/day development time, AI-agent-directed implementation (Claude Code / OpenCode / Codex as execution tools). This plan is written as a strict sequential checklist with exit criteria per phase to counter the documented ~20% abandonment pattern. Do NOT start a phase until the prior phase's exit criteria are met. Do NOT reorder phases.

Anti-abandonment mechanism: each phase has a single-sentence "done means" statement. If a phase drags past its estimated session count without meeting "done means," STOP and reassess scope — do not silently expand the phase.

---

## Phase 0 — Research & Validation (no code)

**Done means**: You have manually run 3-5 real papers through a raw Gemini prompt (outside the app) and confirmed simplification + claim-extraction is qualitatively viable before building any infrastructure around it.

- [x] Pick 5 real papers (mix of fields: at least 1 STEM, 1 social science) as fixed test set.
- [x] Manually prompt Gemini (via AI Studio UI, no code) to simplify one paragraph from each paper at "Simplified" level. Read the output — does it preserve numbers/claims correctly?
- [x] Manually prompt Gemini to extract "list of factual claims" from an original paragraph vs its simplified version. Does the diff approach actually catch a deliberately-introduced error? (Test this — manually corrupt one number in a simplified passage and see if claim-extraction catches it.)
- [x] Manually test chart-data extraction: paste a table from one of the 5 papers, ask Gemini to output structured JSON (labels + values). Check accuracy.
- [x] Decision gate: if claim-diff or chart-data extraction is unreliable in this manual test, STOP and reconsider approach before writing any Go code. This is cheaper to discover now than after Phase 2.

**Est. sessions**: 3-5 (at ~1hr/day, this is under a week)

---

## Phase 1 — Foundation

**Done means**: Empty pipeline runs end-to-end with hardcoded/stub responses (no real Gemini calls yet), server starts, SQLite schema exists.

- [x] Repository setup (Go module, directory structure per ARCHITECTURE.md: `handlers/`, `services/`, `repository/`, `external/`).
- [x] SQLite schema migration (3 tables from ARCHITECTURE.md Section 3). See `migrations/001_init.sql`.
- [x] `chi` router with 2 endpoints (`POST /api/documents`, `GET /api/documents/:id`). Implemented with real logic directly (see Phase 2) rather than hardcoded stubs first — same total work, one less throwaway pass.
- [x] Repository layer: CRUD functions for `documents` table. See `repository/documents.go`.
- [x] Verify: server starts, endpoints respond, data round-trips through SQLite. **Could not execute `go build`/`go test` in the build sandbox** — `modernc.org` and `golang.org` are outside its network egress allowlist, and `modernc.org/sqlite`'s core transitive dependency (`modernc.org/libc`) has no GitHub mirror to route around it. Code was verified by static review + `gofmt` (zero syntax errors across all files) instead. **Run `go build ./...` and `go test ./...` on a normal machine before considering this phase truly verified.**

**Est. sessions**: 5-7

---

## Phase 2 — Core Pipeline: Text Extraction + Simplification

**Done means**: Upload a real PDF, get back real simplified text, no charts yet, no claim-diff yet.

- [x] PDF text extraction wrapper (`external/pdf.go`) — extract raw text, detect "no text layer" case, reject early per ARCHITECTURE.md Failure Scenario 1.
- [x] Gemini API client wrapper (`external/gemini.go`) — single-purpose function, takes prompt + text, returns completion, respects 30s timeout + 1 retry.
- [x] Simplification service (`services/simplification.go`) — prompt engineering for Simplified/ELI5 levels, preserving all facts per PRD.md's correctness requirement.
- [x] Wire up: `POST /api/documents` accepts real upload → extracts → simplifies → saves → returns document_id. See `handlers/documents.go`.
- [x] `GET /api/documents/:id` returns real saved data.
- [x] Validation: file size limit (20MB), MIME type check via magic-byte inspection (`handlers/validation.go`, not just the client's Content-Type header, per AGENTS.md Security Rules).

**Est. sessions**: 8-10

---

## Phase 3 — Claim-Diff Verification

**Done means**: Pipeline flags a deliberately-corrupted simplification as `verification_failed`, and passes a correct one as `complete`.

- [x] Claim-extraction prompt (extract factual claims as structured list from a text passage). See `services/verification.go`.
- [x] Verification service: run claim-extraction on original + simplified, compare, produce `mismatch_detected` boolean + detail.
- [x] Wire into pipeline: verification runs after simplification, before chart processing. See `services/pipeline.go`.
- [ ] Test against Phase 0 corrupted-passage case — confirm it catches the deliberate error. **Not run** — requires a live `GEMINI_API_KEY` and an executable build, neither available in the build sandbox. Run this manually before trusting verification in production.
- [ ] Test against all 5 clean papers — confirm no false positives. **Not run**, same reason as above.
- [x] `claim_diffs` table writes. See `repository/claim_diffs.go`, wired transactionally in `handlers/documents.go`'s `saveResult`.

**Est. sessions**: 6-8

**Fallback if this phase stalls past 2x estimate**: claim-diff logic is isolable — if extraction accuracy is too unreliable to trust, degrade to manual-review-only for personal validation traffic and mark automated verification as a known limitation in AGENTS.md Known Issues, rather than blocking the whole project. Do not let this phase alone cause total abandonment.

---

## Phase 4 — Chart Re-visualization

**Done means**: At least 1 of the 5 test papers produces a correctly re-rendered chart via the data-extraction path, and image-fallback path is demonstrated on a paper where data extraction fails.

- [x] PDF table/data extraction — locate tabular data in extracted text, structure as JSON. See `services/charts.go` (`tryExtractChartData`).
- [x] PDF image extraction wrapper — extract chart images with page location. See `external/pdf.go` (`ExtractImages`, `ExtractTextByPage`).
- [x] Chart re-visualization service: attempt data path first; on failure, fall back to image + generate plain-language annotation via Gemini. See `services/charts.go` (`reVisualizeOne`).
- [x] `charts` table writes, linked to document. See `repository/charts.go`.
- [x] Wire into pipeline as the final stage before `complete` status.
- [ ] **Known gap** (see AGENTS.md Known Issues): the image-fallback path persists `image_blob` but the API contract and frontend have no way to serve/display the actual image yet — only the generated annotation shows. Flagged rather than fixed unilaterally, since it requires an API contract change (new endpoint or response field) outside this phase's original scope.
- [ ] Not run against the live 5-paper test set (no build/API key available in the build sandbox — same caveat as Phase 3).

**Est. sessions**: 8-12 (highest-risk phase — see PRD/Architecture risk note on combined image+data approach)

**Fallback if this phase stalls past 2x estimate**: this is the agreed risk area from scoping discussion. If data-based extraction proves too unreliable within budget, ship image-fallback-only for this phase (original chart + annotation, no re-rendering) as a reduced but shippable version, and revisit data-based extraction as a post-MVP iteration. Do not let this become the reason the whole project stalls at 20%.

---

## Phase 5 — Frontend

**Done means**: Full user flow (upload → select level → wait → view result with text + charts) works in browser end-to-end.

- [x] Vite + React + Tailwind + shadcn/ui scaffold. Demo/starter files removed per repo cleanliness rules.
- [x] Upload page: Upload Dropzone (file + paste-text toggle), Reading Level Selector, submit. See `frontend/src/pages/upload-page.jsx`.
- [x] Result page: Processing Indicator (polling `GET /api/documents/:id` at the required 2s minimum interval), Text Comparison Toggle, Chart Card list, Verification Warning Banner, Empty/Error states. See `frontend/src/pages/result-page.jsx`.
- [x] Wire frontend to backend API. See `frontend/src/lib/api.js`; dev-time proxy configured in `frontend/vite.config.js`.
- [x] Manual pass against DESIGN.md accessibility rules: 44×44px touch targets on buttons, focus-visible ring (`:focus-visible` in `index.css`), max-w-prose reading column. **Not tested with an actual screen reader or keyboard-only pass** — do that before shipping.
- [x] Frontend build verified with a real `vite build` — compiles cleanly, 0 errors, 0 lint warnings (oxlint). This is the one part of the stack fully compiler-verified in the build sandbox (the Go backend could not be, see Phase 1 note).

**Est. sessions**: 10-12

---

## Phase 6 — Expiry & Hardening

**Done means**: Documents actually expire after 7 days; app doesn't crash on malformed input.

- [x] Expiry sweep (startup + periodic interval, per ARCHITECTURE.md diagram) — delete documents where `last_accessed_at` older than 7 days. See `services/expiry.go`, started via `go services.RunExpirySweepLoop(...)` in `cmd/server/main.go`.
- [x] `last_accessed_at` update on every `GET /api/documents/:id` call. See `handlers/documents.go`'s `Get`.
- [x] Error handling pass: malformed uploads, oversized files, non-PDF MIME types (magic-byte check, not just header), empty paste-text. See `handlers/documents.go` and `handlers/validation.go`.
- [x] Structured logging pass (per ARCHITECTURE.md Logging Policy) — JSON via `log/slog`, document ID + byte length only, never full text content.
- [ ] Manual full-flow test against all 5 Phase 0 test papers. **Not run** — requires a live `GEMINI_API_KEY`, a working `go build`, and a real machine (the build sandbox cannot resolve `modernc.org/sqlite`'s dependency tree over its restricted network — see Phase 1 note). This is the single most important remaining step before calling the MVP done: run `go mod tidy && go build ./... && go test ./...` on a normal machine, set `GEMINI_API_KEY`, and walk all 5 test papers through the real app.

**Est. sessions**: 5-7

---

## Future Scope (explicitly deferred, not started until MVP validated)

- Interactive chart parameters (non-goal per ARCHITECTURE.md — requires separate computable-model extraction research).
- arXiv/URL fetch.
- User accounts, persistent history.
- OCR for scanned PDFs.
- TENDR integration (replacing direct Gemini calls) — revisit only once TENDR itself is stable.

## Parking Lot (ideas surfaced during planning, not acted on)

- Multi-language source paper support — surfaced during scoping, explicitly out of MVP, no action.
- Automated test-paper regression suite beyond the 5-paper manual set — worth doing eventually, not blocking MVP validation.
