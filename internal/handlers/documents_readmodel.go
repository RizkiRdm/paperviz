package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/documents"
)

type evidenceResponse struct {
	ID              string  `json:"id"`
	Page            *int    `json:"page,omitempty"`
	FigureID        *string `json:"figure_id,omitempty"`
	TableID         *string `json:"table_id,omitempty"`
	Section         *string `json:"section,omitempty"`
	SourceText      string  `json:"source_text"`
	SourceReference string  `json:"source_reference"`
}

type claimWithEvidenceResponse struct {
	ID         string             `json:"id"`
	ClaimText  string             `json:"claim_text"`
	ClaimType  string             `json:"claim_type"`
	Confidence string             `json:"confidence"`
	SourcePage *int               `json:"source_page,omitempty"`
	SourceText *string            `json:"source_text,omitempty"`
	CreatedAt  int64              `json:"created_at"`
	Evidence   []evidenceResponse `json:"evidence"`
}

type egTableResp struct {
	ID           string  `json:"id"`
	PageNumber   *int    `json:"page_number,omitempty"`
	Caption      *string `json:"caption,omitempty"`
	Headers      string  `json:"headers"`
	Rows         string  `json:"rows"`
	SourceText   *string `json:"source_text,omitempty"`
	DisplayOrder int     `json:"display_order"`
}

type egMethodResp struct {
	ID          string  `json:"id"`
	MethodName  string  `json:"method_name"`
	Description *string `json:"description,omitempty"`
	MethodType  string  `json:"method_type"`
	SourcePage  *int    `json:"source_page,omitempty"`
	SourceText  *string `json:"source_text,omitempty"`
}

type egResultResp struct {
	ID                   string  `json:"id"`
	ResultText           string  `json:"result_text"`
	ResultType           string  `json:"result_type"`
	SupportingEvidenceID *string `json:"supporting_evidence_id,omitempty"`
	SourcePage           *int    `json:"source_page,omitempty"`
	SourceText           *string `json:"source_text,omitempty"`
}

type egCitationResp struct {
	ID           string  `json:"id"`
	CitedPaperID *string `json:"cited_paper_id,omitempty"`
	Authors      *string `json:"authors,omitempty"`
	Title        *string `json:"title,omitempty"`
	Year         *int    `json:"year,omitempty"`
	Venue        *string `json:"venue,omitempty"`
	DOI          *string `json:"doi,omitempty"`
	URL          *string `json:"url,omitempty"`
	SourcePage   *int    `json:"source_page,omitempty"`
}

type evidenceGraphResponse struct {
	Claims    []claimWithEvidenceResponse `json:"claims"`
	Tables    []egTableResp               `json:"tables"`
	Methods   []egMethodResp              `json:"methods"`
	Results   []egResultResp              `json:"results"`
	Citations []egCitationResp            `json:"citations"`
}

func (h *DocumentHandler) GetEvidenceGraph(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	graph, err := h.svc.EvidenceGraph(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	claimResponses := make([]claimWithEvidenceResponse, 0, len(graph.Claims))
	for _, c := range graph.Claims {
		evResponses := make([]evidenceResponse, 0, len(c.Evidence))
		for _, e := range c.Evidence {
			evResponses = append(evResponses, evidenceResponse{
				ID:              e.ID,
				Page:            e.Page,
				FigureID:        e.FigureID,
				TableID:         e.TableID,
				Section:         e.Section,
				SourceText:      e.SourceText,
				SourceReference: e.SourceReference,
			})
		}
		claimResponses = append(claimResponses, claimWithEvidenceResponse{
			ID: c.Claim.ID, ClaimText: c.Claim.ClaimText, ClaimType: c.Claim.ClaimType,
			Confidence: c.Claim.Confidence, SourcePage: c.Claim.SourcePage,
			SourceText: c.Claim.SourceText, CreatedAt: c.Claim.CreatedAt,
			Evidence: evResponses,
		})
	}

	tables := graph.Tables
	tableResponses := make([]egTableResp, 0, len(tables))
	for _, t := range tables {
		tableResponses = append(tableResponses, egTableResp{
			ID: t.ID, PageNumber: t.PageNumber, Caption: t.Caption,
			Headers: t.Headers, Rows: t.Rows, SourceText: t.SourceText,
			DisplayOrder: t.DisplayOrder,
		})
	}

	methods := graph.Methods
	methodResponses := make([]egMethodResp, 0, len(methods))
	for _, m := range methods {
		methodResponses = append(methodResponses, egMethodResp{
			ID: m.ID, MethodName: m.MethodName, Description: m.Description,
			MethodType: m.MethodType, SourcePage: m.SourcePage, SourceText: m.SourceText,
		})
	}

	results := graph.Results
	resultResponses := make([]egResultResp, 0, len(results))
	for _, res := range results {
		resultResponses = append(resultResponses, egResultResp{
			ID: res.ID, ResultText: res.ResultText, ResultType: res.ResultType,
			SupportingEvidenceID: res.SupportingEvidenceID, SourcePage: res.SourcePage,
			SourceText: res.SourceText,
		})
	}

	citations := graph.Citations
	citationResponses := make([]egCitationResp, 0, len(citations))
	for _, c := range citations {
		citationResponses = append(citationResponses, egCitationResp{
			ID: c.ID, CitedPaperID: c.CitedPaperID, Authors: c.Authors,
			Title: c.Title, Year: c.Year, Venue: c.Venue, DOI: c.DOI,
			URL: c.URL, SourcePage: c.SourcePage,
		})
	}

	writeJSON(w, http.StatusOK, evidenceGraphResponse{
		Claims:    claimResponses,
		Tables:    tableResponses,
		Methods:   methodResponses,
		Results:   resultResponses,
		Citations: citationResponses,
	})
}
