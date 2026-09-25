package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"paperviz/internal/app/credentials"
	"paperviz/internal/services"
)

// ImportService defines fetching for DOI/URL imports.
type ImportService interface {
	FetchByDOI(doi string) (string, string, error)
	FetchByURL(url string) (string, string, error)
}

// importCreator handles persistence for imported documents.
type importCreator interface {
	Create(ctx context.Context, sourceType, readingLevel, originalText, title string, userID *string) (string, error)
}

// dbImportCreator delegates persistence to services layer.
type dbImportCreator struct {
	db       *sql.DB
	provider *credentials.Resolver
}

// Create inserts imported document and starts pipeline via service.
func (c *dbImportCreator) Create(ctx context.Context, sourceType, readingLevel, originalText, title string, userID *string) (string, error) {
	resolvedID := ""
	if userID != nil {
		resolvedID = *userID
	}
	client, err := c.provider.For(ctx, resolvedID)
	if err != nil {
		return "", err
	}
	return services.CreateImportedDocument(c.db, client, sourceType, readingLevel, originalText, title, userID)
}

// ImportHandler handles DOI/URL import endpoints.
type ImportHandler struct {
	importService ImportService
	creator       importCreator
}

// NewImportHandler constructs ImportHandler with DB+credential resolver and optional fetcher.
func NewImportHandler(db *sql.DB, provider *credentials.Resolver, importService ...ImportService) *ImportHandler {
	var svc ImportService
	if len(importService) > 0 {
		svc = importService[0]
	}
	return &ImportHandler{
		importService: svc,
		creator:       &dbImportCreator{db: db, provider: provider},
	}
}

// doiPattern validates DOI format.
var doiPattern = regexp.MustCompile(`^10\.\d{4,9}/[^\s]+$`)

// ImportByDOIRequest is JSON for DOI import.
type ImportByDOIRequest struct {
	DOI          string `json:"doi"`
	ReadingLevel string `json:"reading_level"`
}

// ImportByURLRequest is JSON for URL import.
type ImportByURLRequest struct {
	URL          string `json:"url"`
	ReadingLevel string `json:"reading_level"`
}

// ImportByDOI handles POST /api/import/doi.
func (h *ImportHandler) ImportByDOI(w http.ResponseWriter, r *http.Request) {
	if h.importService == nil {
		writeError(w, http.StatusNotImplemented, "import_not_available")
		return
	}
	var req ImportByDOIRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	req.DOI = strings.TrimSpace(req.DOI)
	if req.DOI == "" {
		writeError(w, http.StatusBadRequest, "doi_required")
		return
	}
	if !doiPattern.MatchString(req.DOI) {
		writeError(w, http.StatusBadRequest, "invalid_doi")
		return
	}
	readingLevel := req.ReadingLevel
	if readingLevel == "" {
		readingLevel = "simplified"
	}
	if readingLevel != "simplified" && readingLevel != "eli5" {
		writeError(w, http.StatusBadRequest, "invalid_reading_level")
		return
	}
	var userID *string
	if uid := UserIDFromContext(r.Context()); uid != "" {
		userID = &uid
	}
	originalText, title, err := h.importService.FetchByDOI(req.DOI)
	if err != nil {
		slog.Error("fetch by DOI failed", "doi", req.DOI, "error", err)
		writeError(w, http.StatusBadGateway, "fetch_failed")
		return
	}
	id, err := h.creator.Create(r.Context(), "doi", readingLevel, originalText, title, userID)
	if err != nil {
		if errors.Is(err, credentials.ErrNoCredential) || errors.Is(err, credentials.ErrCredentialUnreadable) {
			writeCredentialError(w, err)
			return
		}
		slog.Error("create imported document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: id, Status: "processing"})
}

// ImportByURL handles POST /api/import/url.
func (h *ImportHandler) ImportByURL(w http.ResponseWriter, r *http.Request) {
	if h.importService == nil {
		writeError(w, http.StatusNotImplemented, "import_not_available")
		return
	}
	var req ImportByURLRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url_required")
		return
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		writeError(w, http.StatusBadRequest, "invalid_url")
		return
	}
	readingLevel := req.ReadingLevel
	if readingLevel == "" {
		readingLevel = "simplified"
	}
	if readingLevel != "simplified" && readingLevel != "eli5" {
		writeError(w, http.StatusBadRequest, "invalid_reading_level")
		return
	}
	var userID *string
	if uid := UserIDFromContext(r.Context()); uid != "" {
		userID = &uid
	}
	originalText, title, err := h.importService.FetchByURL(req.URL)
	if err != nil {
		slog.Error("fetch by URL failed", "url", req.URL, "error", err)
		writeError(w, http.StatusBadGateway, "fetch_failed")
		return
	}
	id, err := h.creator.Create(r.Context(), "url", readingLevel, originalText, title, userID)
	if err != nil {
		if errors.Is(err, credentials.ErrNoCredential) || errors.Is(err, credentials.ErrCredentialUnreadable) {
			writeCredentialError(w, err)
			return
		}
		slog.Error("create imported document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: id, Status: "processing"})
}
