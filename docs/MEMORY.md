Goal
- Execute PaperViz's PAPERVIZ_AGENT_TASKS.md chunks 0→9 in PLAN→APPROVE→EXECUTE order, one at a time; chunks 0-2 done, currently awaiting approval for CHUNK 3 (public-link disclaimer).
Constraints & Preferences
- Workflow (verbatim): "PLAN PHASE… Output a plan… Do NOT write or edit any code in this phase" → "STOP. Wait for explicit approval ('execute' / 'lanjut' / 'go') before touching any file."
- Strict chunk order 0 → 9; no batching into one plan.
- CHUNK 1: edit result-page.jsx only; "Do NOT change POLL_INTERVAL_MS (2000ms) — that's a documented backend contract"; "Do NOT touch lib/api.js (that's Chunk 5)."
- Chunk 2 spec: clipboard button next to "← New document"/VerificationBadge row, navigator.clipboard.writeText(window.location.href) with try/catch, fallback = visible selectable text input, "Copied!" ~2s via local useState + setTimeout, NO new toast dependency (grep package.json first — confirmed none exists).
- Chunk 3 spec: "do this chunk AFTER Chunk 2 is approved and merged"; disclaimer under copy button; exact proposed copy "Anyone with this link can view this document."
- Proof: raw npm run build output + manual/Playwright observation.
- AGENTS.md: UI work must follow DESIGN.md; layers handlers → services → repository/external.
- User model note (active): "the model i use is deepseek or sometimes model without image analysis, and just code" — do NOT delegate image/pixel checks to other models; use DOM assertions instead of pixel-by-pixel verification. Also "use deepseek v4 flash free (new) only for all case"; after failure stop after 1-2 attempts, report, don't retry in background.
- Anti-hallucination: never claim done without pasted real output.
Progress
Done
- CHUNK 0 (investigation only): chapter↔chart correlation = trust-the-LLM-with-no-check (charts.go:301-345 accepts JSON as-is; no excerpt grounding length limits; annotation fmt.Sprintf("From chapter: %s", chapter.Title) at charts.go:342).
- CHUNK 1 (polling timeout): POLL_TIMEOUT_MS = 120000, timedOut, retryNonce, pollStartRef, effect deps [documentId, retryNonce], timeout UI branch ("This is taking longer than usual." + Retry + Start over). Proven: forced 3000 → Playwright showed fallback UI at 3s persisting through 11s, Retry restarts polling, Start over → / upload page. Reverted to 120000, final build passed.
- CHUNK 2 (copy-link button): CopyLinkButton local component + COPY_FEEDBACK_MS = 2000 in result-page.jsx; placed top-right flex group beside VerificationBadge; Button import added from @/components/ui/button (secondary variant); success → "Copied!" 2s (accent-verified-soft tint + active:scale-[0.98] + transition); catch → visible read-only input "Copy this link manually" with onFocus select-all. Proven: Playwright stub captured writes: ["http://localhost:5173/doc123"], "Copied!" 2s revert verified, rejection path rendered input with correct URL. Build ✓ built in 6.33s.
- Committed + pushed: e7515c0 feat(frontend): polling timeout + copy-link button → origin/main (git@github.com:RizkiRdm/paperviz.git), f2d03bd..e7515c0.
- CHUNK 3 plan delivered, awaiting approval.
- Read in chunk 2: frontend/src/components/ui/button.jsx (secondary variant classes), frontend/package.json (no toast lib), frontend/src/index.css:106 (global :focus-visible ring exists), upload-page.jsx (layout patterns).
In Progress
- (none) — CHUNK 2 complete/merged; CHUNK 3 plan pending approval.
Blocked
- CHUNK 3 execution blocked on user approval ("execute"/"lanjut"/"go").
Key Decisions
- POLL_TIMEOUT_MS = 120000 assumption value, negotiable — approved.
- Clipboard fallback: visible selectable input instead of silent failure (spec-required precedent).
- Copy button shows for both complete and verification_failed; disclaimer sits OUTSIDE CopyLinkButton's aria-live="polite" wrapper (avoids re-announcing on state change).
- Screenshot /tmp/opencode/chunk2-copy-button.png saved but model can't view images — visual QA delegated to multimodal sub-agent aborted per user order; DOM assertions used instead.
- Left unrelated worktree changes uncommitted (AGENTS.md, docs/ moves, DESIGN.md, agent configs, paperviz binary, graphify-out/).
Next Steps
1. Get approval on CHUNK 3 plan → edit result-page.jsx: top row right cluster → flex flex-col items-end gap-1, disclaimer <span className="text-caption text-ink-secondary">Anyone with this link can view this document.</span> under copy button/badge.
2. npm run build raw output; Playwright DOM assertion of disclaimer text + outerHTML snippet; confirm copy button + badge no regression.
3. Commit + push chunk 3 (same pattern as chunk 1-2).
4. Plan CHUNK 4 next.
Critical Context
- Playwright MCP: playwright / browser_run_code_unsafe, route mocking **/api/documents/*, addInitScript clipboard stubbing pattern already proven.
- Vite dev on :5173 (5174 earlier). Kill via pkill -f vite after tests.
- Chunk 5 later: lib/api.js — AbortController, ApiError("network_timeout", 0), ERROR_MESSAGES additions.
- Chunk 9 later: security headers middleware; CSP must allow Google Fonts.
- Comment-hook fired on new comments — justified by mirroring existing file comment conventions.
- LSP typescript NOT installed/declined (lsp-install-decisions.json); build used as syntax check.
- context7 available (Tailwind v4 docs); codegraph available.
Relevant Files
- frontend/src/pages/result-page.jsx: chunk 1-2 changes (committed e7515c0); chunk 3 target (top row ~lines 160-167, CopyLinkButton ~lines 36-94).
- frontend/src/components/ui/button.jsx: shared Button (secondary variant used for copy button).
- frontend/src/components/ui/status-banners.jsx: VerificationBadge, WarningBanner, ErrorBanner.
- frontend/src/lib/api.js: off-limits until chunk 5.
- docs/PAPERVIZ_AGENT_TASKS.md, AGENTS.md, docs/PRD.md, docs/ARCHITECTURE.md, DESIGN.md: governing docs (all read).
- Repo: git@github.com:RizkiRdm/paperviz.git, branch main, HEAD e7515c0.
