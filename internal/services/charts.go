package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"paperviz/internal/external"
)

// chartValues is a lenient JSON unmarshaler for number arrays. Gemini
// flash-lite sometimes returns `"values":"72, 89"` (string) instead of
// `"values":[72, 89]` (number array). This type handles both forms:
// standard []float64, or comma/space-separated numbers with optional %.
type chartValues []float64

func (v *chartValues) UnmarshalJSON(b []byte) error {
	var nums []float64
	if err := json.Unmarshal(b, &nums); err == nil {
		*v = nums
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	cleaned := strings.NewReplacer("%", "", ",", " ").Replace(s)
	fields := strings.Fields(cleaned)
	for _, f := range fields {
		n, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return fmt.Errorf("chartValues: cannot parse %q: %w", f, err)
		}
		nums = append(nums, n)
	}
	*v = nums
	return nil
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

// chartDataExtractionPrompt asks Gemini to determine whether the page text
// contains data suitable for visualization (numbers, comparisons, trends,
// percentages, or tables), and if so, to return it as structured JSON with a
// recommended chart type. This is the PRIMARY path (ARCHITECTURE.md: "Primary
// path: extract underlying data points ... re-render as a new, annotated
// chart"). When no chart-worthy data is found the model returns has_chart
// false, so the caller can fall back to the image-annotation path.
const chartDataExtractionPrompt = `Based on the content of the following page text from an academic paper,
determine whether there is any data suitable for visualization as a chart
(numbers, comparisons, trends, percentages, or experimental results in the
form of tables/numbers).

If there IS suitable data, return ONLY JSON in the following format:
{
  "has_chart": true,
  "chart_type": "bar",
  "title": "short descriptive title",
  "labels": ["label1", "label2", ...],
  "values": [number1, number2, ...]
}

chart_type must be one of: "bar", "line", "pie", "scatter".

If there is NO data suitable for visualization, return ONLY:
{"has_chart": false}

Do not return any explanatory text outside of this JSON. Use the specified format exactly.

Page text:
%s`

// fullTextChartPrompt scans the entire paper text for numerical data —
// tables, percentages, statistics, comparisons, scores, counts — even when
// embedded in prose (not just in tabular format). The prompt uses explicit
// prose-to-chart examples so the model extracts data from sentences like
// "accuracy improved from 72% to 89%" rather than skipping them. Includes
// academic-statistics patterns (F-tests, means, p-values, regression
// coefficients) that real papers use.
const fullTextChartPrompt = `You are a data extraction engine. Scan the following academic paper text
and extract EVERY numerical data point, comparison, percentage, and
statistical result. Format each dataset as a chart object.

Examples of prose data to extract:
- "accuracy improved from 72%% to 89%%" -> labels=["Before","After"], values=[72,89]
- "Model A scored 85, Model B scored 92" -> labels=["A","B"], values=[85,92]
- "25%% chose X, 75%% chose Y" -> labels=["X","Y"], values=[25,75]
- "F(1,24) = 4.82, p < 0.05" -> labels=["F-value"], values=[4.82]
- "M = 3.42, SD = 0.51 for control; M = 4.18, SD = 0.47 for treatment" -> labels=["Control","Treatment"], values=[3.42,4.18]
- "beta = -0.23, SE = 0.08, p = 0.003" -> labels=["beta"], values=[-0.23]
- "Group A: 85%%, Group B: 92%%, Group C: 78%%" -> labels=["A","B","C"], values=[85,92,78]
- Tables with rows and columns -> labels are row/column headers, values are numbers
- Any comparison of two or more numbers, even if scattered across sentences

Return ONLY a JSON array. Each element must be:
{
  "chart_type": "bar",
  "title": "short descriptive title",
  "labels": ["label1", "label2", ...],
  "values": [number1, number2, ...]
}

chart_type must be one of: "bar", "line", "pie", "scatter".

If no numerical data exists at all, return [].

Paper text:
%s`

// textChartElem is the per-element struct for parsing the JSON array returned
// by the fullTextChartPrompt. Unexported — only used inside this file for
// validation before constructing Chart values.
type textChartElem struct {
	ChartType string     `json:"chart_type"`
	Labels    []string   `json:"labels"`
	Values    chartValues `json:"values"`
	Title     string     `json:"title"`
}

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
	Labels []string    `json:"labels"`
	Values chartValues `json:"values"`
	Title  string      `json:"title"`
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
	raw, err := client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return "", false
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "null" {
		return "", false
	}

	// Strip markdown code fences — smaller models sometimes wrap JSON in
	// ```json … ``` blocks that would cause json.Unmarshal to fail.
	trimmed = stripJSONFences(trimmed)

	var parsed chartDataJSON
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		slog.Error("per-chart JSON parse failed",
			"stage", "chart",
			"error", err,
			"page_text_bytes", len(text),
		)
		return "", false
	}
	if len(parsed.Labels) == 0 || len(parsed.Values) == 0 {
		return "", false
	}

	return trimmed, true
}

// ExtractChartsFromText scans the full paper text for chart-worthy data
// (tables, stats, numerical comparisons) using Gemini and returns a slice
// of Chart structs. Always runs, independent of PDF image extraction.
// Returns nil when no chart-worthy data exists or the AI call fails.
func ExtractChartsFromText(ctx context.Context, client *external.GeminiClient, text string) []Chart {
	prompt := fmt.Sprintf(fullTextChartPrompt, text)
	raw, err := client.Generate(ctx, prompt, true, 2048)
	if err != nil {
		slog.Error("text chart extraction failed", "stage", "chart", "error", err)
		return nil
	}

	trimmed := strings.TrimSpace(raw)
	trimmed = stripJSONFences(trimmed)
	snippet := trimmed
	if len(snippet) > 300 {
		snippet = snippet[:300]
	}

	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		slog.Info("text chart extraction: no chart-worthy data found by model",
			"stage", "chart",
			"source", "textscan",
			"response_snippet", snippet,
		)
		return nil
	}

	var parsed []textChartElem
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		slog.Error("text chart JSON parse failed",
			"stage", "chart",
			"error", err,
			"response_snippet", snippet,
		)
		return nil
	}

	if len(parsed) == 0 {
		slog.Info("text chart extraction: empty array after parse",
			"stage", "chart",
			"source", "textscan",
			"response_snippet", snippet,
		)
		return nil
	}

	charts := make([]Chart, 0, len(parsed))
	for i, elem := range parsed {
		if len(elem.Labels) == 0 || len(elem.Values) == 0 {
			continue
		}
		dataRaw, err := json.Marshal(elem)
		if err != nil {
			continue
		}
		charts = append(charts, Chart{
			SourceMethod: chartSourceDataExtracted,
			ChartData:    string(dataRaw),
			DisplayOrder: i,
		})
	}

	if len(charts) == 0 {
		return nil
	}
	slog.Info("text chart extraction complete",
		"stage", "chart",
		"charts_count", len(charts),
		"gemini_returned", len(parsed),
		"source", "textscan",
	)
	return charts
}

// stripJSONFences removes markdown code fences (```json … ``` or ``` … ```)
// surrounding a JSON string, which smaller LLMs sometimes add even when
// instructed not to.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	// Skip the opening fence line (```json or ```)
	nl := strings.Index(s, "\n")
	if nl < 0 {
		return s
	}
	s = s[nl+1:]
	// Remove closing fence if present
	if idx := strings.LastIndex(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

// annotateImage calls Gemini to write a short plain-language caption for a
// chart image, using surrounding page text as context.
func annotateImage(ctx context.Context, client *external.GeminiClient, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no page context available for annotation")
	}
	prompt := fmt.Sprintf(imageAnnotationPrompt, text)
	return client.Generate(ctx, prompt, false, 0)
}
