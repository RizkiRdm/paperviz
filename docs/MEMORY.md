Goal
Security hardening, Gemini call-count reduction, and chapter-based chart pipeline redesign for PaperViz.
Constraints & Preferences
- maxRetries=10 and backoff (max 32s) in gemini.go — never touch (free-tier survival)
- Image fallback path (ReVisualizeCharts) unchanged by Group C
- textChartElem kept for test compat (charts_test.go)
- B3 (single-call verification) deferred — needs live API key + real-paper regression testing
- All chunks sequential, each verified with go build + go vet + go test before next starts
- After code changes, always update AGENTS.md, MEMORY.md, PLAN.md + run graphify update
Progress
Done
- A1: Removed debug log line (slog.Info("gemini debug", ...)) from gemini.go:174 — was leaking model name + URL on every Gemini call
- A2: Added IP-based rate limiting on POST /api/documents (1 req/30s, burst 2) via golang.org/x/time/rate. New file internal/handlers/ratelimit.go. Only POST wrapped, GET unrestricted. Client gets HTTP 429 {"error":"rate_limited"} on 3rd request within 30s.
- A3: Added slog.Warn logs on PDF extraction timeout branches in pdf.go (ExtractText, ExtractTextByPage, ExtractImages) — makes leaked goroutines observable in logs.
- B1: Added maxImageChartsPerDocument = 5 constant in pipeline.go. Truncates extracted.Charts before passing to ReVisualizeCharts. Logs when truncation occurs.
- B2: Replaced 2 separate extractClaims calls with 1 dual-extraction call using dualClaimExtractionPrompt in verification.go. DiffClaims now 2 Gemini calls total (down from 3). Deleted extractClaims function and old claimExtractionPrompt.
- C1: Added Chapter struct to types.go. Created internal/services/chapters.go with DetectChapters() — calls Gemini once to split simplified text into ≤10 chapters. Not wired into pipeline yet.
- C2: Added chapterChartPrompt + chapterChartJSON struct + GenerateChapterChart() to charts.go. Per-chapter chart generation with varied types (bar/line/pie/scatter). Not wired into pipeline yet.
- C3: Replaced Stage 3 text-scan path in pipeline.go (ExtractChartsFromText) with chapter-based flow (DetectChapters → per-chapter GenerateChapterChart). Removed fullTextChartPrompt, ExtractChartsFromText from charts.go. Kept textChartElem (test compat). Image fallback path unchanged.
- go build ./... / go vet ./... / go test ./... (17/17 pass) — all clean.
- graphify update . ran (1915 nodes, 1796 edges, 211 communities)
- Updated AGENTS.md, docs/MEMORY.md, docs/PLAN.md
- git commit + push: feat(charts,security): chapter-based pipeline, rate limit, claim merge
In Progress
- Live verification: upload 3 test papers to verify chapters detected + chart types varied + Gemini call counts (needs running server + GEMINI_API_KEY)
Blocked
- None
Key Decisions
- Rate limit POST only (not GET) — GET endpoint doesn't call Gemini, so no quota risk
- 30s/2 burst deliberate — conservative for free-tier key protection
- Chapter detection uses simplifiedText (not in.OriginalText) — chapters match what reader sees
- Max 10 chapters + max 5 image charts = combined ceiling 15 Gemini chart calls per doc
- textChartElem kept despite ExtractChartsFromText removal because charts_test.go uses it to test chartValues unmarshal
- B3 deferred (optional) — needs live API key + output comparison against B2 on real papers before shipping
Next Steps
1. DB reset + server restart (kill server; rm paperviz.db*; go run ./cmd/server)
2. Upload 3 test papers of different structure (short blog-style, standard 6-section, long multi-subsection)
3. For each paper, report: chapters detected (count + titles), charts generated (count + type distribution), total Gemini calls, relevance spot-check
4. Confirm chart types are NOT 100% bar across all papers
Critical Context
- Free-tier gemini-2.5-flash-lite has 20-request sliding window quota (~1 req/3.5s)
- B3 (combined single verification call) described in plan but intentionally deferred — merging extraction + comparison into one call risks accuracy regression
- Combined ceiling: 10 chapter charts + 5 image fallback charts = 15 per document
- Old fullTextChartPrompt and ExtractChartsFromText fully removed — cannot fall back to old full-text-scan path
Relevant Files
- internal/external/gemini.go: debug log line removed (L174)
- internal/external/pdf.go: slog.Warn added on 3 timeout branches (ExtractText, ExtractTextByPage, ExtractImages)
- internal/handlers/ratelimit.go: new — per-IP token bucket (1 req/30s, burst 2)
- internal/handlers/router.go: POST route wrapped with rateLimitDocumentCreate middleware
- internal/services/pipeline.go: maxImageChartsPerDocument=5, chapter-based chart flow replaces ExtractChartsFromText
- internal/services/charts.go: new GenerateChapterChart, chapterChartPrompt, chapterChartJSON; removed fullTextChartPrompt, ExtractChartsFromText; kept textChartElem
- internal/services/chapters.go: new — DetectChapters(), chapterDetectionPrompt, maxChapters=10
- internal/services/types.go: new Chapter struct (Title, Summary, Excerpt)
- internal/services/verification.go: dualClaimExtractionPrompt, dualClaimExtractionResult; deleted extractClaims, claimExtractionPrompt; DiffClaims now 2 calls (down from 3)
- go.mod / go.sum: added golang.org/x/time/rate v0.15.0
- AGENTS.md: updated with Round 2 changes, known issues
- docs/MEMORY.md: updated with Round 2 progress
- docs/PLAN.md: new Phase 7 tracking all chunks