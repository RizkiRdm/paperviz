package services

import (
	"context"

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
// PDFDoc carries pre-parsed PDF text/pages/charts when caller already
// parsed PDF (intake stage); if nil and PDFBytes present, pipeline parses
// pages/images once without re-extracting flattened text.
type PipelineInput struct {
	OriginalText string
	SourceType   string             // "pdf" | "pasted_text"
	ReadingLevel string             // "simplified" | "eli5"
	PDFBytes     []byte             // nil for pasted_text; used only for chart extraction
	PDFDoc       *PDFDocument       // optional pre-parsed PDF; avoids duplicate ParsePDF
	OnStage      func(stage string) // called at each pipeline stage transition
}

// Pipeline status values. See charts.go for why these are redeclared
// locally instead of imported from repository.
const (
	pipelineStatusComplete           = "complete"
	pipelineStatusFailed             = "failed"
	pipelineStatusVerificationFailed = "verification_failed"
)

// RunPipeline is canonical sequential pipeline: simplification → claim-diff
// verification → chapter chart generation → image chart fallback → return
// result for caller persistence. Business logic lives in pipeline_stages.go.
func RunPipeline(ctx context.Context, gemini *external.GeminiClient, in PipelineInput) PipelineOutput {
	emit := func(s string) {
		if in.OnStage != nil {
			in.OnStage(s)
		}
	}

	// Stage 1: Simplify original text at target reading level.
	simplifiedText, failOut := runSimplifyStage(ctx, gemini, in.OriginalText, in.ReadingLevel, emit)
	if failOut != nil {
		return *failOut
	}

	// Stage 2: Verify simplified text preserves original claims.
	// Runs BEFORE chart processing — mismatched docs don't reach charts.
	verifyResult, failOut := runVerifyStage(ctx, gemini, in.OriginalText, simplifiedText, emit)
	if failOut != nil {
		return *failOut
	}

	// Stage 3: Chapter detection + per-chapter charts + image fallback.
	charts, chartDegraded, chapters := runFiguresStage(ctx, gemini, simplifiedText, in, emit)

	return PipelineOutput{
		Status:                  pipelineStatusComplete,
		SimplifiedText:          simplifiedText,
		Verify:                  verifyResult,
		Charts:                  charts,
		ChartExtractionDegraded: chartDegraded,
		Chapters:                chapters,
	}
}
