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
func tryExtractChartData(ctx context.Context, client *external.GeminiClient, text string) (dataJSON string, ok bool) {
	prompt := fmt.Sprintf(chartDataExtractionPrompt, text)
	parsed, err := external.ExtractJSON[chartDataJSON](ctx, client, prompt, 0)
	if err != nil || len(parsed.Labels) == 0 || len(parsed.Values) == 0 {
		return "", false
	}

	b, err := json.Marshal(parsed)
	if err != nil {
		return "", false
	}

	return string(b), true
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

const chapterChartPrompt = `You are deciding whether this chapter of a paper contains data worth
visualizing as a chart, and if so, producing that chart.

Chapter title: %s
Chapter summary: %s
Chapter text:
%s

First, decide: does this chapter contain a SPECIFIC, meaningful set of
numbers worth a reader seeing as a chart (comparisons, trends over time,
proportions, before/after results, multiple measured values)? A chapter
that only mentions numbers in passing, with no real comparison or pattern,
does NOT qualify — do not force a chart out of weak material.

If it does NOT qualify, respond with ONLY:
{"has_chart": false}

If it DOES qualify, choose the chart_type that best fits the shape of the
data, using this rule:
- "bar": comparing distinct discrete categories or groups against each other
- "line": a trend over time, steps, or an ordered sequence
- "pie": parts of a whole that sum to ~100%% or a fixed total
- "scatter": relationship between two independent numeric variables

Then provide:
- x_axis: label for the horizontal axis (e.g. "Model", "Time (months)", "Treatment Group")
- y_axis: label for the vertical axis (e.g. "Accuracy (%%)", "Revenue ($M)", "Sample Size")
- key_takeaway: ONE sentence stating the most important finding this chart reveals
- limitations: ONE sentence noting what the chart does NOT show or what caveats apply
- confidence: "high" if the data is clearly stated and unambiguous,
  "medium" if some interpretation was needed,
  "low" if the numbers are vague, incomplete, or you are uncertain about the extraction

Respond with ONLY JSON in this exact shape:
{
  "has_chart": true,
  "chart_type": "bar" | "line" | "pie" | "scatter",
  "title": "short descriptive title tied to this chapter",
  "labels": ["label1", "label2", ...],
  "values": [number1, number2, ...],
  "x_axis": "horizontal axis label",
  "y_axis": "vertical axis label",
  "key_takeaway": "one sentence — most important finding",
  "limitations": "one sentence — what the chart does NOT show",
  "confidence": "high" | "medium" | "low"
}

Do not return any explanatory text outside of this JSON.`

type chapterChartJSON struct {
	HasChart    bool        `json:"has_chart"`
	ChartType   string      `json:"chart_type"`
	Title       string      `json:"title"`
	Labels      []string    `json:"labels"`
	Values      chartValues `json:"values"`
	XAxis       string      `json:"x_axis,omitempty"`
	YAxis       string      `json:"y_axis,omitempty"`
	KeyTakeaway string      `json:"key_takeaway,omitempty"`
	Limitations string      `json:"limitations,omitempty"`
	Confidence  string      `json:"confidence,omitempty"`
}

func GenerateChapterChart(ctx context.Context, client *external.GeminiClient, chapter Chapter, displayOrder int) (chart Chart, ok bool, degraded bool) {
	prompt := fmt.Sprintf(chapterChartPrompt, chapter.Title, chapter.Summary, chapter.Excerpt)
	parsed, err := external.ExtractJSON[chapterChartJSON](ctx, client, prompt, 0)
	if err != nil {
		slog.Error("chapter chart generation failed", "stage", "chart", "chapter", chapter.Title, "error", err)
		return Chart{}, false, true
	}

	if !parsed.HasChart || len(parsed.Labels) == 0 || len(parsed.Values) == 0 {
		slog.Info("chapter chart: no chart warranted", "stage", "chart", "chapter", chapter.Title)
		return Chart{}, false, false
	}

	validTypes := map[string]bool{"bar": true, "line": true, "pie": true, "scatter": true}
	if !validTypes[parsed.ChartType] {
		parsed.ChartType = "bar"
	}

	dataRaw, err := json.Marshal(parsed)
	if err != nil {
		return Chart{}, false, true
	}

	slog.Info("chapter chart generated",
		"stage", "chart",
		"chapter", chapter.Title,
		"chart_type", parsed.ChartType,
	)

	return Chart{
		SourceMethod: chartSourceDataExtracted,
		ChartData:    string(dataRaw),
		Annotation:   fmt.Sprintf("From chapter: %s", chapter.Title),
		DisplayOrder: displayOrder,
		ChapterIndex: displayOrder,
	}, true, false
}
