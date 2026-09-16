package mcp

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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
		Description: "Search ingested documents by title query. Returns matching document summaries with relevant chapter sections.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchDocumentsInput) (*mcp.CallToolResult, SearchDocumentsResult, error) {
		return handleSearchDocuments(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_document",
		Description: "Retrieve document data by document_id. Use 'include' to select sections: metadata,sections,evidence,figures.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetDocumentInput) (*mcp.CallToolResult, GetDocumentResult, error) {
		return handleGetDocument(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_figures",
		Description: "Retrieve structured chart/figure data for a document. Returns chart data, source location, and provenance — no AI-generated explanations.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, FiguresResult, error) {
		return handleGetFigures(ctx, srv, args)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_evidence",
		Description: "Retrieve evidence and claims for a document. Returns evidence with linked claims, source text, page/section/figure refs, and provenance. User model judges support/contradiction.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DocIDInput) (*mcp.CallToolResult, EvidenceResult, error) {
		return handleGetEvidence(ctx, srv, args)
	})
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

// handleSearchDocuments searches documents by title and returns matching summaries with chapter sections (P29).
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

	queryLower := strings.ToLower(args.Query)
	infos := make([]SearchDocumentInfo, 0, len(docs))
	for _, d := range docs {
		info := SearchDocumentInfo{
			DocumentID: d.ID,
			Title:      d.Title,
			Status:     d.Status,
			CreatedAt:  d.CreatedAt,
		}

		// P29: find chapters whose title or summary matches the query.
		chapters, err := repository.NewChapterRepo(srv.db).ListByDocument(d.ID)
		if err == nil {
			for _, ch := range chapters {
				if strings.Contains(strings.ToLower(ch.Title), queryLower) ||
					strings.Contains(strings.ToLower(ch.Summary), queryLower) {
					info.MatchedSections = append(info.MatchedSections, ChapterSummary{
						ChapterID: ch.ID,
						Title:     ch.Title,
					})
				}
			}
		}

		infos = append(infos, info)
	}

	return nil, SearchDocumentsResult{Documents: infos}, nil
}

// handleGetDocument returns document data with selective section retrieval (P30).
func handleGetDocument(ctx context.Context, srv *MCPServer, args GetDocumentInput) (*mcp.CallToolResult, GetDocumentResult, error) {
	if !srv.rateLimiter.AllowRead(srv.apiKey) {
		return errResult(ErrRateLimited), GetDocumentResult{}, nil
	}

	doc, err := getDocumentOrError(srv.db, args.DocumentID)
	if err != nil {
		return errResult(err), GetDocumentResult{}, nil
	}

	// Parse include param — default to metadata only.
	includeSet := parseInclude(args.Include)

	result := GetDocumentResult{
		DocumentID:       doc.ID,
		Title:            doc.Title,
		Status:           doc.Status,
		SourceType:       doc.SourceType,
		ReadingLevel:     doc.ReadingLevel,
		CreatedAt:        doc.CreatedAt,
		ProcessingTimeMs: doc.ProcessingTimeMs,
		ErrorMessage:     doc.ErrorMessage,
	}

	// P30: populate only requested sections.
	if includeSet["sections"] {
		chapters, err := repository.NewChapterRepo(srv.db).ListByDocument(doc.ID)
		if err == nil {
			result.Sections = make([]ChapterInfo, 0, len(chapters))
			for _, ch := range chapters {
				result.Sections = append(result.Sections, ChapterInfo{
					ChapterID:    ch.ID,
					Title:        ch.Title,
					Summary:      ch.Summary,
					Excerpt:      ch.Excerpt,
					DisplayOrder: ch.DisplayOrder,
				})
			}
		}
	}

	if includeSet["evidence"] {
		evidenceList, err := repository.NewEvidenceRepo(srv.db).ListByPaper(doc.ID)
		if err == nil {
			result.Evidence = buildEvidenceInfos(evidenceList)
		}
	}

	if includeSet["figures"] {
		charts, err := repository.NewChartRepo(srv.db).ListByDocument(doc.ID)
		if err == nil {
			result.Figures = buildFigureInfos(srv.db, charts)
		}
	}

	return nil, result, nil
}

// handleGetFigures returns structured chart data with source text and provenance (P32).
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

	return nil, FiguresResult{
		Figures: buildFigureInfos(srv.db, charts),
	}, nil
}

// handleGetEvidence returns evidence with merged claims and provenance (P31).
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

	// P31: fetch all claims for this document.
	claimList, err := repository.NewClaimRepo(srv.db).ListByPaper(args.DocumentID)
	if err != nil {
		claimList = nil // non-fatal — return evidence without claims
	}

	// Build claim_evidence link map: evidence_id → []ClaimSummary.
	evidenceToClaims := buildEvidenceClaimMap(claimList, srv.db, args.DocumentID)

	// Build evidence_id → evidence lookup for cross-referencing.
	evidenceByID := make(map[string]EvidenceInfo, len(evidenceList))
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
		info.Claims = evidenceToClaims[e.ID]
		evidenceByID[e.ID] = info
		infos = append(infos, info)
	}

	// Build standalone claims list with linked evidence IDs.
	claimInfos := make([]ClaimInfo, 0, len(claimList))
	for _, c := range claimList {
		ci := ClaimInfo{
			ID:         c.ID,
			ClaimText:  c.ClaimText,
			ClaimType:  c.ClaimType,
			Confidence: c.Confidence,
			SourcePage: c.SourcePage,
			SourceText: c.SourceText,
		}
		// Collect evidence IDs linked to this claim.
		links, err := repository.NewClaimEvidenceRepo(srv.db).GetByClaim(c.ID)
		if err == nil {
			ci.EvidenceIDs = make([]string, 0, len(links))
			for _, link := range links {
				ci.EvidenceIDs = append(ci.EvidenceIDs, link.EvidenceID)
			}
		}
		claimInfos = append(claimInfos, ci)
	}

	return nil, EvidenceResult{
		Evidence: infos,
		Claims:   claimInfos,
	}, nil
}

// --- Shared builders ---

// buildEvidenceInfos converts repository Evidence rows into MCP EvidenceInfo schemas.
func buildEvidenceInfos(evidenceList []repository.Evidence) []EvidenceInfo {
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
	return infos
}

// buildFigureInfos converts repository Chart rows into FigureInfo schemas (P32: no annotation).
func buildFigureInfos(db *sql.DB, charts []repository.Chart) []FigureInfo {
	// Pre-fetch evidence source text keyed by figure_id for source_text population.
	figureEvidence := prefetchFigureEvidence(db, charts)

	figures := make([]FigureInfo, 0, len(charts))
	for _, c := range charts {
		info := FigureInfo{
			ID:           c.ID,
			SourceMethod: c.SourceMethod,
			DisplayOrder: c.DisplayOrder,
			Provenance:   c.SourceMethod,
		}
		if c.ChartData != nil {
			info.ChartData = json.RawMessage(*c.ChartData)
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
		// P32: derive grounding from source method.
		info.Grounding = deriveGrounding(c)

		// P32: source_text from linked evidence, not AI annotation.
		if ev, ok := figureEvidence[c.ID]; ok {
			info.SourceText = ev.SourceText
			if ev.Page != nil && info.PageNumber == 0 {
				info.PageNumber = *ev.Page
			}
		}

		figures = append(figures, info)
	}
	return figures
}

// prefetchFigureEvidence builds a map of figure_id → first linked Evidence for source text.
func prefetchFigureEvidence(db *sql.DB, charts []repository.Chart) map[string]repository.Evidence {
	// Collect unique figure IDs from charts that have chapter links.
	figureIDs := make(map[string]bool)
	for _, c := range charts {
		if c.ChapterID != nil {
			figureIDs[c.ID] = true
		}
	}
	if len(figureIDs) == 0 {
		return nil
	}

	// Query evidence linked to these charts via figure_id.
	rows, err := db.Query(
		`SELECT id, paper_id, page, figure_id, table_id, section, source_text, source_reference
		FROM evidence WHERE figure_id IN (`+placeholderList(len(figureIDs))+`)`,
		figureIDArgs(charts)...,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]repository.Evidence)
	for rows.Next() {
		var e repository.Evidence
		if err := rows.Scan(&e.ID, &e.PaperID, &e.Page, &e.FigureID, &e.TableID, &e.Section, &e.SourceText, &e.SourceReference); err != nil {
			continue
		}
		if e.FigureID != nil && !evidenceExists(result, *e.FigureID) {
			result[*e.FigureID] = e
		}
	}
	return result
}

// placeholderList builds N comma-separated ? placeholders for an IN clause.
func placeholderList(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n*2-1)
	for i := range b {
		if i%2 == 0 {
			b[i] = '?'
		} else {
			b[i] = ','
		}
	}
	return string(b)
}

// figureIDArgs extracts figure IDs from charts as []any for query args.
func figureIDArgs(charts []repository.Chart) []any {
	ids := make([]any, 0)
	for _, c := range charts {
		ids = append(ids, c.ID)
	}
	return ids
}

// evidenceExists checks if a figure_id key already exists in the result map.
func evidenceExists(m map[string]repository.Evidence, figureID string) bool {
	_, ok := m[figureID]
	return ok
}

// deriveGrounding maps source method to a grounding status string.
func deriveGrounding(c repository.Chart) string {
	switch c.SourceMethod {
	case repository.ChartSourceDataExtracted:
		return "verified"
	case repository.ChartSourceImageFallback:
		return "partial"
	case repository.ChartSourceOmitted:
		return "unsupported"
	default:
		return "unsupported"
	}
}

// buildEvidenceClaimMap builds a map of evidence_id → []ClaimSummary for P31 merge.
func buildEvidenceClaimMap(claimList []repository.Claim, db *sql.DB, docID string) map[string][]ClaimSummary {
	result := make(map[string][]ClaimSummary)
	if len(claimList) == 0 {
		return result
	}

	for _, c := range claimList {
		links, err := repository.NewClaimEvidenceRepo(db).GetByClaim(c.ID)
		if err != nil {
			continue
		}
		for _, link := range links {
			summary := ClaimSummary{
				ClaimID:      c.ID,
				ClaimText:    c.ClaimText,
				ClaimType:    c.ClaimType,
				Confidence:   c.Confidence,
				Relationship: link.RelationshipType,
				SourcePage:   c.SourcePage,
				SourceText:   c.SourceText,
			}
			result[link.EvidenceID] = append(result[link.EvidenceID], summary)
		}
	}
	return result
}

// --- Helpers ---

// parseInclude converts a comma-separated include string into a set of section names.
func parseInclude(include string) map[string]bool {
	set := map[string]bool{"metadata": true} // metadata always included
	if include == "" {
		return set
	}
	for _, s := range strings.Split(include, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			set[s] = true
		}
	}
	return set
}

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
