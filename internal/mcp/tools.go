package mcp

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// registerTools wires all 6 MCP tools to the server.
func registerTools(server *mcp.Server, db *sql.DB, gemini *external.GeminiClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_paper",
		Description: "Submit a paper for analysis. Accepts raw text, runs the full simplification + verification + chart pipeline, and returns the document ID once processing begins.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args AnalyzePaperInput) (*mcp.CallToolResult, AnalysisResult, error) {
		return handleAnalyzePaper(ctx, db, gemini, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_summary",
		Description: "Retrieve the simplified text and detected chapters for an analyzed paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, SummaryResult, error) {
		return handleGetSummary(ctx, db, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_figures",
		Description: "Retrieve re-visualized charts and figures for an analyzed paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, FiguresResult, error) {
		return handleGetFigures(ctx, db, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_claims",
		Description: "Retrieve claim verification data comparing original vs simplified text for a paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, ClaimsResult, error) {
		return handleGetClaims(ctx, db, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_evidence",
		Description: "Retrieve evidence references linking claims to source material for a paper.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, EvidenceResult, error) {
		return handleGetEvidence(ctx, db, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_papers",
		Description: "Generate a structured comparison across 2+ analyzed papers. Returns side-by-side dimensions, agreements, disagreements, and cross-paper evidence claims.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ComparePapersInput) (*mcp.CallToolResult, CompareResult, error) {
		return handleComparePapers(ctx, db, gemini, args)
	})
}

// --- Input/Output types ---

// AnalyzePaperInput is the input for the analyze_paper tool.
type AnalyzePaperInput struct {
	Text         string `json:"text" jsonschema:"description=Raw paper text to analyze (paste the full text here)"`
	ReadingLevel string `json:"reading_level,omitempty" jsonschema:"description=Target reading level: simplified (default) or eli5"`
}

// AnalysisResult is the output of the analyze_paper tool.
type AnalysisResult struct {
	DocumentID       string `json:"document_id" jsonschema:"description=Unique document identifier"`
	Title            string `json:"title" jsonschema:"description=Paper title derived from first line"`
	Status           string `json:"status" jsonschema:"description=processing status (processing, complete, failed, verification_failed)"`
	SourceType       string `json:"source_type" jsonschema:"description=Input source type (pdf or pasted_text)"`
	ProcessingTimeMs *int   `json:"processing_time_ms,omitempty" jsonschema:"description=Total pipeline processing time in milliseconds (null while processing)"`
	ErrorMessage     string `json:"error_message,omitempty" jsonschema:"description=Error details if status is failed"`
}

// DocIDInput is the common input for single-document tools.
type DocIDInput struct {
	DocumentID string `json:"document_id" jsonschema:"description=Document ID returned by analyze_paper"`
}

// ChapterInfo holds one detected chapter section.
type ChapterInfo struct {
	ID           string `json:"id" jsonschema:"description=Chapter identifier"`
	Title        string `json:"title" jsonschema:"description=Section heading"`
	Summary      string `json:"summary" jsonschema:"description=One-sentence section summary"`
	Excerpt      string `json:"excerpt" jsonschema:"description=First paragraph or key text"`
	DisplayOrder int    `json:"display_order" jsonschema:"description=Ordering within document"`
}

// SummaryResult is the output of the get_summary tool.
type SummaryResult struct {
	SimplifiedText string        `json:"simplified_text" jsonschema:"description=Markdown-formatted simplified version of the paper"`
	Chapters       []ChapterInfo `json:"chapters" jsonschema:"description=Detected sections of the simplified text"`
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

// ClaimDiffInfo holds claim verification data.
type ClaimDiffInfo struct {
	OriginalClaims   []string `json:"original_claims" jsonschema:"description=Claims extracted from original text"`
	SimplifiedClaims []string `json:"simplified_claims" jsonschema:"description=Claims extracted from simplified text"`
	MismatchDetected bool     `json:"mismatch_detected" jsonschema:"description=True if original and simplified claims differ"`
	MismatchDetail   string   `json:"mismatch_detail,omitempty" jsonschema:"description=Details of any claim mismatch"`
}

// ClaimsResult is the output of the get_claims tool.
type ClaimsResult struct {
	ClaimDiff *ClaimDiffInfo `json:"claim_diff" jsonschema:"description=Claim verification data (null if not yet verified)"`
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

// ComparePapersInput is the input for the compare_papers tool.
type ComparePapersInput struct {
	DocumentIDs []string `json:"document_ids" jsonschema:"description=Array of 2+ document IDs to compare (minimum 2)"`
}

// CompareResult is the output of the compare_papers tool.
type CompareResult struct {
	Papers         []services.PaperSummary    `json:"papers" jsonschema:"description=Individual paper summaries"`
	Dimensions     []services.ComparisonDimension `json:"dimensions" jsonschema:"description=Side-by-side comparison dimensions"`
	Agreement      []string                   `json:"agreement" jsonschema:"description=Areas where papers agree"`
	Disagreement   []string                   `json:"disagreement" jsonschema:"description=Areas where papers disagree"`
	EvidenceClaims []services.EvidenceClaim   `json:"evidence_claims" jsonschema:"description=Cross-paper claims with per-paper stance"`
}

// --- Tool handlers ---

// handleAnalyzePaper runs the full pipeline for pasted text and returns the document ID.
func handleAnalyzePaper(ctx context.Context, db *sql.DB, gemini *external.GeminiClient, args AnalyzePaperInput) (*mcp.CallToolResult, AnalysisResult, error) {
	if args.Text == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: `{"error":"text is required"}`}},
			IsError: true,
		}, AnalysisResult{}, nil
	}

	readingLevel := args.ReadingLevel
	if readingLevel == "" {
		readingLevel = repository.ReadingLevelSimplified
	}
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"error":"invalid reading_level %q: must be simplified or eli5"}`, readingLevel)}},
			IsError: true,
		}, AnalysisResult{}, nil
	}

	// Validate, extract text, and insert document row.
	intake, _, err := services.ValidateAndInsert(db, readingLevel, false, nil, args.Text, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"error":"intake failed: %s"}`, err.Error())}},
			IsError: true,
		}, AnalysisResult{}, nil
	}

	startTime := time.Now()

	// Run the full pipeline asynchronously (simplification + verification + charts).
	go services.RunPipelineAndPersist(db, gemini, intake.DocumentID, services.PipelineInput{
		OriginalText: intake.OriginalText,
		SourceType:   intake.SourceType,
		ReadingLevel: readingLevel,
	})

	elapsed := int(time.Since(startTime).Milliseconds())

	return nil, AnalysisResult{
		DocumentID:       intake.DocumentID,
		Title:            deriveTitle(intake.OriginalText),
		Status:           repository.StatusProcessing,
		SourceType:       intake.SourceType,
		ProcessingTimeMs: &elapsed,
	}, nil
}

// handleGetSummary returns simplified text + chapters for a document.
func handleGetSummary(ctx context.Context, db *sql.DB, args DocIDInput) (*mcp.CallToolResult, SummaryResult, error) {
	doc, err := getDocumentOrError(db, args.DocumentID)
	if err != nil {
		return errResult(err), SummaryResult{}, nil
	}

	chapters, err := repository.NewChapterRepo(db).ListByDocument(args.DocumentID)
	if err != nil {
		return errResult(err), SummaryResult{}, nil
	}

	chapterInfos := make([]ChapterInfo, 0, len(chapters))
	for _, ch := range chapters {
		chapterInfos = append(chapterInfos, ChapterInfo{
			ID:           ch.ID,
			Title:        ch.Title,
			Summary:      ch.Summary,
			Excerpt:      ch.Excerpt,
			DisplayOrder: ch.DisplayOrder,
		})
	}

	simplifiedText := ""
	if doc.SimplifiedText != nil {
		simplifiedText = *doc.SimplifiedText
	}

	return nil, SummaryResult{
		SimplifiedText: simplifiedText,
		Chapters:       chapterInfos,
	}, nil
}

// handleGetFigures returns charts for a document.
func handleGetFigures(ctx context.Context, db *sql.DB, args DocIDInput) (*mcp.CallToolResult, FiguresResult, error) {
	if _, err := getDocumentOrError(db, args.DocumentID); err != nil {
		return errResult(err), FiguresResult{}, nil
	}

	charts, err := repository.NewChartRepo(db).ListByDocument(args.DocumentID)
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

// handleGetClaims returns claim verification data for a document.
func handleGetClaims(ctx context.Context, db *sql.DB, args DocIDInput) (*mcp.CallToolResult, ClaimsResult, error) {
	if _, err := getDocumentOrError(db, args.DocumentID); err != nil {
		return errResult(err), ClaimsResult{}, nil
	}

	cd, err := repository.NewClaimDiffRepo(db).GetByDocument(args.DocumentID)
	if err != nil {
		// No claim_diff row yet — document may still be processing.
		return nil, ClaimsResult{ClaimDiff: nil}, nil
	}

	var originalClaims, simplifiedClaims []string
	_ = json.Unmarshal([]byte(cd.OriginalClaims), &originalClaims)
	_ = json.Unmarshal([]byte(cd.SimplifiedClaims), &simplifiedClaims)

	detail := ""
	if cd.MismatchDetail != nil {
		detail = *cd.MismatchDetail
	}

	return nil, ClaimsResult{
		ClaimDiff: &ClaimDiffInfo{
			OriginalClaims:   originalClaims,
			SimplifiedClaims: simplifiedClaims,
			MismatchDetected: cd.MismatchDetected,
			MismatchDetail:   detail,
		},
	}, nil
}

// handleGetEvidence returns evidence references for a document.
func handleGetEvidence(ctx context.Context, db *sql.DB, args DocIDInput) (*mcp.CallToolResult, EvidenceResult, error) {
	if _, err := getDocumentOrError(db, args.DocumentID); err != nil {
		return errResult(err), EvidenceResult{}, nil
	}

	evidenceList, err := repository.NewEvidenceRepo(db).ListByPaper(args.DocumentID)
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

// handleComparePapers extracts paper summaries and generates a cross-paper comparison.
func handleComparePapers(ctx context.Context, db *sql.DB, gemini *external.GeminiClient, args ComparePapersInput) (*mcp.CallToolResult, CompareResult, error) {
	if len(args.DocumentIDs) < 2 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: `{"error":"at least 2 document_ids required for comparison"}`}},
			IsError: true,
		}, CompareResult{}, nil
	}

	docRepo := repository.NewDocumentRepo(db)
	var papers []services.PaperSummary
	for _, docID := range args.DocumentIDs {
		doc, err := docRepo.Get(docID)
		if err != nil {
			return errResult(fmt.Errorf("document %s: %w", docID, err)), CompareResult{}, nil
		}
		summary, err := services.ExtractPaperSummary(ctx, gemini, doc.ID, doc.Title, doc.OriginalText)
		if err != nil {
			return errResult(fmt.Errorf("extract summary for %s: %w", docID, err)), CompareResult{}, nil
		}
		papers = append(papers, summary)
	}

	comparison, err := services.ComparePapers(ctx, gemini, papers)
	if err != nil {
		return errResult(err), CompareResult{}, nil
	}

	return nil, CompareResult{
		Papers:         comparison.Papers,
		Dimensions:     comparison.Dimensions,
		Agreement:      comparison.Agreement,
		Disagreement:   comparison.Disagreement,
		EvidenceClaims: comparison.EvidenceClaims,
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
