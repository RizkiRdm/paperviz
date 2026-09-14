package mcp

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// registerTools registers exactly 5 MCP tools — the locked agent-first surface.
// ingest_document is deterministic extraction only; no LLM is called in any tool path.
func registerTools(server *mcp.Server, srv *MCPServer) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ingest_document",
		Description: "Ingest paper text via deterministic extraction only. No LLM, no summarization, no reasoning. Returns document_id for polling via get_document.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args IngestDocumentInput) (*mcp.CallToolResult, IngestDocumentResult, error) {
		return handleIngestDocument(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_documents",
		Description: "Search ingested documents by title query. Returns matching document summaries.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchDocumentsInput) (*mcp.CallToolResult, SearchDocumentsResult, error) {
		return handleSearchDocuments(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_document",
		Description: "Retrieve full metadata for an ingested document by its document_id.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, GetDocumentResult, error) {
		return handleGetDocument(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_figures",
		Description: "Retrieve re-visualized charts and figures for an analyzed paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, FiguresResult, error) {
		return handleGetFigures(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_evidence",
		Description: "Retrieve evidence references linking claims to source material for a paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, EvidenceResult, error) {
		return handleGetEvidence(ctx, srv, args)
	})
}

// --- Input/Output types ---

// IngestDocumentInput is the input for the ingest_document tool.
type IngestDocumentInput struct {
	Text         string `json:"text" jsonschema:"description=Paper text to ingest (paste the full text here)"`
	ReadingLevel string `json:"reading_level,omitempty" jsonschema:"description=Target reading level: simplified (default) or eli5"`
}

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

// SearchDocumentsInput is the input for the search_documents tool.
type SearchDocumentsInput struct {
	Query string `json:"query" jsonschema:"description=Search query to match against document titles"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default 10, max 50)"`
}

// SearchDocumentsResult is the output of the search_documents tool.
type SearchDocumentsResult struct {
	Documents []SearchDocumentInfo `json:"documents" jsonschema:"description=Matching documents"`
}

// SearchDocumentInfo holds summary info for a search result.
type SearchDocumentInfo struct {
	DocumentID string `json:"document_id" jsonschema:"description=Document identifier"`
	Title      string `json:"title" jsonschema:"description=Document title"`
	Status     string `json:"status" jsonschema:"description=Processing status"`
	CreatedAt  int64  `json:"created_at" jsonschema:"description=Unix timestamp of creation"`
}

// GetDocumentResult is the output of the get_document tool.
type GetDocumentResult struct {
	DocumentID       string  `json:"document_id" jsonschema:"description=Unique document identifier"`
	Title            string  `json:"title" jsonschema:"description=Paper title"`
	Status           string  `json:"status" jsonschema:"description=Processing status"`
	SourceType       string  `json:"source_type" jsonschema:"description=Input source type"`
	ReadingLevel     string  `json:"reading_level" jsonschema:"description=Target reading level"`
	CreatedAt        int64   `json:"created_at" jsonschema:"description=Unix timestamp of creation"`
	ProcessingTimeMs *int    `json:"processing_time_ms,omitempty" jsonschema:"description=Processing time in milliseconds"`
	ErrorMessage     *string `json:"error_message,omitempty" jsonschema:"description=Error details if status is failed"`
}

// DocIDInput is the common input for single-document tools.
type DocIDInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=Document ID to query"`
}

// ChartInfo holds one re-visualized figure.
type ChartInfo struct {
	ID           string          `json:"id" jsonschema:"description=Chart identifier"`
	SourceMethod string          `json:"source_method" jsonschema:"description=How the chart was created (data_extracted, image_fallback, omitted)"`
	ChartType    string          `json:"chart_type,omitempty" jsonschema:"description=Chart type (bar, line, pie, scatter, area, radar)"`
	ChartData    json.RawMessage `json:"chart_data,omitempty" jsonschema:"description=Structured chart data (labels, values, title) or null for image_fallback"`
	Annotation   string          `json:"annotation,omitempty" jsonschema:"description=Plain-language explanation of the chart"`
	SourceText   string          `json:"source_text,omitempty" jsonschema:"description=Original text backing this chart"`
	PageNumber   int             `json:"page_number,omitempty" jsonschema:"description=Source page in original document (0 if not applicable)"`
	DisplayOrder int             `json:"display_order" jsonschema:"description=Ordering within document"`
	ChapterID    string          `json:"chapter_id,omitempty" jsonschema:"description=Linked chapter ID"`
	ImageURL     string          `json:"image_url,omitempty" jsonschema:"description=Base64-encoded original chart image (present for image_fallback charts)"`
}

// FiguresResult is the output of the get_figures tool.
type FiguresResult struct {
	Charts []ChartInfo `json:"charts" jsonschema:"description=Re-visualized figures from the paper"`
}

// EvidenceInfo holds one piece of evidence linking claims to source material.
type EvidenceInfo struct {
	ID              string `json:"id" jsonschema:"description=Evidence identifier"`
	Page            *int   `json:"page,omitempty" jsonschema:"description=Source page number"`
	FigureID        string `json:"figure_id,omitempty" jsonschema:"description=Linked chart/figure ID"`
	TableID         string `json:"table_id,omitempty" jsonschema:"description=Linked table ID"`
	Section         string `json:"section,omitempty" jsonschema:"description=Section name"`
	SourceText      string `json:"source_text" jsonschema:"description=Verbatim source text"`
	SourceReference string `json:"source_reference" jsonschema:"description=Human-readable citation (e.g. Page 3, Figure 2)"`
}

// EvidenceResult is the output of the get_evidence tool.
type EvidenceResult struct {
	Evidence []EvidenceInfo `json:"evidence" jsonschema:"description=Evidence references for this paper"`
}

// --- Tool handlers ---

// handleIngestDocument performs deterministic text extraction only — no LLM, no pipeline.
func handleIngestDocument(ctx context.Context, srv *MCPServer, args IngestDocumentInput) (*mcp.CallToolResult, IngestDocumentResult, error) {
	if args.Text == "" {
		return errResult(fmt.Errorf("text is required")), IngestDocumentResult{}, nil
	}

	if len(args.Text) > 500*1024 {
		return errResult(ErrSizeLimit), IngestDocumentResult{}, nil
	}

	if !srv.rateLimiter.AllowAnalyze(srv.apiKey) {
		return errResult(ErrRateLimited), IngestDocumentResult{}, nil
	}

	readingLevel := args.ReadingLevel
	if readingLevel == "" {
		readingLevel = repository.ReadingLevelSimplified
	}
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
		return errResult(fmt.Errorf("invalid reading_level %q: must be simplified or eli5", readingLevel)), IngestDocumentResult{}, nil
	}

	intake, _, err := services.ValidateAndInsert(srv.db, readingLevel, false, nil, args.Text, nil)
	if err != nil {
		return errResult(err), IngestDocumentResult{}, nil
	}

	return nil, IngestDocumentResult{
		DocumentID: intake.DocumentID,
		Title:      deriveTitle(intake.OriginalText),
		Status:     repository.StatusProcessing,
		SourceType: intake.SourceType,
		Metadata: IngestMetadata{
			ReadingLevel: readingLevel,
			CreatedAt:    time.Now().Unix(),
		},
	}, nil
}

// handleSearchDocuments searches documents by title and returns matching summaries.
func handleSearchDocuments(ctx context.Context, srv *MCPServer, args SearchDocumentsInput) (*mcp.CallToolResult, SearchDocumentsResult, error) {
	if args.Query == "" {
		return errResult(fmt.Errorf("query is required")), SearchDocumentsResult{}, nil
	}

	if !srv.rateLimiter.AllowRead(srv.apiKey) {
		return errResult(ErrRateLimited), SearchDocumentsResult{}, nil
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	docs, err := repository.NewDocumentRepo(srv.db).SearchByTitle(args.Query, limit)
	if err != nil {
		return errResult(err), SearchDocumentsResult{}, nil
	}

	infos := make([]SearchDocumentInfo, 0, len(docs))
	for _, d := range docs {
		infos = append(infos, SearchDocumentInfo{
			DocumentID: d.ID,
			Title:      d.Title,
			Status:     d.Status,
			CreatedAt:  d.CreatedAt,
		})
	}

	return nil, SearchDocumentsResult{Documents: infos}, nil
}

// handleGetDocument returns full metadata for a single document.
func handleGetDocument(ctx context.Context, srv *MCPServer, args DocIDInput) (*mcp.CallToolResult, GetDocumentResult, error) {
	if !srv.rateLimiter.AllowRead(srv.apiKey) {
		return errResult(ErrRateLimited), GetDocumentResult{}, nil
	}

	doc, err := getDocumentOrError(srv.db, args.DocumentID)
	if err != nil {
		return errResult(err), GetDocumentResult{}, nil
	}

	return nil, GetDocumentResult{
		DocumentID:       doc.ID,
		Title:            doc.Title,
		Status:           doc.Status,
		SourceType:       doc.SourceType,
		ReadingLevel:     doc.ReadingLevel,
		CreatedAt:        doc.CreatedAt,
		ProcessingTimeMs: doc.ProcessingTimeMs,
		ErrorMessage:     doc.ErrorMessage,
	}, nil
}

// handleGetFigures returns charts for a document.
func handleGetFigures(ctx context.Context, srv *MCPServer, args DocIDInput) (*mcp.CallToolResult, FiguresResult, error) {
	if !srv.rateLimiter.AllowRead(srv.apiKey) {
		return errResult(ErrRateLimited), FiguresResult{}, nil
	}

	if _, err := getDocumentOrError(srv.db, args.DocumentID); err != nil {
		return errResult(err), FiguresResult{}, nil
	}

	charts, err := repository.NewChartRepo(srv.db).ListByDocument(args.DocumentID)
	if err != nil {
		return errResult(err), FiguresResult{}, nil
	}

	chartInfos := make([]ChartInfo, 0, len(charts))
	for _, c := range charts {
		info := ChartInfo{
			ID:           c.ID,
			SourceMethod: c.SourceMethod,
			DisplayOrder: c.DisplayOrder,
		}
		if c.ChartData != nil {
			info.ChartData = json.RawMessage(*c.ChartData)
		}
		if c.Annotation != nil {
			info.Annotation = *c.Annotation
		}
		if c.PageNumber != nil {
			info.PageNumber = *c.PageNumber
		}
		if c.ChapterID != nil {
			info.ChapterID = *c.ChapterID
		}
		if len(c.ImageBlob) > 0 {
			info.ImageURL = base64.StdEncoding.EncodeToString(c.ImageBlob)
		}
		chartInfos = append(chartInfos, info)
	}

	return nil, FiguresResult{
		Charts: chartInfos,
	}, nil
}

// handleGetEvidence returns evidence references for a document.
func handleGetEvidence(ctx context.Context, srv *MCPServer, args DocIDInput) (*mcp.CallToolResult, EvidenceResult, error) {
	if !srv.rateLimiter.AllowRead(srv.apiKey) {
		return errResult(ErrRateLimited), EvidenceResult{}, nil
	}

	if _, err := getDocumentOrError(srv.db, args.DocumentID); err != nil {
		return errResult(err), EvidenceResult{}, nil
	}

	evidenceList, err := repository.NewEvidenceRepo(srv.db).ListByPaper(args.DocumentID)
	if err != nil {
		return errResult(err), EvidenceResult{}, nil
	}

	infos := make([]EvidenceInfo, 0, len(evidenceList))
	for _, e := range evidenceList {
		info := EvidenceInfo{
			ID:              e.ID,
			SourceText:      e.SourceText,
			SourceReference: e.SourceReference,
		}
		if e.Page != nil {
			info.Page = e.Page
		}
		if e.FigureID != nil {
			info.FigureID = *e.FigureID
		}
		if e.TableID != nil {
			info.TableID = *e.TableID
		}
		if e.Section != nil {
			info.Section = *e.Section
		}
		infos = append(infos, info)
	}

	return nil, EvidenceResult{
		Evidence: infos,
	}, nil
}

// --- Helpers ---

// getDocumentOrError fetches a document or returns a user-facing error.
func getDocumentOrError(db *sql.DB, docID string) (*repository.Document, error) {
	if docID == "" {
		return nil, fmt.Errorf("document_id is required")
	}
	doc, err := repository.NewDocumentRepo(db).Get(docID)
	if err != nil {
		return nil, fmt.Errorf("document %s: %w", docID, err)
	}
	return doc, nil
}

// errResult builds a CallToolResult with an error message for the MCP client.
func errResult(err error) *mcp.CallToolResult {
	msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

// deriveTitle extracts a title from the first non-empty line of text.
func deriveTitle(text string) string {
	for _, line := range splitLines(text) {
		if len(line) > 200 {
			return line[:200]
		}
		if line != "" {
			return line
		}
	}
	return "Untitled paper"
}

// splitLines breaks text into lines for title extraction.
func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	lines = append(lines, text[start:])
	return lines
}
