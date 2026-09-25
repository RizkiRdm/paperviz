package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"paperviz/internal/external"
	"paperviz/internal/models"
)

// ChartFailureCategory classifies chart processing errors for diagnostics.
type ChartFailureCategory string

const (
	FailureExtractionError     ChartFailureCategory = "EXTRACTION_ERROR"      // evidence/table extraction failed
	FailureDatasetError        ChartFailureCategory = "DATASET_ERROR"         // no candidate datasets built
	FailureChartSelectionError ChartFailureCategory = "CHART_SELECTION_ERROR" // LLM plan failed or invalid
	FailureGroundingError      ChartFailureCategory = "GROUNDING_ERROR"       // validation rejected chart
	FailureSchemaError         ChartFailureCategory = "SCHEMA_ERROR"          // JSON marshal/unmarshal failed
	FailureRenderError         ChartFailureCategory = "RENDER_ERROR"          // chart rendering failed
)

// logChartFailure emits a structured slog.Error with diagnostic fields for chart failures.
func logChartFailure(ctx context.Context, category ChartFailureCategory, stage, chapter, datasetID string, err error, fallbackUsed bool) {
	slog.ErrorContext(ctx, "chart failure",
		"error_category", string(category),
		"stage", stage,
		"chapter", chapter,
		"dataset_id", datasetID,
		"safe_error_message", err.Error(),
		"fallback_used", fallbackUsed,
	)
}

// Chart source-method values. Mirrors repository's CHECK-constrained enum
// (see migrations/001_init.sql) but is redeclared here rather than imported,
// because ARCHITECTURE.md's dependency direction is handlers -> services ->
// repository — services MUST NOT import repository, so these are the
// services-layer's own copy of the same three allowed strings. The pipeline
// (pipeline.go) is what maps these onto repository.Chart when persisting.
const (
	chartSourceDataExtracted = "data_extracted"
	chartSourceImageFallback = "image_fallback"
	chartSourceOmitted       = "omitted"
)

// ReVisualizeCharts attempts to reconstruct each detected chart as
// structured data first; if that fails for a given chart, it falls back to
// keeping the original image with a generated annotation; if both fail,
// the chart is marked "omitted" so the rest of the document is unaffected.
//
// Each chart is handled independently and a failure on one MUST NOT abort
// the others — this is ARCHITECTURE.md Failure Scenario 3 and 4 verbatim.
func ReVisualizeCharts(ctx context.Context, client *external.LLM, extracted []ExtractedChart, pages pageText) []Chart {
	results := make([]Chart, 0, len(extracted))

	for i, ec := range extracted {
		chart := reVisualizeOne(ctx, client, ec, pages[ec.PageNumber], i)
		results = append(results, chart)
	}

	return results
}

// reVisualizeOne runs the data-extraction-then-image-fallback-then-omit
// decision for a single chart. Split out from ReVisualizeCharts so each
// chart's error handling is isolated and testable on its own.
func reVisualizeOne(ctx context.Context, client *external.LLM, ec ExtractedChart, pageContext string, displayOrder int) Chart {
	base := Chart{
		PageNumber:   ec.PageNumber,
		DisplayOrder: displayOrder,
		SourceText:   pageContext,
		ChapterIndex: -1, // explicit: image-extracted charts are not chapter-linked
	}

	// Primary path: try to extract structured chart data from the page text.
	if strings.TrimSpace(pageContext) != "" {
		if dataJSON, ok := tryExtractChartData(ctx, client, pageContext); ok {
			base.SourceMethod = chartSourceDataExtracted
			base.ChartData = dataJSON
			return base
		}
	}

	// Fallback path: keep the original image, generate a plain-language
	// annotation from surrounding text.
	if len(ec.ImageBytes) > 0 {
		annotation, err := annotateImage(ctx, client, pageContext)
		if err == nil {
			base.SourceMethod = chartSourceImageFallback
			base.ImageBlob = ec.ImageBytes
			base.Annotation = annotation
			return base
		}
	}

	// Both paths failed — omit with an inline note (Failure Scenario 4).
	base.SourceMethod = chartSourceOmitted
	base.Annotation = fmt.Sprintf("Original chart could not be reprocessed — refer to source PDF page %d.", ec.PageNumber)
	return base
}

// tryExtractChartData calls Gemini to attempt structured data extraction
// for one page's text. Returns ok=false if the model reports no table
// present, or if the call/parse fails — either way, the caller falls back.
// The returned string is the raw validated JSON, ready to store as-is in
// charts.chart_data (ARCHITECTURE.md Section 3).
func tryExtractChartData(ctx context.Context, client *external.LLM, text string) (dataJSON string, ok bool) {
	prompt := fmt.Sprintf(chartDataExtractionPrompt, text)
	parsed, err := external.ExtractJSON[chartDataJSON](ctx, client, prompt, 0)
	if err != nil || len(parsed.Labels) == 0 || len(parsed.Values) == 0 {
		return "", false
	}

	b, err := parsed.marshalJSON()
	if err != nil {
		return "", false
	}

	return string(b), true
}

// annotateImage calls Gemini to write a short plain-language caption for a
// chart image, using surrounding page text as context.
func annotateImage(ctx context.Context, client *external.LLM, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no page context available for annotation")
	}
	prompt := fmt.Sprintf(imageAnnotationPrompt, text)
	return client.Generate(ctx, prompt, false, 0)
}

// findDatasetByID returns the dataset matching the given ID, or nil.
func findDatasetByID(datasets []models.CandidateDataset, id string) *models.CandidateDataset {
	for i := range datasets {
		if datasets[i].ID == id {
			return &datasets[i]
		}
	}
	return nil
}

// evidenceIDsFromDataset extracts non-empty EvidenceID strings from dataset points.
func evidenceIDsFromDataset(ds models.CandidateDataset) []string {
	var ids []string
	for _, p := range ds.Points {
		if p.EvidenceID != "" {
			ids = append(ids, p.EvidenceID)
		}
	}
	return ids
}
