package services

import "errors"

// ErrNoTextLayer and ErrInvalidReadingLevel are sentinel errors surfaced to
// handlers, which map them to snake_case error codes without leaking
// implementation detail (ARCHITECTURE.md Internal Contracts).
var (
	ErrNoTextLayer         = errors.New("pdf has no extractable text layer")
	ErrInvalidReadingLevel = errors.New("reading level must be simplified or eli5")
)

// ExtractedChart is a chart/table candidate found during PDF extraction,
// before the re-visualization service decides data vs. image-fallback path.
// Page-local text context for the chart comes from ExtractResult.Pages
// (keyed by PageNumber), not stored on this struct directly.
type ExtractedChart struct {
	PageNumber int
	ImageBytes []byte // raw chart image bytes, if extracted
}

// ExtractResult is the output of the Extraction Service.
type ExtractResult struct {
	Text   string
	Pages  map[int]string   // 1-indexed page number -> that page's text; used for per-chart context
	Charts []ExtractedChart // empty for pasted_text source
}

// Chart is a fully processed chart ready for persistence.
type Chart struct {
	SourceMethod string // data_extracted | image_fallback | omitted
	ChartData    string // JSON, empty if not data_extracted
	ImageBlob    []byte
	Annotation   string
	SourceText   string // original page text backing this chart; empty when none available
	PageNumber   int
	DisplayOrder int
	ChapterIndex int // index in chapters array; -1 if not linked to a chapter
}

// Chapter is one detected section of the paper, used to drive per-chapter
// chart generation. Detected from SIMPLIFIED text — chapters match what
// the reader sees, not the original paper's raw section boundaries.
type Chapter struct {
	Title   string
	Summary string
	Excerpt string
}

// VerifyResult is the output of the Claim-Diff Verification Service.
type VerifyResult struct {
	OriginalClaims   []string
	SimplifiedClaims []string
	MismatchDetected bool
	MismatchDetail   string
}

// PipelineOutput is the full result of one run, ready for the caller
// (handlers, via repository) to persist. Status follows the same enum as
// repository.Document.Status, redeclared here for the same layer-boundary
// reason documented in charts.go.
type PipelineOutput struct {
	Status                  string // "complete" | "failed" | "verification_failed"
	SimplifiedText          string
	ErrorMessage            string
	Verify                  VerifyResult
	Charts                  []Chart
	ChartExtractionDegraded bool
	Chapters                []Chapter
}

// PaperSummary holds the extracted structured fields for one paper,
// used as input to the multi-paper comparison service.
type PaperSummary struct {
	DocumentID       string
	Title            string
	ResearchQuestion string
	Methodology      string
	Dataset          string
	SampleSize       string
	Findings         []string
	Limitations      []string
	Figures          []string
	Evidence         []string
	Conclusions      string
}

// ComparisonDimension represents a single dimension of comparison across papers.
type ComparisonDimension struct {
	Dimension string            // e.g. "research_question", "methodology", "findings"
	Values    map[string]string // document_id -> value for that paper
	Notes     string            // optional synthesis/observation
}

// EvidenceClaim represents a cross-paper claim with per-paper stance.
type EvidenceClaim struct {
	Claim      string            `json:"claim"`
	Stances    map[string]string `json:"stances"`     // document_id → "supporting" | "contradicting" | "unclear"
	SourceRefs map[string]string `json:"source_refs"` // document_id → evidence reference text
}

// PaperComparison holds the full structured comparison across multiple papers.
type PaperComparison struct {
	Papers         []PaperSummary        // individual paper summaries
	Dimensions     []ComparisonDimension // side-by-side comparison dimensions
	Agreement      []string              // areas where papers agree
	Disagreement   []string              // areas where papers disagree
	EvidenceClaims []EvidenceClaim       // cross-paper claims with per-paper stance
}
