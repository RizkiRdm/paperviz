# Chunk X — Fix chart chapter-linkage bug (backend, Go)

## Scope
ONE concern: image-extracted charts (`reVisualizeOne` in `internal/services/charts.go`)
are silently assigned to chapter index 0 instead of "no chapter" because
`ChapterIndex` is left at Go's zero-value (`0`) instead of `-1`.

Do NOT touch: chapter detection logic, chapter chart generation
(`GenerateChapterChart`), frontend files, prompt wording, chart type logic.
Those are out of scope for this chunk even if you notice something you want
to improve. Flag anything else you find in your final report — do not fix it.

## Mandatory investigation (do this BEFORE writing any code)

1. Run: `grep -rn "ChapterIndex" internal/services internal/repository`
   Paste the full output in your report. Confirm your understanding matches:
   - `services/types.go`: `ChapterIndex int` field, comment says
     `-1 if not linked to a chapter`.
   - `services/charts.go`: `GenerateChapterChart` sets `ChapterIndex: displayOrder`
     (correct — this is the chapter-based chart path, expected to link).
   - `services/charts.go`: `reVisualizeOne`'s `base := Chart{...}` does NOT set
     `ChapterIndex` at all.
   - `services/intake.go` around line 182: `if c.ChapterIndex >= 0 { ... }`
     is the code that turns `ChapterIndex` into a `ChapterID` FK on save.
2. Run: `grep -n "chapterIndexToID" internal/services/intake.go` and read the
   surrounding ~15 lines. Confirm how the map is built and what happens when
   `ChapterIndex` doesn't match any key (should be a no-op / nil ChapterID).
3. Read `internal/services/save_pipeline_result_test.go` in full. Note the
   existing test cases around `ChapterIndex: -1` and `ChapterIndex: 0` — your
   fix must not break these, and you should add a new case that specifically
   covers an image-fallback/omitted chart (the `reVisualizeOne` output shape)
   getting `ChapterID == nil` after save.

If your grep results don't match what's described above, STOP and report the
discrepancy instead of proceeding — the investigation step exists to catch
exactly this ("the code changed since this task was written").

## The fix

In `internal/services/charts.go`, function `reVisualizeOne`, the `base :=
Chart{...}` struct literal currently looks like:

```go
base := Chart{
    PageNumber:   ec.PageNumber,
    DisplayOrder: displayOrder,
    SourceText:   pageContext,
}
```

Add the missing field:

```go
base := Chart{
    PageNumber:   ec.PageNumber,
    DisplayOrder: displayOrder,
    SourceText:   pageContext,
    ChapterIndex: -1,
}
```

That is the entire intended fix. If you find the struct literal doesn't match
this exactly (field order, other fields present), adapt accordingly but do
not add fields that weren't there before.

## Test requirement

1. Add a table-driven test case (or a new test function if that fits the
   existing file's convention better — check `charts_test.go` first) that
   asserts `reVisualizeOne` returns a `Chart` with `ChapterIndex == -1` for:
   - the data-extracted path (chart data found in page text)
   - the image-fallback path (annotation generated)
   - the omitted path (both failed)
2. Add or extend a `save_pipeline_result_test.go` case: given a document with
   2+ chapters and a chart with `ChapterIndex: -1`, assert the saved chart's
   `ChapterID` is `nil` in the database (not chapter 0's ID).
3. Do NOT modify any existing test's assertions, use `t.Skip()`, or change
   expected values to make a test pass. If an existing test fails after your
   change and you believe the test itself encoded the bug (i.e. it asserted
   the old, wrong behavior), report this explicitly in your final report —
   do not silently "fix" the test.

## Proof of completion (required, not optional)

Paste raw terminal output for:
```bash
go build ./...
go test ./internal/services/... -run Chart -v
go test ./internal/services/... -run SavePipelineResult -v
go vet ./...
```
A report that says "tests pass" without this raw output is not acceptable —
re-run and paste before submitting.

## Out of scope / do not touch
- `GenerateChapterChart` and `chapterChartPrompt` (chart-type variety —
  separate concern, separate chunk)
- Frontend `chapterCharts` filter logic in `result-page.jsx` (no change
  needed there — once `ChapterID` is correctly `nil`, confirm in your report
  whether the existing filter already handles `null` chapter_id charts
  correctly when `hasChapters` is true, i.e. whether they need a separate
  "General" bucket in the UI — but do NOT implement that; just report it,
  it's Chunk Y's or a follow-up's call)
- Any DB migration — `charts.chapter_id` is already nullable per
  `migrations/004_chapter_charts.sql`, no schema change needed
