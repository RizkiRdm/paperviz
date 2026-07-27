Goal: security hardening, Gemini call-count reduction, chapter-based chart pipeline redesign.

Constraints:
- maxRetries=10, backoff max 32s — untouched (free-tier survival).
- Image fallback path (ReVisualizeCharts) unchanged by Group C.
- textChartElem kept for test compat (charts_test.go).
- B3 skipped (needs live API key + real-paper regression testing).

Progress
Done (Round 2):
- A1: Removed debug log line from gemini.go:174 (leaked model+URL)
- A2: Added IP rate limiting on POST /api/documents (1/30s, burst 2) — internal/handlers/ratelimit.go, router.go
- A3: Added slog.Warn on PDF extraction timeouts in pdf.go (ExtractText, ExtractTextByPage, ExtractImages)
- B1: Capped maxImageChartsPerDocument=5 in pipeline.go, logged when truncated
- B2: Merged 2 claim-extraction calls into 1 (dualClaimExtractionPrompt) — DiffClaims: 3→2 Gemini calls
- C1: Added Chapter struct (types.go) + DetectChapters() (new chapters.go) — splits simplified text into ≤10 chapters
- C2: Added GenerateChapterChart() + chapterChartPrompt to charts.go — per-chapter, varied types (bar/line/pie/scatter)
- C3: Replaced ExtractChartsFromText full-text-scan with DetectChapters→GenerateChapterChart loop in pipeline.go. Removed fullTextChartPrompt, ExtractChartsFromText. Image fallback untouched.
- graphify update ran

In Progress:
- Live verification with 3 test papers (needs API key + running server)

Blocked:
- None

Key Decisions:
- Rate limit POST only, not GET (GET doesn't call Gemini)
- 30s/2 burst deliberate — conservative on free-tier key
- Chapter detection uses simplifiedText (not originalText) — chapters match what reader sees
- Max 10 chapters, max 5 image charts — combined ceiling 15 Gemini chart calls per doc
- textChartElem kept despite ExtractChartsFromText removal (test compats)
- B3 deferred (optional, needs real-paper regression)

Next Steps:
1. DB reset + server restart
2. Upload 3 papers (different structure), verify chapters detected + charts varied
3. Report chapter count, chart type distribution, total Gemini calls
4. Commit + push all changes

Relevant Files:
- internal/external/gemini.go: line 174 removed
- internal/handlers/ratelimit.go: new — per-IP token bucket
- internal/handlers/router.go: POST route wrapped with rateLimitDocumentCreate
- internal/external/pdf.go: slog.Warn on 3 timeout branches
- internal/services/pipeline.go: maxImageChartsPerDocument=5, chapter-based chart flow
- internal/services/charts.go: GenerateChapterChart, chapterChartPrompt; removed fullTextChartPrompt, ExtractChartsFromText
- internal/services/chapters.go: new — DetectChapters
- internal/services/types.go: new Chapter struct
- internal/services/verification.go: dualClaimExtractionPrompt, DiffClaims 3→2 calls
- go.mod: added golang.org/x/time/rate v0.15.0
