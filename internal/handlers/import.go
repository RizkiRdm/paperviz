package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// ImportService defines the interface for fetching content by DOI or URL.
type ImportService interface {
	// FetchByDOI retrieves paper content by DOI identifier.
	FetchByDOI(doi string) (string, string, error)
	// FetchByURL retrieves paper content from a URL.
	FetchByURL(url string) (string, string, error)
}

// ImportHandler handles paper import via DOI and URL endpoints.
type ImportHandler struct {
	db            *sql.DB
	gemini        *external.GeminiClient
	importService ImportService // optional — nil means import endpoints return 501
}

// NewImportHandler constructs an ImportHandler with required dependencies.
// importService may be nil if the import feature is not yet wired up.
func NewImportHandler(db *sql.DB, gemini *external.GeminiClient, importService ...ImportService) *ImportHandler {
	var svc ImportService
	if len(importService) > 0 {
		svc = importService[0]
	}
	return &ImportHandler{db: db, gemini: gemini, importService: svc}
}

// doiPattern validates standard DOI format (10.XXXX/xxxxx).
var doiPattern = regexp.MustCompile(`^10\.\d{4,9}/[^\s]+$`)

// ImportByDOIRequest is the JSON shape for DOI import requests.
type ImportByDOIRequest struct {
	DOI          string `json:"doi"`
	ReadingLevel string `json:"reading_level"`
}

// ImportByURLRequest is the JSON shape for URL import requests.
type ImportByURLRequest struct {
	URL          string `json:"url"`
	ReadingLevel string `json:"reading_level"`
}

// ImportByDOI handles POST /api/import/doi — fetches paper by DOI and starts processing.
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
		readingLevel = repository.ReadingLevelSimplified
	}
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
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

	id, err := repository.NewID()
	if err != nil {
		slog.Error("generate document id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().Unix()
	doc := repository.Document{
		ID:             id,
		CreatedAt:      now,
		LastAccessedAt: now,
		Status:         repository.StatusProcessing,
		SourceType:     "doi",
		ReadingLevel:   readingLevel,
		Title:          title,
		OriginalText:   originalText,
		UserID:         userID,
	}

	docRepo := repository.NewDocumentRepo(h.db)
	if err := docRepo.Insert(doc); err != nil {
		slog.Error("insert document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	go services.RunPipelineAndPersist(h.db, h.gemini, id, services.PipelineInput{
		OriginalText: originalText,
		SourceType:   "doi",
		ReadingLevel: readingLevel,
	})

	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: id, Status: repository.StatusProcessing})
}

// ImportByURL handles POST /api/import/url — fetches paper by URL and starts processing.
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
		readingLevel = repository.ReadingLevelSimplified
	}
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
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

	id, err := repository.NewID()
	if err != nil {
		slog.Error("generate document id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().Unix()
	doc := repository.Document{
		ID:             id,
		CreatedAt:      now,
		LastAccessedAt: now,
		Status:         repository.StatusProcessing,
		SourceType:     "url",
		ReadingLevel:   readingLevel,
		Title:          title,
		OriginalText:   originalText,
		UserID:         userID,
	}

	docRepo := repository.NewDocumentRepo(h.db)
	if err := docRepo.Insert(doc); err != nil {
		slog.Error("insert document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	go services.RunPipelineAndPersist(h.db, h.gemini, id, services.PipelineInput{
		OriginalText: originalText,
		SourceType:   "url",
		ReadingLevel: readingLevel,
	})

	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: id, Status: repository.StatusProcessing})
}
