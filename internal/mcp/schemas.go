package mcp

import "encoding/json"

// --- Input schemas ---

// IngestDocumentInput is the input for the ingest_document tool.
type IngestDocumentInput struct {
	Text         string `json:"text" jsonschema:"description=Paper text to ingest (paste the full text here)"`
	ReadingLevel string `json:"reading_level,omitempty" jsonschema:"description=Target reading level: simplified (default) or eli5"`
}

// SearchDocumentsInput is the input for the search_documents tool.
type SearchDocumentsInput struct {
	Query string `json:"query" jsonschema:"description=Search query to match against document titles"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default 10, max 50)"`
}

// GetDocumentInput is the input for the get_document tool.
// Include is a comma-separated list: metadata,sections,evidence,figures.
type GetDocumentInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=Document ID to query"`
	Include    string `json:"include,omitempty" jsonschema:"description=Comma-separated sections to include: metadata,sections,evidence,figures (default: metadata)"`
}

// DocIDInput is the common input for single-document tools without selective retrieval.
type DocIDInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=Document ID to query"`
}

// --- Output schemas ---

// IngestDocumentResult is the output of the ingest_document tool.
type IngestDocumentResult struct {
	DocumentID string         `json:"document_id" jsonschema:"description=Unique document identifier"`
	Title      string         `json:"title" jsonschema:"description=Paper title derived from first line"`
	Status     string         `json:"status" jsonschema:"description=Document status (processing)"`
	SourceType string         `json:"source_type" jsonschema:"description=Input source type (pdf or pasted_text)"`
	Metadata   IngestMetadata `json:"metadata" jsonschema:"description=Ingestion metadata"`
}

// IngestMetadata holds ingestion metadata for the ingest_document response.
type IngestMetadata struct {
	ReadingLevel string `json:"reading_level" jsonschema:"description=Applied reading level"`
	CreatedAt    int64  `json:"created_at" jsonschema:"description=Unix timestamp of creation"`
}

// SearchDocumentsResult is the output of the search_documents tool.
type SearchDocumentsResult struct {
	Documents []SearchDocumentInfo `json:"documents" jsonschema:"description=Matching documents"`
}

// SearchDocumentInfo holds summary info for a search result (P29: adds matched_sections).
type SearchDocumentInfo struct {
	DocumentID      string           `json:"document_id" jsonschema:"description=Document identifier"`
	Title           string           `json:"title" jsonschema:"description=Document title"`
	Status          string           `json:"status" jsonschema:"description=Processing status"`
	CreatedAt       int64            `json:"created_at" jsonschema:"description=Unix timestamp of creation"`
	MatchedSections []ChapterSummary `json:"matched_sections,omitempty" jsonschema:"description=Chapters whose title or summary matches the query"`
}

// ChapterSummary is a lightweight chapter reference for search results.
type ChapterSummary struct {
	ChapterID string `json:"chapter_id" jsonschema:"description=Chapter identifier"`
	Title     string `json:"title" jsonschema:"description=Chapter title"`
}

// GetDocumentResult is the output of the get_document tool (P30: selective retrieval).
// Nil pointer fields indicate the section was not requested.
type GetDocumentResult struct {
	DocumentID       string         `json:"document_id" jsonschema:"description=Unique document identifier"`
	Title            string         `json:"title" jsonschema:"description=Paper title"`
	Status           string         `json:"status" jsonschema:"description=Processing status"`
	SourceType       string         `json:"source_type" jsonschema:"description=Input source type"`
	ReadingLevel     string         `json:"reading_level" jsonschema:"description=Target reading level"`
	CreatedAt        int64          `json:"created_at" jsonschema:"description=Unix timestamp of creation"`
	ProcessingTimeMs *int           `json:"processing_time_ms,omitempty" jsonschema:"description=Processing time in milliseconds"`
	ErrorMessage     *string        `json:"error_message,omitempty" jsonschema:"description=Error details if status is failed"`
	Sections         []ChapterInfo  `json:"sections,omitempty" jsonschema:"description=Simplified text chapters (if included)"`
	Evidence         []EvidenceInfo `json:"evidence,omitempty" jsonschema:"description=Evidence references (if included)"`
	Figures          []FigureInfo   `json:"figures,omitempty" jsonschema:"description=Structured figures (if included)"`
}

// ChapterInfo is a chapter of simplified text.
type ChapterInfo struct {
	ChapterID    string `json:"chapter_id" jsonschema:"description=Chapter identifier"`
	Title        string `json:"title" jsonschema:"description=Chapter title"`
	Summary      string `json:"summary" jsonschema:"description=Chapter summary"`
	Excerpt      string `json:"excerpt" jsonschema:"description=Chapter excerpt"`
	DisplayOrder int    `json:"display_order" jsonschema:"description=Ordering within document"`
}

// FigureInfo is a structured figure record (P32: no AI annotation, source_text included).
type FigureInfo struct {
	ID           string          `json:"id" jsonschema:"description=Chart identifier"`
	SourceMethod string          `json:"source_method" jsonschema:"description=How the chart was created (data_extracted, image_fallback, omitted)"`
	ChartType    string          `json:"chart_type,omitempty" jsonschema:"description=Chart type (bar, line, pie, scatter, area, radar)"`
	ChartData    json.RawMessage `json:"chart_data,omitempty" jsonschema:"description=Structured chart data or null for image_fallback"`
	SourceText   string          `json:"source_text,omitempty" jsonschema:"description=Original text backing this chart"`
	PageNumber   int             `json:"page_number,omitempty" jsonschema:"description=Source page in original document (0 if not applicable)"`
	DisplayOrder int             `json:"display_order" jsonschema:"description=Ordering within document"`
	ChapterID    string          `json:"chapter_id,omitempty" jsonschema:"description=Linked chapter ID"`
	ImageURL     string          `json:"image_url,omitempty" jsonschema:"description=Base64-encoded original chart image (present for image_fallback charts)"`
	Grounding    string          `json:"grounding,omitempty" jsonschema:"description=Grounding status (verified, partial, unsupported, failed)"`
	Provenance   string          `json:"provenance,omitempty" jsonschema:"description=Source method summary for provenance tracking"`
}

// FiguresResult is the output of the get_figures tool.
type FiguresResult struct {
	Figures []FigureInfo `json:"figures" jsonschema:"description=Re-visualized figures from the paper"`
}

// EvidenceInfo holds one piece of evidence (P31: merged with claim references).
type EvidenceInfo struct {
	ID              string         `json:"id" jsonschema:"description=Evidence identifier"`
	Page            *int           `json:"page,omitempty" jsonschema:"description=Source page number"`
	FigureID        string         `json:"figure_id,omitempty" jsonschema:"description=Linked chart/figure ID"`
	TableID         string         `json:"table_id,omitempty" jsonschema:"description=Linked table ID"`
	Section         string         `json:"section,omitempty" jsonschema:"description=Section name"`
	SourceText      string         `json:"source_text" jsonschema:"description=Verbatim source text"`
	SourceReference string         `json:"source_reference" jsonschema:"description=Human-readable citation (e.g. Page 3, Figure 2)"`
	Claims          []ClaimSummary `json:"claims,omitempty" jsonschema:"description=Claims linked to this evidence via claim_evidence"`
}

// ClaimSummary is a lightweight claim reference embedded in evidence responses.
type ClaimSummary struct {
	ClaimID      string  `json:"claim_id" jsonschema:"description=Claim identifier"`
	ClaimText    string  `json:"claim_text" jsonschema:"description=Verbatim claim text"`
	ClaimType    string  `json:"claim_type" jsonschema:"description=Claim category (hypothesis, finding, conclusion, method, limitation)"`
	Confidence   string  `json:"confidence" jsonschema:"description=Confidence level (high, medium, low)"`
	Relationship string  `json:"relationship" jsonschema:"description=How this claim relates to the evidence (supports, contradifies, clarifies)"`
	SourcePage   *int    `json:"source_page,omitempty" jsonschema:"description=Page where the claim appears in the original"`
	SourceText   *string `json:"source_text,omitempty" jsonschema:"description=Original text supporting this claim"`
}

// EvidenceResult is the output of the get_evidence tool (P31: includes claims).
type EvidenceResult struct {
	Evidence []EvidenceInfo `json:"evidence" jsonschema:"description=Evidence references for this paper"`
	Claims   []ClaimInfo    `json:"claims,omitempty" jsonschema:"description=All claims extracted from this paper"`
}

// ClaimInfo is a standalone claim record returned in the evidence response.
type ClaimInfo struct {
	ID          string   `json:"id" jsonschema:"description=Claim identifier"`
	ClaimText   string   `json:"claim_text" jsonschema:"description=Verbatim claim text"`
	ClaimType   string   `json:"claim_type" jsonschema:"description=Claim category (hypothesis, finding, conclusion, method, limitation)"`
	Confidence  string   `json:"confidence" jsonschema:"description=Confidence level (high, medium, low)"`
	SourcePage  *int     `json:"source_page,omitempty" jsonschema:"description=Page where the claim appears in the original"`
	SourceText  *string  `json:"source_text,omitempty" jsonschema:"description=Original text supporting this claim"`
	EvidenceIDs []string `json:"evidence_ids,omitempty" jsonschema:"description=Evidence IDs linked to this claim"`
}
