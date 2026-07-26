Goal
- Make chart generation work for ALL paper uploads by fixing context propagation in Gemini retry loop + removing conflicting pipeline timeouts + raising top-level ceiling.
Constraints & Preferences
- Do NOT reduce maxRetries (10) or backoff (max 32s) — deliberate for free-tier survival
- Do NOT touch files not named in current chunk
- One single-sentence proof per chunk before next starts
- Commit messages follow caveman-commit style
- Log format: universal JSONL (time, message, severity keys)
Progress
Done
- Chunk 1: fixed `context.Background()`→`ctx` in gemini.go:87. `go build`+`go vet` pass.
- Chunk 1: added chartValues lenient UnmarshalJSON (string→[]float64)
- Chunk 1: added response_snippet logging to all ExtractChartsFromText exit paths
- Chunk 1: strengthened fullTextChartPrompt with academic stats patterns (F-test, mean/SD, regression)
- Chunk 1: added chart count log in RunPipeline
- Chunk 1/2: added pull request #1 test suite (17 tests, chartValues + struct embedding)
- Created internal/external/logger.go — custom JSONL handler (not yet wired)
In Progress
- Chunk 2: remove 4 stageTimeout wraps in pipeline.go, delete stageTimeout const
- Chunk 3: raise backgroundPipelineTimeout 5min→20min
- Chunk 4: add chart_extraction_degraded flag (7 files)
- Chunk 6: enable WAL mode in db.go (journal_mode=WAL, synchronous=NORMAL)
Blocked
- (none) until live paper verification after chunks 1-3
Key Decisions
- Chunk order: 1→2→3→VERIFY→4→6 (WAL)
- DB reset once: after chunk 4 code ready (new column + WAL in same wipe)
- WAL + synchronous=NORMAL: acceptable data-loss risk (ephemeral data, 7-day expiry)
- 3s stagger sleeps between stages: keep until live retry evidence proves they are redundant
Next Steps
1. Chunks 2-3 code changes → build/vet pass
2. Commit chunks 1-3 (squash as single logical fix)
3. Live paper upload → grep logs for retries, charts in response
4. If charts appear → chunk 4 (degraded flag) + chunk 6 (WAL)
5. DB reset → rebuild → test frontend
6. Final commit + push
Relevant Files
- internal/external/gemini.go: Generate() line 87 — context.Background() bug
- internal/services/pipeline.go: RunPipeline() — 4 stageTimeout wraps to remove
- internal/handlers/documents.go: backgroundPipelineTimeout — raise to 20min
- internal/services/charts.go: ExtractChartsFromText, chartValues, fullTextChartPrompt
- internal/repository/db.go: add WAL pragmas
- internal/repository/types.go: add ChartExtractionDegraded
- internal/repository/documents.go: Insert/UpdateStatus/Get to include new column
- migrations/001_init.sql: add chart_extraction_degraded column
- frontend/src/pages/result-page.jsx: degraded message when empty charts
