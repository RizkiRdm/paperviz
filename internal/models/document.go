package models

// Document is the core domain entity representing an uploaded paper.
type Document struct {
	ID                      string
	CreatedAt               int64
	LastAccessedAt          int64
	Status                  string // processing | complete | failed | verification_failed
	SourceType              string // pdf | pasted_text
	ReadingLevel            string // simplified | eli5
	Title                   string
	OriginalText            string
	SimplifiedText          *string
	ErrorMessage            *string
	ChartExtractionDegraded bool
	ProcessingStage         *string
	UserID                  *string
	Saved                   bool
	Visibility              string // public | unlisted | private
	ShareToken              *string
	ProcessingTimeMs        *int
}

// Chapter represents a section of simplified text split from the document.
type Chapter struct {
	ID           string
	DocumentID   string
	Title        string
	Summary      string
	Excerpt      string
	DisplayOrder int
}

// Collection groups documents under a user-owned label.
type Collection struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt int64
}

// ClaimDiff stores the before/after claim comparison for verification.
type ClaimDiff struct {
	ID               string
	DocumentID       string
	OriginalClaims   string // JSON array
	SimplifiedClaims string // JSON array
	MismatchDetected bool
	MismatchDetail   *string
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
