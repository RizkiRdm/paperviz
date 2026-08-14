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
