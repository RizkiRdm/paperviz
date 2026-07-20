package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"paperviz/external"
)

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

// chartDataExtractionPrompt asks Gemini to find a data table in the raw
// extracted text and structure it as JSON labels/values. This is the
// PRIMARY path (ARCHITECTURE.md: "Primary path: extract underlying data
// points ... re-render as a new, annotated chart"). If no table is found,
// the model is instructed to return null so the caller can fall back.
const chartDataExtractionPrompt = `The text below is extracted from one page of an academic paper. If it
contains a data table (rows/columns of numbers with labels), extract it as
JSON in this exact shape:
{"labels": ["label1", "label2", ...], "values": [number1, number2, ...], "title": "short descriptive title"}

If the page text does NOT contain a clear data table, respond with exactly: null

Page text:
%s`

// imageAnnotationPrompt asks Gemini to write a plain-language caption for a
// chart image that couldn't be converted to structured data. This is the
// FALLBACK path (ARCHITECTURE.md: "extract the original chart image ...
// overlay a plain-language annotation"). We send a text description of the
// page context rather than the image itself, since Gemini's image-input
// path is out of scope for this MVP's direct-HTTP client (text-only prompts
// keep the Gemini client's request shape simple — see external/gemini.go).
const imageAnnotationPrompt = `The text below is extracted from a page of an academic paper that contains
a chart or figure. In 1-2 plain-language sentences, explain what this part
of the paper is likely showing and why it matters, based on the surrounding
text. If the text gives no useful context, say so plainly instead of
guessing.

Page text:
%s`

// chartDataJSON mirrors the JSON shape chartDataExtractionPrompt asks for.
type chartDataJSON struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
	Title  string    `json:"title"`
}

// pageText groups extracted body text by page number, used to give the
// chart-data and annotation prompts local context for a specific chart
// instead of the entire paper. Built by the pipeline from ExtractResult.
type pageText map[int]string

// ReVisualizeCharts attempts to reconstruct each detected chart as
// structured data first; if that fails for a given chart, it falls back to
// keeping the original image with a generated annotation; if both fail,
// the chart is marked "omitted" so the rest of the document is unaffected.
//
// Each chart is handled independently and a failure on one MUST NOT abort
// the others — this is ARCHITECTURE.md Failure Scenario 3 and 4 verbatim.
func ReVisualizeCharts(ctx context.Context, client *external.GeminiClient, extracted []ExtractedChart, pages pageText) []Chart {
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
func reVisualizeOne(ctx context.Context, client *external.GeminiClient, ec ExtractedChart, pageContext string, displayOrder int) Chart {
	base := Chart{
		PageNumber:   ec.PageNumber,
		DisplayOrder: displayOrder,
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
func tryExtractChartData(ctx context.Context, client *external.GeminiClient, text string) (dataJSON string, ok bool) {
	prompt := fmt.Sprintf(chartDataExtractionPrompt, text)
	raw, err := client.Generate(ctx, prompt, true)
	if err != nil {
		return "", false
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "null" {
		return "", false
	}

	var parsed chartDataJSON
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return "", false
	}
	if len(parsed.Labels) == 0 || len(parsed.Values) == 0 {
		return "", false
	}

	return trimmed, true
}

// annotateImage calls Gemini to write a short plain-language caption for a
// chart image, using surrounding page text as context.
func annotateImage(ctx context.Context, client *external.GeminiClient, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no page context available for annotation")
	}
	prompt := fmt.Sprintf(imageAnnotationPrompt, text)
	return client.Generate(ctx, prompt, false)
}
