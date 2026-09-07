package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
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

const chapterChartPlanPrompt = `You are deciding whether this chapter of a paper contains data worth
visualizing as a chart. Candidate datasets extracted from the chapter text
are provided below as JSON. Choose the BEST dataset (or none) and specify
how to visualize it.

Chapter title: %s
Chapter summary: %s
Chapter text:
%s

Candidate datasets:
%s

Rules:
- If none of the datasets are meaningful for visualization, set has_chart false.
- Pick the dataset_id of the best dataset if has_chart is true.
- Choose chart_type that fits the data shape:
  "bar" for categorical comparison, "line" for trend over time,
  "pie" for parts-of-a-whole, "scatter" for two-variable relationship.
- Provide title, x_label, y_label, and a one-sentence takeaway.
- Do NOT invent or modify numeric values — the values come from the dataset.

Respond with ONLY JSON:
{
  "has_chart": true,
  "dataset_id": "ds_accuracy__",
  "chart_type": "bar",
  "title": "short descriptive title",
  "x_label": "horizontal axis label",
  "y_label": "vertical axis label",
  "takeaway": "one sentence — key finding"
}

If no chart is warranted:
{"has_chart": false}

Do not return any explanatory text outside of this JSON.`

// chapterChartPlan is the LLM's chart planning response (no numeric values).
type chapterChartPlan struct {
	HasChart  bool   `json:"has_chart"`
	DatasetID string `json:"dataset_id"`
	ChartType string `json:"chart_type"`
	Title     string `json:"title"`
	XLabel    string `json:"x_label,omitempty"`
	YLabel    string `json:"y_label,omitempty"`
	Takeaway  string `json:"takeaway,omitempty"`
}

// perDatasetChartPlanPrompt asks the LLM to evaluate a single candidate
// dataset in the context of its chapter and decide whether it warrants a chart.
const perDatasetChartPrompt = `You are deciding whether ONE specific dataset from a paper chapter is worth
visualizing as a chart. Evaluate this dataset in the context of the chapter.

Chapter title: %s
Chapter summary: %s

Dataset to evaluate:
%s

Rules:
- If the dataset has fewer than 2 data points, set has_chart false.
- If the values are trivial, repetitive, or not meaningful for visualization, set has_chart false.
- If the dataset IS worth charting, set has_chart true and choose chart_type:
  "bar" for categorical comparison, "line" for trend over time,
  "pie" for parts-of-a-whole, "scatter" for two-variable relationship.
- Provide a short descriptive title, x_label, y_label, and a one-sentence takeaway.
- Do NOT invent or modify numeric values — the values come from the dataset.

Respond with ONLY JSON:
{
  "has_chart": true,
  "chart_type": "bar",
  "title": "short descriptive title",
  "x_label": "horizontal axis label",
  "y_label": "vertical axis label",
  "takeaway": "one sentence — key finding"
}

If no chart is warranted:
{"has_chart": false}

Do not return any explanatory text outside of this JSON.`

// perDatasetPlan is the LLM's per-dataset chart planning response.
type perDatasetPlan struct {
	HasChart  bool   `json:"has_chart"`
	ChartType string `json:"chart_type"`
	Title     string `json:"title"`
	XLabel    string `json:"x_label,omitempty"`
	YLabel    string `json:"y_label,omitempty"`
	Takeaway  string `json:"takeaway,omitempty"`
}

// GenerateChapterCharts extracts numeric evidence from a chapter, builds
// candidate datasets, then asks the LLM to evaluate each dataset for charting.
func GenerateChapterCharts(ctx context.Context, client *external.GeminiClient, chapter Chapter, displayOrder int) (charts []Chart, degraded bool) {
	evidence := ExtractNumericEvidence(chapter.Excerpt, 1)
	datasets := BuildCandidateDatasets(evidence)

	if len(datasets) == 0 {
		slog.Info("chapter chart: no evidence extracted", "stage", "chart", "chapter", chapter.Title)
		return nil, false
	}

	if client == nil {
		slog.Warn("chapter chart: nil Gemini client, skipping", "stage", "chart", "chapter", chapter.Title)
		return nil, true
	}

	validTypes := map[string]bool{"bar": true, "line": true, "pie": true, "scatter": true}
	nextOrder := displayOrder

	for _, ds := range datasets {
		datasetJSON, err := json.MarshalIndent(ds, "", "  ")
		if err != nil {
			logChartFailure(ctx, FailureSchemaError, "dataset_marshal", chapter.Title, ds.ID, err, true)
			degraded = true
			continue
		}

		prompt := fmt.Sprintf(perDatasetChartPrompt, chapter.Title, chapter.Summary, string(datasetJSON))
		plan, err := external.ExtractJSON[perDatasetPlan](ctx, client, prompt, 0)
		if err != nil {
			logChartFailure(ctx, FailureChartSelectionError, "chart_plan", chapter.Title, ds.ID, err, true)
			degraded = true
			continue
		}

		if !plan.HasChart {
			slog.Info("chapter chart: no chart warranted for dataset", "stage", "chart", "chapter", chapter.Title, "dataset_id", ds.ID)
			continue
		}

		if !validTypes[plan.ChartType] {
			plan.ChartType = "bar"
		}

		chartData := chartDataJSON{
			Labels: ds.Labels(),
			Values: ds.Values(),
			Title:  plan.Title,
		}
		dataRaw, err := json.Marshal(chartData)
		if err != nil {
			logChartFailure(ctx, FailureSchemaError, "chart_data_marshal", chapter.Title, ds.ID, err, true)
			degraded = true
			continue
		}

		slog.Info("chapter chart generated",
			"stage", "chart",
			"chapter", chapter.Title,
			"chart_type", plan.ChartType,
			"dataset_id", ds.ID,
		)

		provenance := models.ChartProvenance{
			EvidenceIDs:     evidenceIDsFromDataset(ds),
			DatasetID:       ds.ID,
			SourceMethod:    "text_evidence",
			GroundingStatus: "verified",
		}

		charts = append(charts, Chart{
			SourceMethod: chartSourceDataExtracted,
			ChartData:    string(dataRaw),
			Annotation:   fmt.Sprintf("From chapter: %s", chapter.Title),
			DisplayOrder: nextOrder,
			ChapterIndex: nextOrder,
			Provenance:   provenance,
		})
		nextOrder++
	}

	return charts, degraded
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
