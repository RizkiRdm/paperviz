Goal
- Chart generation working for ALL paper uploads.
Progress
- Chunk 1 ctx fix: `context.Background()`→`ctx` in gemini.go:87
- Chunk 2 stageTimeout rm: 4 wraps + const deleted from pipeline.go
- Chunk 3 timeout raise: `backgroundPipelineTimeout` 5min→20min
- Chunk 4 degraded flag: `chart_extraction_degraded` across 7 files
  - ExtractChartsFromText returns `([]Chart, bool)` — true on err, false on legit empty
  - DB column, API field, frontend msg when degraded+empty
- Chunk 6 WAL: journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000 in db.go
- chartValues lenient UnmarshalJSON (string→[]float64)
- fullTextChartPrompt: academic stats (F-test, mean/SD, regression)
- maxTokens 2048→0 (truncation cause)
- trailing-text trim: parse-first, trim-fallback for `[]` and `{}`
- response_snippet logging (2000 chars) on parse failure
- Frontend chart-card: rm JSON.parse — backend sends native JSON obj, not string
- Frontend result-page: distinct msg for degraded vs legit empty
- 17 tests: chartValues + struct embedding. All pass.
- logger.go created (not yet wired)
- DB reset: fresh schema + WAL after chunk 4+6
- All commits pushed.

Key Decisions
- WAL+synchronous=NORMAL OK (ephemeral data, 7-day expiry)
- maxRetries=10, backoff max 32s preserved (free-tier survival)
- 3s stagger sleeps: kept (no live retry evidence yet)
- DB reset once after chunk 4 code ready

Remaining
- 70 charts per paper excessive — prompt over-extracts. Separate quality task.
- logger.go wiring in main.go (deferred)
- image_fallback endpoint missing (known issue, not blocking)<｜end▁of▁thinking｜>

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="bash">
<｜｜DSML｜｜parameter name="command" string="true">rtk git add docs/MEMORY.md && rtk git commit -m 'docs: compact MEMORY.md to session-completed state' && rtk git push