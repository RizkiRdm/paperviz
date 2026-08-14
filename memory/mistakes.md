# Lessons and Mistakes

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
