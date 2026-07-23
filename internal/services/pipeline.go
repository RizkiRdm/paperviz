package services

import (
	"context"
	"log/slog"
	"time"

	"paperviz/internal/external"
)

// PipelineInput is everything the pipeline needs to process one document.
// SourceType distinguishes PDF uploads (which get chart processing) from
// pasted text (which skips it — PRD.md Acceptance Scenario 2: "chart
// pipeline skipped, no table/figure data available from plain text").
type PipelineInput struct {
	OriginalText string
	SourceType   string // "pdf" | "pasted_text"
	ReadingLevel string // "simplified" | "eli5"
	PDFBytes     []byte // nil for pasted_text; used only for chart extraction
}

// PipelineOutput is the full result of one run, ready for the caller
// (handlers, via repository) to persist. Status follows the same enum as
// repository.Document.Status, redeclared here for the same layer-boundary
// reason documented in charts.go.
type PipelineOutput struct {
	Status         string // "complete" | "failed" | "verification_failed"
	SimplifiedText string
	ErrorMessage   string
	Verify         VerifyResult
	Charts         []Chart
}

// Pipeline status values. See charts.go for why these are redeclared
// locally instead of imported from repository.
const (
	pipelineStatusComplete           = "complete"
	pipelineStatusFailed             = "failed"
	pipelineStatusVerificationFailed = "verification_failed"
)

// stageTimeout bounds each pipeline stage's own work, layered on top of the
// per-call timeouts already enforced inside the Gemini and PDF clients
// (external/gemini.go, external/pdf.go). This is a second, coarser safety
// net — if a stage's internal timeout logic has a bug, the pipeline still
// won't hang the whole request forever.
const stageTimeout = 45 * time.Second

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
	// Stage 1: Simplify. (Extraction already happened before RunPipeline is
	// called — see handlers/documents.go — because extraction failure needs
	// to reject the upload before any Gemini call is made at all, per
	// ARCHITECTURE.md Failure Scenario 1. By the time we're here, we already
	// have OriginalText.)
	simplifyCtx, cancel := context.WithTimeout(ctx, stageTimeout)
	simplifiedText, err := Simplify(simplifyCtx, gemini, in.OriginalText, in.ReadingLevel)
	cancel()
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
	verifyCtx, cancel := context.WithTimeout(ctx, stageTimeout)
	verifyResult, err := DiffClaims(verifyCtx, gemini, in.OriginalText, simplifiedText)
	cancel()
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
	// Primary path: scan the full paper text for chart-worthy data — works
	// for ALL input types (PDF and pasted text), independent of embedded
	// images. Most academic PDFs render charts as vector graphics rather
	// than embedded image streams, so the old image-only path produced
	// zero results for real papers.
	//
	// Supplemental path: for PDFs that DO have embedded chart images,
	// run per-image data extraction on top of the text scan.
	var charts []Chart
	if in.OriginalText != "" {
		chartCtx, cancel := context.WithTimeout(ctx, stageTimeout)
		textCharts := ExtractChartsFromText(chartCtx, gemini, in.OriginalText)
		cancel()
		charts = textCharts
	}

	// Supplemental: image-based extraction (PDFs with embedded chart images).
	if in.SourceType == "pdf" && len(in.PDFBytes) > 0 {
		extracted, err := ExtractFromPDF(in.PDFBytes)
		if err == nil {
			pages := pageText(extracted.Pages)
			if pages == nil {
				pages = pageText{1: extracted.Text}
			}
			chartCtx, cancel := context.WithTimeout(ctx, stageTimeout)
			imageCharts := ReVisualizeCharts(chartCtx, gemini, extracted.Charts, pages)
			cancel()
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
		Status:         pipelineStatusComplete,
		SimplifiedText: simplifiedText,
		Verify:         verifyResult,
		Charts:         charts,
	}
}
