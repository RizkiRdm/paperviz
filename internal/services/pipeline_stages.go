package services

import (
	"context"
	"log/slog"

	"paperviz/internal/external"
	"paperviz/internal/infra/ratelimit"
)

// runSimplifyStage runs simplification or returns failure output.
// emit notifies stage transitions for persistence.
func runSimplifyStage(ctx context.Context, gemini *external.LLM, originalText, readingLevel string, emit func(string)) (string, *PipelineOutput) {
	emit("simplifying")
	simplifiedText, err := Simplify(ctx, gemini, originalText, readingLevel)
	if err != nil {
		slog.Error("pipeline stage failed", "stage", "simplify", "error", err)
		return "", &PipelineOutput{Status: pipelineStatusFailed, ErrorMessage: "simplification_failed"}
	}
	return simplifiedText, nil
}

// runVerifyStage runs claim-diff with rate-limit pause or returns failure.
// emit notifies stage transitions for persistence.
func runVerifyStage(ctx context.Context, gemini *external.LLM, originalText, simplifiedText string, emit func(string)) (VerifyResult, *PipelineOutput) {
	ratelimit.WaitRateLimit(ctx, ratelimit.DefaultDelay)
	emit("verifying")
	verifyResult, err := DiffClaims(ctx, gemini, originalText, simplifiedText)
	if err != nil {
		slog.Error("pipeline stage failed", "stage", "verify", "error", err)
		return VerifyResult{}, &PipelineOutput{Status: pipelineStatusFailed, ErrorMessage: "verification_failed_to_run"}
	}
	if verifyResult.MismatchDetected {
		return verifyResult, &PipelineOutput{Status: pipelineStatusVerificationFailed, SimplifiedText: simplifiedText, Verify: verifyResult}
	}
	return verifyResult, nil
}

// runFiguresStage runs chapter detection, per-chapter chart generation and image fallback.
func runFiguresStage(ctx context.Context, gemini *external.LLM, simplifiedText string, in PipelineInput, emit func(string)) ([]Chart, bool, []Chapter) {
	ratelimit.WaitRateLimit(ctx, ratelimit.DefaultDelay)
	emit("generating_charts")
	var charts []Chart
	var chartDegraded bool
	chapters, err := DetectChapters(ctx, gemini, simplifiedText)
	if err != nil {
		slog.Error("pipeline: chapter detection failed", "stage", "chapters", "error", err)
		chartDegraded = true
	} else if len(chapters) == 0 {
		slog.Info("pipeline: no chapters detected, skipping chart generation", "stage", "chapters")
	} else {
		for _, chapter := range chapters {
			chapterCharts, degraded := GenerateChapterCharts(ctx, gemini, chapter, len(charts))
			if degraded {
				chartDegraded = true
			}
			charts = append(charts, chapterCharts...)
		}
		slog.Info("pipeline: chapter-based chart generation complete", "stage", "chart", "chapters_detected", len(chapters), "charts_generated", len(charts))
	}
	policy := GetSourcePolicy(in.SourceType)
	if policy.ImageChartsAllowed() {
		src, pages := resolveImageSources(in)
		if len(src) > 0 || pages != nil {
			imageCharts := ReVisualizeCharts(ctx, gemini, src, pages)
			offset := len(charts)
			for i := range imageCharts {
				imageCharts[i].DisplayOrder = offset + i
			}
			charts = append(charts, imageCharts...)
		}
	}
	return charts, chartDegraded, chapters
}

// resolveImageSources builds image chart sources and page map from intake.
func resolveImageSources(in PipelineInput) ([]ExtractedChart, pageText) {
	var pages pageText
	var src []ExtractedChart
	if in.PDFDoc != nil {
		pages = pageText(in.PDFDoc.Pages)
		src = in.PDFDoc.Charts
		if pages == nil && in.PDFDoc.Text != "" {
			pages = pageText{1: in.PDFDoc.Text}
		}
		return src, pages
	}
	if len(in.PDFBytes) == 0 {
		return nil, nil
	}
	images, imgErr := external.ExtractImages(in.PDFBytes)
	if imgErr != nil {
		slog.Error("chart image extraction failed", "stage", "chart", "error", imgErr)
	}
	pg, pgErr := external.ExtractTextByPage(in.PDFBytes)
	if pgErr != nil {
		slog.Error("chart per-page text extraction failed", "stage", "chart", "error", pgErr)
	} else {
		pages = pageText(pg)
	}
	if pages == nil && in.OriginalText != "" {
		pages = pageText{1: in.OriginalText}
	}
	src = make([]ExtractedChart, 0, len(images))
	for _, img := range images {
		src = append(src, ExtractedChart{PageNumber: img.PageNumber, ImageBytes: img.Bytes})
	}
	if len(src) > maxImageChartsPerDocument {
		src = src[:maxImageChartsPerDocument]
	}
	return src, pages
}
