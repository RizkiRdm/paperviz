package repository

import "paperviz/internal/models"

// Domain types — thin mirrors of models for repo-layer access.
// Canonical definitions live in internal/models.
type Document = models.Document
type Chapter = models.Chapter
type Collection = models.Collection
type ClaimDiff = models.ClaimDiff

// Status enum values, per ARCHITECTURE.md Section 3 CHECK constraint.
const (
	StatusProcessing         = models.StatusProcessing
	StatusComplete           = models.StatusComplete
	StatusFailed             = models.StatusFailed
	StatusVerificationFailed = models.StatusVerificationFailed
)

// SourceType enum values.
const (
	SourceTypePDF        = models.SourceTypePDF
	SourceTypePastedText = models.SourceTypePastedText
)

// ReadingLevel enum values.
const (
	ReadingLevelSimplified = models.ReadingLevelSimplified
	ReadingLevelELI5       = models.ReadingLevelELI5
)

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

// CollectionListItem is a lightweight row for the collections list.
type CollectionListItem struct {
	ID            string
	Name          string
	CreatedAt     int64
	DocumentCount int
}

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

// ClaimEvidence mirrors the claim_evidence table.
type ClaimEvidence struct {
	ID               string
	ClaimID          string
	EvidenceID       string
	RelationshipType string // supports | contradicts | clarifies
	CreatedAt        int64
}

// PaperRelationship mirrors the paper_relationships table.
type PaperRelationship struct {
	ID               string
	SourcePaperID    string
	TargetPaperID    string
	RelationshipType string // supporting | contradicting | citing | similar_methodology | different_findings
	EvidenceText     *string
	CreatedAt        int64
}

// ClaimEvidenceRelationshipType enum values.
const (
	ClaimEvidenceRelationshipSupports    = "supports"
	ClaimEvidenceRelationshipContradicts = "contradicts"
	ClaimEvidenceRelationshipClarifies   = "clarifies"
)

// PaperRelationshipType enum values.
const (
	PaperRelationshipSupporting         = "supporting"
	PaperRelationshipContradicting      = "contradicting"
	PaperRelationshipCiting             = "citing"
	PaperRelationshipSimilarMethodology = "similar_methodology"
	PaperRelationshipDifferentFindings  = "different_findings"
)

// ClaimType enum values.
const (
	ClaimTypeHypothesis = "hypothesis"
	ClaimTypeFinding    = "finding"
	ClaimTypeConclusion = "conclusion"
	ClaimTypeMethod     = "method"
	ClaimTypeLimitation = "limitation"
)

// Confidence enum values.
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// MethodType enum values.
const (
	MethodTypeExperimental  = "experimental"
	MethodTypeSurvey        = "survey"
	MethodTypeQualitative   = "qualitative"
	MethodTypeQuantitative  = "quantitative"
	MethodTypeComputational = "computational"
	MethodTypeOther         = "other"
)

// ResultType enum values.
const (
	ResultTypePrimary   = "primary"
	ResultTypeSecondary = "secondary"
	ResultTypeNegative  = "negative"
)
