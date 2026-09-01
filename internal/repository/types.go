package repository

// Document mirrors the documents table (ARCHITECTURE.md Section 3).
type Document struct {
	ID                      string
	CreatedAt               int64
	LastAccessedAt          int64
	Status                  string // processing | complete | failed | verification_failed
	SourceType              string // pdf | pasted_text
	ReadingLevel            string // simplified | eli5
	Title                   string // paper title derived from first line of text
	OriginalText            string
	SimplifiedText          *string
	ErrorMessage            *string
	ChartExtractionDegraded bool
	ProcessingStage         *string
	UserID                  *string // nullable — anonymous docs have NULL user_id
	Saved                   bool
	Visibility              string  // public | unlisted | private
	ShareToken              *string // nullable — lazy-generated for paper share pages
	ProcessingTimeMs        *int    // nullable — total pipeline processing time in milliseconds (Chunk 6.1)
}

// DocumentListItem is a lightweight row for the paper-history list.
// Carries preview + counts instead of full text so the list stays cheap.
type DocumentListItem struct {
	ID               string
	Title            string
	CreatedAt        int64
	Status           string
	SummaryPreview   string
	ChartCount       int
	ExplanationCount int
	Saved            bool
}

// Chart mirrors the charts table.
type Chart struct {
	ID           string
	DocumentID   string
	SourceMethod string // data_extracted | image_fallback | omitted
	ChartData    *string
	ImageBlob    []byte
	Annotation   *string
	PageNumber   *int
	DisplayOrder int
	ChapterID    *string // nullable — links chart to a chapter
	ShareToken   *string // nullable — lazy-generated for public share pages
}

// ClaimDiff mirrors the claim_diffs table.
type ClaimDiff struct {
	ID               string
	DocumentID       string
	OriginalClaims   string // JSON array
	SimplifiedClaims string // JSON array
	MismatchDetected bool
	MismatchDetail   *string
}

// Chapter mirrors the chapters table.
type Chapter struct {
	ID           string
	DocumentID   string
	Title        string
	Summary      string
	Excerpt      string
	DisplayOrder int
}

// Evidence mirrors the evidence table.
type Evidence struct {
	ID              string
	PaperID         string
	Page            *int
	FigureID        *string
	TableID         *string
	Section         *string
	SourceText      string
	SourceReference string
}

// Collection mirrors the collections table.
type Collection struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt int64
}

// CollectionListItem is a lightweight row for the collections list.
type CollectionListItem struct {
	ID            string
	Name          string
	CreatedAt     int64
	DocumentCount int
}

// Status enum values, per ARCHITECTURE.md Section 3 CHECK constraint.
const (
	StatusProcessing         = "processing"
	StatusComplete           = "complete"
	StatusFailed             = "failed"
	StatusVerificationFailed = "verification_failed"
)

// SourceType enum values.
const (
	SourceTypePDF        = "pdf"
	SourceTypePastedText = "pasted_text"
)

// ReadingLevel enum values.
const (
	ReadingLevelSimplified = "simplified"
	ReadingLevelELI5       = "eli5"
)

// Chart SourceMethod enum values.
const (
	ChartSourceDataExtracted = "data_extracted"
	ChartSourceImageFallback = "image_fallback"
	ChartSourceOmitted       = "omitted"
)

// Claim mirrors the claims table.
type Claim struct {
	ID         string
	PaperID    string
	ClaimText  string
	ClaimType  string
	Confidence string
	SourcePage *int
	SourceText *string
	CreatedAt  int64
}

// PaperTable mirrors the paper_tables table.
type PaperTable struct {
	ID           string
	DocumentID   string
	PageNumber   *int
	Caption      *string
	Headers      string // JSON array of column headers
	Rows         string // JSON array of row arrays
	SourceText   *string
	DisplayOrder int
}

// Method mirrors the methods table.
type Method struct {
	ID          string
	PaperID     string
	MethodName  string
	Description *string
	MethodType  string
	SourcePage  *int
	SourceText  *string
}

// Result mirrors the results table.
type Result struct {
	ID                   string
	PaperID              string
	ResultText           string
	ResultType           string
	SupportingEvidenceID *string
	SourcePage           *int
	SourceText           *string
}

// Citation mirrors the citations table.
type Citation struct {
	ID           string
	PaperID      string
	CitedPaperID *string
	Authors      *string
	Title        *string
	Year         *int
	Venue        *string
	DOI          *string
	URL          *string
	SourcePage   *int
}

// ClaimType enum values.
const (
	ClaimTypeHypothesis  = "hypothesis"
	ClaimTypeFinding     = "finding"
	ClaimTypeConclusion  = "conclusion"
	ClaimTypeMethod      = "method"
	ClaimTypeLimitation  = "limitation"
)

// Confidence enum values.
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// MethodType enum values.
const (
	MethodTypeExperimental   = "experimental"
	MethodTypeSurvey         = "survey"
	MethodTypeQualitative    = "qualitative"
	MethodTypeQuantitative   = "quantitative"
	MethodTypeComputational  = "computational"
	MethodTypeOther          = "other"
)

// ResultType enum values.
const (
	ResultTypePrimary   = "primary"
	ResultTypeSecondary = "secondary"
	ResultTypeNegative  = "negative"
)
