# PaperViz P0 — Core Feature Agent Task Chunks

Target: low-effort agent.
Scope: ONLY the two P0 items — (A) expose existing claim-diff data,
(B) persist + expose chapter structure. Nothing else.

## Ground Rules
- NEVER modify test assertions, use `t.Skip()`, or change expected values.
- ALWAYS `grep -rn` the target symbol/file BEFORE editing it. Paste the
  output as proof you looked.
- ALWAYS end each chunk with `BUKTI SELESAI`: raw terminal output (build,
  test, or curl) proving the change works. No exceptions.
- Do not start chunk N+1 until chunk N's BUKTI SELESAI is confirmed.
- One concern per chunk. Do not bundle fixes across chunks.

---

## BATCH A — Expose existing claim-diff verification detail

Fact: `claim_diffs` rows are already inserted in `saveResult`
(`internal/handlers/documents.go`) via `ClaimDiffRepo.Insert`. They are
never read back. This batch only wires existing data through — no new
Gemini calls, no new prompts.

### A1 — Backend: add claim_diff to the Get response

**File:** `internal/handlers/documents.go`

1. Run `grep -n "getDocumentResponse\|chartResponse\|func (h \*DocumentHandler) Get" internal/handlers/documents.go` first. Paste output.
2. Add a new response struct near `chartResponse`:
   ```go
   type claimDiffResponse struct {
       OriginalClaims   json.RawMessage `json:"original_claims,omitempty"`
       SimplifiedClaims json.RawMessage `json:"simplified_claims,omitempty"`
       MismatchDetected bool            `json:"mismatch_detected"`
       MismatchDetail   string          `json:"mismatch_detail,omitempty"`
   }
   ```
3. Add `ClaimDiff *claimDiffResponse \`json:"claim_diff,omitempty"\`` field to `getDocumentResponse`.
4. In the `Get` handler, after the existing `chartRepo` block and before
   `writeJSON`, add:
   ```go
   claimDiffRepo := repository.NewClaimDiffRepo(h.db)
   var claimDiffResp *claimDiffResponse
   if cd, err := claimDiffRepo.GetByDocument(id); err == nil {
       claimDiffResp = &claimDiffResponse{
           OriginalClaims:   json.RawMessage(cd.OriginalClaims),
           SimplifiedClaims: json.RawMessage(cd.SimplifiedClaims),
           MismatchDetected: cd.MismatchDetected,
           MismatchDetail:   cd.MismatchDetail,
       }
   }
   ```
   Treat `err != nil` as "no claim_diff yet" (e.g. document still
   processing) — do NOT return an error to the client for this case.
5. Pass `ClaimDiff: claimDiffResp` into the `getDocumentResponse{}` literal.

**Constraints:** `OriginalClaims`/`SimplifiedClaims` in the repository
struct are already JSON-encoded strings (per `ClaimDiffRepo.Insert`
comment) — do NOT re-marshal them, just wrap in `json.RawMessage`.

**BUKTI SELESAI:** `go build ./...` output with no errors, plus `curl`
output of `GET /api/documents/{id}` for a `complete` or
`verification_failed` document showing the new `claim_diff` field
populated with real claim arrays.

---

### A2 — Frontend: expandable claim comparison under the Verified badge

**Files:** `frontend/src/components/ui/status-banners.jsx`,
`frontend/src/pages/result-page.jsx`

1. Run `grep -n "VerificationBadge" frontend/src/components/ui/status-banners.jsx frontend/src/pages/result-page.jsx` first. Paste output.
2. In `status-banners.jsx`, change `VerificationBadge` to accept an
   `onClick` prop and render as a `<button>` instead of `<span>`, keeping
   all existing classes. Do not change its visual appearance when
   collapsed — only make it clickable.
3. Add a new component in the same file:
   ```jsx
   export function ClaimComparisonPanel({ claimDiff, onClose }) {
     const original = JSON.parse(claimDiff.original_claims || "[]")
     const simplified = JSON.parse(claimDiff.simplified_claims || "[]")
     return (
       <div className="mt-3 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4">
         <div className="flex items-center justify-between mb-3">
           <p className="text-xs font-semibold text-[#0a0a0a]">Claims checked</p>
           <button onClick={onClose} className="text-[11px] text-[#737373] hover:text-[#0a0a0a]">Hide</button>
         </div>
         <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
           <div>
             <p className="font-medium text-[#737373] mb-1.5">Original</p>
             <ul className="space-y-1 text-[#171717]">
               {original.map((c, i) => <li key={i}>• {c}</li>)}
             </ul>
           </div>
           <div>
             <p className="font-medium text-[#737373] mb-1.5">Simplified</p>
             <ul className="space-y-1 text-[#171717]">
               {simplified.map((c, i) => <li key={i}>• {c}</li>)}
             </ul>
           </div>
         </div>
       </div>
     )
   }
   ```
4. In `result-page.jsx`: grep where `<VerificationBadge />` is rendered
   (header area). Add `const [showClaims, setShowClaims] = useState(false)`,
   pass `onClick={() => setShowClaims(v => !v)}` to `VerificationBadge`,
   and render `{showClaims && doc.claim_diff && <ClaimComparisonPanel claimDiff={doc.claim_diff} onClose={() => setShowClaims(false)} />}` directly below the header, inside `<main>`.

**Constraints:** Do not fetch claim data separately — it must already be
on `doc` from the A1 API change. Do not add a new API call.

**BUKTI SELESAI:** `npm run build` output with no errors, plus a
description (or screenshot filename if available) confirming clicking
the "Verified" badge reveals the two claim columns and clicking "Hide"
collapses it again.

---

## BATCH B — Persist and display chapter structure

Fact: `DetectChapters()` (`internal/services/chapters.go`) already
produces `Title`/`Summary`/`Excerpt` per chapter but the result is
discarded after chart generation in `pipeline.go`. This batch persists it
and renders it as document structure.

### B1 — Backend: migration for chapters table

**File:** new `migrations/003_chapters.sql`

1. Run `grep -n "CREATE TABLE" migrations/001_init.sql` first to match
   existing style (TEXT PRIMARY KEY id, INTEGER for booleans, etc).
2. Create:
   ```sql
   CREATE TABLE chapters (
       id TEXT PRIMARY KEY,
       document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
       title TEXT NOT NULL,
       summary TEXT NOT NULL,
       excerpt TEXT NOT NULL,
       display_order INTEGER NOT NULL
   );

   CREATE INDEX idx_chapters_document_id ON chapters(document_id);
   ```
3. Do NOT touch `001_init.sql` or `002_*.sql` if it exists — additive
   migration only.

**BUKTI SELESAI:** Terminal output of the migration applying cleanly to a
fresh SQLite file, plus `.schema chapters` output.

---

### B2 — Backend: repository for chapters

**File:** new `internal/repository/chapters.go`

1. Run `grep -n "type Chart\b\|ChartRepo" internal/repository/*.go` first
   — copy the exact structure/style of the existing `ChartRepo`
   (`charts.go` in repository package, NOT services), it is the closest
   analog (one document has many charts, ordered by `display_order`).
2. Add `ChapterRepo` struct with the same `dbExecutor` pattern, and two
   methods: `Insert(c Chapter) error` (single row) and
   `ListByDocument(documentID string) ([]Chapter, error)` (ordered by
   `display_order ASC`).
3. Add a matching `Chapter` struct to `internal/repository/types.go`:
   `ID, DocumentID, Title, Summary, Excerpt string; DisplayOrder int`.

**Constraints:** Follow the exact same file organization as
`internal/repository/charts.go` — do not invent a different pattern.

**BUKTI SELESAI:** `go build ./...` with no errors.

---

### B3 — Backend: carry chapters through the pipeline

**File:** `internal/services/pipeline.go`, `internal/services/types.go`

1. Run `grep -n "DetectChapters\|Chapters\|PipelineOutput" internal/services/pipeline.go internal/services/types.go` first.
2. Add `Chapters []Chapter` field to `PipelineOutput` in `types.go`.
3. In `pipeline.go`, in `RunPipeline`, the `chapters` variable already
   exists (from `DetectChapters`). Add `Chapters: chapters` to every
   `return PipelineOutput{...}` statement in the `pipelineStatusComplete`
   path. Chapters are unavailable for the `failed` / `verification_failed`
   early returns — leave those as-is (nil slice is fine).

**Constraints:** Do not change chart generation logic, prompts, or the
chapter-detection call itself in this chunk — pass-through only.

**BUKTI SELESAI:** `go build ./...` with no errors, plus the diff shown
for `types.go` and `pipeline.go`.

---

### B4 — Backend: save chapters + expose in Get response

**Files:** `internal/handlers/documents.go`

1. Run `grep -n "func (h \*DocumentHandler) saveResult" internal/handlers/documents.go` first, read the full function body it returns.
2. Inside `saveResult`'s transaction (same tx used for charts and
   claim_diff), loop `output.Chapters` and call
   `repository.NewChapterRepo(tx).Insert(...)` for each, with
   `DisplayOrder` = loop index. Generate IDs the same way charts do in
   this function (grep for how chart IDs are generated in this same
   function first, reuse that exact pattern).
3. In the `Get` handler (same file), add a `chapterRepo.ListByDocument(id)`
   call alongside the existing `chartRepo` call, map to a
   `chapterResponse` struct (`ID, Title, Summary, DisplayOrder` —
   deliberately omit `Excerpt`, the full text is not needed by the
   frontend and bloats the payload), add `Chapters []chapterResponse` to
   `getDocumentResponse`.

**Constraints:** If `output.Chapters` is empty (short papers, or pasted
text with no chapters detected), the loop must simply do nothing — do not
error.

**BUKTI SELESAI:** `go build ./...` output, plus `curl` output of
`GET /api/documents/{id}` for a newly processed PDF document showing a
non-empty `chapters` array with `title` and `summary` fields.

---

### B5 — Frontend: render chapter headings + sticky TOC

**File:** `frontend/src/pages/result-page.jsx`

1. Run `grep -n "displayedText\|<article" frontend/src/pages/result-page.jsx` first — this is the block being changed.
2. If `doc.chapters && doc.chapters.length > 0`: render a sticky nav list
   above `<article>` (same visual style as the existing action bar — reuse
   `border-[#e5e5e5]`, `rounded-[12px]` tokens from DESIGN.md, do not
   invent new colors) listing each chapter's `title` as an anchor link
   (`#chapter-{index}`).
3. Chapters currently have no reliable anchor point inside
   `displayedText` (it is one flat string, not split by chapter). For
   this chunk, do NOT attempt to split the paragraph text by chapter
   boundaries — that requires a separate, riskier text-matching chunk.
   Instead: render the chapter list as a simple "Sections in this paper"
   summary card (title + one-line summary each, no anchor jump yet)
   directly above `<article>`. This alone already fixes the "wall of
   text with zero structure" problem even without inline anchors.
4. If `doc.chapters` is empty or absent, render nothing extra — must not
   break documents processed before this change shipped (no chapters
   persisted for old rows).

**BUKTI SELESAI:** `npm run build` output with no errors, plus
confirmation that a document with `chapters` populated shows the new
summary card, and a document with empty/missing `chapters` renders
exactly as before (no crash, no empty box).

---

## Explicitly out of scope for this batch

Splitting `displayedText` into per-chapter blocks with real inline
anchors, chart-in-context placement, and chart annotation rewrites are
separate P1 work — do not attempt them inside these chunks.
