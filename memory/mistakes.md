# Lessons and Mistakes

## 2026-08-24 — Public share page image serving needs no extra auth check

- **Problem:**担心 GetChartImage endpoint lacks visibility check for share pages.
- **Root cause:** Overthinking. Share page service (GetSharedFigure) already checks document visibility and only returns image URL for non-private docs. GetChartImage itself doesn't need visibility check — the URL is only exposed when appropriate.
- **Prevention:** Don't add auth checks to endpoints that are already gated by the data layer. Trace the full request path before adding security controls.

## 2026-08-18 — Test migration maps must be updated with every new migration

- **Problem:** After adding migration 006, all existing tests failed: "table documents has no column named title". Three separate test files each hardcode a migration map for their in-memory test DB.
- **Root cause:** openTestDB (charts_test.go), intake_test.go, and save_pipeline_result_test.go each have their own `map[int]string` of migrations. Adding a migration without updating ALL test helpers breaks every test that inserts into documents.
- **Fix:** Added migration 6 to all three test helpers. intake_test.go was also missing migration 5 (pre-existing gap from Chunk 1.1).
- **Prevention:** When adding a migration, grep for `001_init.sql` across all test files — every migration map must be updated. Consider a shared test helper in the future.

## 2026-08-17 — Radix primitives: three integration traps

- **Problem:** Share dialog never opened with Radix Dialog; tooltip never appeared on VerificationBadge; charts rendered twice in flat mode.
- **Root cause 1 (dialog):** Conditional-mount (`{showShare && <ShareDialog/>}`) + no `DialogTrigger` → uncontrolled Radix Root mounts in its default closed state forever. Fix: pass controlled `open` prop.
- **Root cause 2 (tooltip):** `TooltipTrigger asChild` clones the child and forwards ref + pointer handlers. `VerificationBadge({ onClick })` destructured only `onClick`, silently dropping Radix's props → tooltip dead. Fix: spread `...props` onto the button. Any component wrapped in asChild must forward refs + spread unknown props.
- **Root cause 3 (duplicate render):** Pre-existing bug from chapters feature — in flat mode (`hasChapters` false) `chapterCharts` fell back to ALL charts, rendering them in both the tabpanel section and the flat section. Only visible when a doc has image-origin charts but 1 chapter. Fix: guard tabpanel charts section with `hasChapters`.
- **Prevention:** Radix asChild children must spread props. Controlled vs uncontrolled Radix Root depends on mount pattern — conditional mount needs explicit `open`. Re-check render-site conditions when adding a mode toggle (chapter tabs).

## 2026-08-15 — Rules of Hooks violation caused React crash on status transition

- **Problem:** "Something went wrong. Please refresh the page." error on result page when document transitions from `processing` to `complete`.
- **Root cause:** `useState(hasChapters ? 0 : -1)` placed after conditional early returns in `result-page.jsx`. Hook count changed between renders (28→29) when status flipped, triggering React invariant error caught by ErrorBoundary.
- **Fix:** Moved `useState` to top of component with other hooks (initialized to `-1`), added `useEffect` to set to `0` when chapters first appear.
- **Prevention:** All hooks must be declared unconditionally before any early returns. Comment added to prevent regression.

## 2026-08-15 — chartData.type never matched backend JSON key chart_type

- **Problem:** All charts rendered as bar type regardless of what Gemini specified, because `data-chart.jsx` read `chartData.type` but backend JSON used `chart_type`.
- **Root cause:** Frontend referenced wrong JSON key. Pre-existing bug from initial chart implementation.
- **Fix:** Changed `chartData.type` to `chartData.chart_type`.
- **Prevention:** Verify JSON key names match between backend struct tags and frontend property access.

## 2026-08-15 — Pre-existing gofmt drift needed normalization before commit

- **Problem:** `gofmt -l` flagged `internal/services/types.go`, `internal/services/charts_test.go`, and a closing block in `charts.go` — files untouched by the current chunk.
- **Root cause:** Prior commits shipped unformatted alignment (whitespace drift) that `go test` tolerates but `gofmt -l` catches.
- **Fix:** `gofmt -w` on the flagged files before committing (AGENTS.md mandates gofmt clean).
- **Prevention:** Run `gofmt -l internal/ cmd/`/`gofmt -l .` as part of validation every time, not just "gofmt on files I edited". Normalize pre-existing drift rather than committing a mixed-format tree.

## 2026-08-15 — Stale LSP diagnostics after import fixes

- **Problem:** After adding a missing import (`errors`, `sql`), the editor's LSP continued reporting "undefined" at the previous line, twice in a row, while the code was already correct.
- **Root cause:** Language-server diagnostics lag one edit behind on incremental Go analysis; fixes that don't recompile leave stale markers.
- **Fix:** Ignored LSP output; verified with `go build`/`go vet`/`go test` which passed.
- **Prevention:** Treat compiler/`go vet` output as authoritative over inline LSP diagnostics after edit cycles; run `go test ./...` before considering a fix complete.

## 2026-08-14 — Migration-count test broke after adding migration 005

- **Problem:** `go test` failed after adding `migrations/005_evidence.sql` — `TestLoadMigrationsRegistersChapterCharts` expected exactly 4 migrations.
- **Root cause:** `cmd/server/main_test.go` asserts `len(migrations) == len(want)`; adding a migration without updating the test's expected map fails the count assertion.
- **Fix:** Added `5: {"CREATE TABLE evidence"}` to the expected map. 49/49 tests pass.
- **Prevention:** Any new migration must be added to BOTH `loadMigrations` (main.go) AND the migration-count test map. Run `go test ./...` after every migration.

## 2026-08-02 — Static import remained inside lazy wrapper

- **Problem:** First Recharts fix moved the static import to a lazy-loaded file, but React Doctor still reported it.
- **Root cause:** The rule detects any static `ImportDeclaration`; it does not infer that the containing module is lazy-loaded.
- **Fix:** Use runtime `import("recharts")`, with rejection handling and unmount protection.
- **Prevention:** Re-run the exact diagnostic tool after each canonical fix; never infer rule behavior from bundle behavior alone.

## 2026-08-02 — Dynamic import lacked rejection handling

- **Problem:** Initial runtime import introduced a new React Doctor warning for `.then` state updates without `.catch`.
- **Root cause:** Successful loading path was implemented before considering chunk/network failure.
- **Fix:** Add load-failure state and visible fallback copy.
- **Prevention:** Every dynamic import must define loading, success, failure, and unmount behavior before implementation.

## 2026-08-02 — File input was initially nested inside button

- **Problem:** Converting the dropzone `div` to a button initially left the hidden file input nested inside it.
- **Root cause:** Mechanical tag replacement overlooked nested interactive-content validity and click bubbling risk.
- **Fix:** Move hidden input beside the native button.
- **Prevention:** When replacing static markup with native controls, inspect descendants for nested interactive elements before runtime QA.

## 2026-08-02 — Playwright QA script syntax error

- **Problem:** First combined runtime QA script failed with `Unexpected token ';'`.
- **Root cause:** Route-fulfillment callback had an extra closing delimiter.
- **Fix:** Rewrote and reran the script successfully.
- **Prevention:** Keep browser scripts small or format multiline callbacks before execution.
