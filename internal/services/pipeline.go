package services

import (
	"context"
	"log/slog"
	"time"

	"paperviz/internal/external"
)

// maxImageChartsPerDocument caps how many embedded-image charts are sent
// to Gemini per document. Each image can cost up to 2 Gemini calls
// (data-extraction attempt + annotation fallback), so this directly bounds
// free-tier quota usage per document. Images beyond this cap are marked
// "omitted" rather than processed — this is a deliberate cost/completeness
// tradeoff, not a bug.
const maxImageChartsPerDocument = 5

// PipelineInput is everything the pipeline needs to process one document.
// SourceType distinguishes PDF uploads (which get chart processing) from
// pasted text (which skips it — PRD.md Acceptance Scenario 2: "chart
// pipeline skipped, no table/figure data available from plain text").
type PipelineInput struct {
	OriginalText string
	SourceType   string             // "pdf" | "pasted_text"
	ReadingLevel string             // "simplified" | "eli5"
	PDFBytes     []byte             // nil for pasted_text; used only for chart extraction
	OnStage      func(stage string) // called at each pipeline stage transition
}

// Pipeline status values. See charts.go for why these are redeclared
// locally instead of imported from repository.
const (
	pipelineStatusComplete           = "complete"
	pipelineStatusFailed             = "failed"
	pipelineStatusVerificationFailed = "verification_failed"
)

// RunPipeline is the single sequential extract -> simplify -> verify ->
// chart flow required by ARCHITECTURE.md Section 5 ("Pipeline Service MUST
// orchestrate the full flow ... as an explicit sequential function"). It
// does not touch the database or HTTP layer directly — the caller
// (handlers/documents.go) is responsible for turning PipelineOutput into
// persisted rows via the repository layer, inside a single transaction.
//
// Read this function top-to-bottom; it IS the product's core logic. Every
// early return corresponds to one of ARCHITECTURE.md Section 6's Failure
// Scenarios — the comment above each return says which one.
func RunPipeline(ctx context.Context, gemini *external.GeminiClient, in PipelineInput) PipelineOutput {
	stage := func(s string) {
		if in.OnStage != nil {
			in.OnStage(s)
		}
	}

	stage("simplifying")
	simplifiedText, err := Simplify(ctx, gemini, in.OriginalText, in.ReadingLevel)
	if err != nil {
		// Failure Scenario 2: Gemini call times out after retry -> status
		// "failed", error_message populated, nothing published.
		slog.Error("pipeline stage failed", "stage", "simplify", "error", err)
		return PipelineOutput{
			Status:       pipelineStatusFailed,
			ErrorMessage: "simplification_failed",
		}
	}

	// Stage 2: Verify. Runs BEFORE chart processing (ARCHITECTURE.md
	// Section 6 sequence diagram) — a document that fails claim-diff never
	// reaches the chart pipeline, since it won't be published as complete
	// regardless of chart quality.
	//
	// Stagger between Gemini-heavy stages so free-tier rate limits recover.
	time.Sleep(3 * time.Second)
	stage("verifying")
	verifyResult, err := DiffClaims(ctx, gemini, in.OriginalText, simplifiedText)
	if err != nil {
		slog.Error("pipeline stage failed", "stage", "verify", "error", err)
		return PipelineOutput{
			Status:       pipelineStatusFailed,
			ErrorMessage: "verification_failed_to_run",
		}
	}

	if verifyResult.MismatchDetected {
		// Acceptance Scenario 4: mismatch detected -> status
		// "verification_failed", result page shows a warning banner. This is
		// NOT the same as status "failed" — the simplified text still exists
		// and is still returned to the client, just flagged as unverified
		// (see DESIGN.md's --state-warning tokens for how the frontend
		// distinguishes this from a hard failure).
		return PipelineOutput{
			Status:         pipelineStatusVerificationFailed,
			SimplifiedText: simplifiedText,
			Verify:         verifyResult,
		}
	}

	// Stage 3: Chart re-visualization.
	//
	// Primary path: detect chapters/sections from simplified text, then
	// generate at most one chart per chapter. This produces charts tied to
	// the paper's actual structure rather than a flat scan for any number.
	//
	// Supplemental path: for PDFs that DO have embedded chart images,
	// run per-image data extraction on top of the chapter charts.
	time.Sleep(3 * time.Second)
	stage("generating_charts")
	var charts []Chart
	var chartDegraded bool

	chapters, err := DetectChapters(ctx, gemini, simplifiedText)
	if err != nil {
		slog.Error("pipeline: chapter detection failed", "stage", "chapters", "error", err)
		chartDegraded = true
	} else if len(chapters) == 0 {
		slog.Info("pipeline: no chapters detected, skipping chart generation", "stage", "chapters")
	} else {
		for i, chapter := range chapters {
			chart, ok, degraded := GenerateChapterChart(ctx, gemini, chapter, i)
			if degraded {
				chartDegraded = true
			}
			if ok {
				charts = append(charts, chart)
			}
		}
		slog.Info("pipeline: chapter-based chart generation complete",
			"stage", "chart",
			"chapters_detected", len(chapters),
			"charts_generated", len(charts),
		)
	}

	// Supplemental: image-based extraction (PDFs with embedded chart images).
	if in.SourceType == "pdf" && len(in.PDFBytes) > 0 {
		pdfDoc, err := ParsePDF(in.PDFBytes, maxImageChartsPerDocument)
		if err == nil {
			pages := pageText(pdfDoc.Pages)
			if pages == nil {
				pages = pageText{1: pdfDoc.Text}
			}
			imageCharts := ReVisualizeCharts(ctx, gemini, pdfDoc.Charts, pages)
			// Offset display order to append after text-scan charts.
			offset := len(charts)
			for i := range imageCharts {
				imageCharts[i].DisplayOrder = offset + i
			}
			charts = append(charts, imageCharts...)
		} else {
			slog.Error("chart extraction from PDF failed", "stage", "chart", "error", err)
		}
	}

	return PipelineOutput{
		Status:                  pipelineStatusComplete,
		SimplifiedText:          simplifiedText,
		Verify:                  verifyResult,
		Charts:                  charts,
		ChartExtractionDegraded: chartDegraded,
		Chapters:                chapters,
	}
}
