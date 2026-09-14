package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"paperviz/internal/app/documents"
	"paperviz/internal/repository"
)

// maxUploadBytes caps request body per ARCHITECTURE.md Section 4.
const maxUploadBytes = 20 << 20 // 20 MiB

// pollMinIntervalHint documents client polling contract (2s min).
const pollMinIntervalHint = 2 * time.Second

type createDocumentResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
}

// Create handles POST /api/documents via Service.Create.
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1<<20)

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "file_too_large")
		return
	}

	readingLevel := r.FormValue("reading_level")
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
		writeError(w, http.StatusBadRequest, "invalid_reading_level")
		return
	}

	pastedText := r.FormValue("text")
	file, header, fileErr := r.FormFile("file")

	hasFile := fileErr == nil
	hasText := pastedText != ""
	if hasFile == hasText {
		writeError(w, http.StatusBadRequest, "missing_input")
		return
	}

	var pdfBytes []byte
	if hasFile {
		defer file.Close()
		if header.Size > maxUploadBytes {
			writeError(w, http.StatusBadRequest, "file_too_large")
			return
		}
		var err error
		pdfBytes, err = io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "file_too_large")
			return
		}
		if !isPDFContent(pdfBytes) {
			writeError(w, http.StatusBadRequest, "invalid_file_type")
			return
		}
	}

	var userID *string
	if uid := UserIDFromContext(r.Context()); uid != "" {
		userID = &uid
	}

	svc := documents.New(h.db, h.gemini)
	docID, code, err := svc.Create(readingLevel, hasFile, pdfBytes, pastedText, userID)
	if err != nil {
		if code == "no_text_layer" {
			writeError(w, http.StatusUnprocessableEntity, "no_text_layer")
			return
		}
		slog.Error("document intake failed", "error", err)
		writeError(w, http.StatusBadRequest, code)
		return
	}

	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: docID, Status: repository.StatusProcessing})
}
