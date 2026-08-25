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
