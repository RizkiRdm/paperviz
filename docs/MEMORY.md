Goal
- P0 Core Features COMPLETE. All 7 chunks (A1-A2, B1-B5) committed and pushed.
- Awaiting user direction for next phase.
Constraints & Preferences
- Workflow: PLAN→APPROVE→EXECUTE, one chunk at a time, BUKTI SELESAI after each.
- AGENTS.md: UI work must follow DESIGN.md; layers handlers → services → repository/external.
- User model note: "the model i use is deepseek or sometimes model without image analysis" — do NOT delegate image/pixel checks; use DOM assertions.
- Anti-hallucination: never claim done without pasted real output.
Progress
Done
- P0 Batch A (claim-diff): A1 backend claim_diff in GET response; A2 frontend expandable claim comparison panel.
- P0 Batch B (chapters): B1 migration 003_chapters.sql; B2 ChapterRepo + type; B3 pipeline carries chapters; B4 handler saves + exposes; B5 frontend summary card.
- All committed: 0a9df0f through be738e9.
- Phase 1 SaaS Foundation: ALL 10 CHUNKS COMPLETE (2026-08-06).
In Progress
- (none) — P0 complete, awaiting next phase.
Blocked
- DB schema changed (003_chapters.sql) — dev DB needs reset before testing.
Key Decisions
- PipelineOutput moved from pipeline.go to types.go (redeclaration fix).
- MismatchDetail kept as *string in claimDiffResponse (matches repository type).
- Chapter summary card uses DESIGN.md tokens (12px radius, #e5e5e5 border, #f5f5f5 bg).
- ClaimComparisonPanel parses JSON claims client-side (data already on doc.claim_diff).
Next Steps
1. Delete paperviz.db* and restart server to apply new schema.
2. Upload PDF → verify chapters appear in GET response + "Sections in this paper" card.
3. Complete doc → click "Verified" badge → verify claim comparison panel toggles.
4. Await user direction for next phase.
Critical Context
- graphify-out/ updated (2217 nodes, 2112 edges).
- context7 available (Tailwind v4 docs); codegraph available.
- Repo: git@github.com:RizkiRdm/paperviz.git, branch main.
Relevant Files
- internal/handlers/documents.go: getDocumentResponse (claim_diff, chapters), saveResult (chapters loop).
- internal/repository/chapters.go: ChapterRepo (new).
- internal/repository/types.go: Chapter struct (new).
- internal/services/types.go: PipelineOutput with Chapters field.
- internal/services/pipeline.go: passes chapters through.
- migrations/003_chapters.sql: new migration.
- frontend/src/components/ui/status-banners.jsx: VerificationBadge (onClick), ClaimComparisonPanel (new).
- frontend/src/pages/result-page.jsx: chapter summary card, claim toggle state.
