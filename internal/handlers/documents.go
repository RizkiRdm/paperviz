// Package handlers contains PaperViz's HTTP layer only: parsing requests,
// validating input shape, calling into services, and serializing responses.
// Per ARCHITECTURE.md Section 2, handlers MUST NOT call repository directly
// — every persistence operation goes through a service function first (even
// though today those service functions are thin wrappers, keeping the call
// path handlers -> services -> repository consistent is what lets us add
// real business logic later without restructuring call sites).
package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/credentials"
	"paperviz/internal/app/documents"
	"paperviz/internal/apperr"
	"paperviz/internal/external"
	"paperviz/internal/services"
)

// DocumentHandler holds everything the two document endpoints need: a DB
// handle (for opening transactions) and a credential resolver, which produces
// the model client for whoever is making the request. It has no other state —
// request-scoped values are never stored on this struct, per AGENTS.md "MUST
// NOT use global mutable state for request-scoped data."
type DocumentHandler struct {
	db       *sql.DB
	provider *credentials.Resolver
	// svc owns document domain logic; the handler is transport only.
	svc *documents.Service
	// images is nil when external chart-image storage is unconfigured, which
	// is a supported deployment that keeps images in the database column.
	images external.ChartImageStore
}

func NewDocumentHandler(db *sql.DB, provider *credentials.Resolver) *DocumentHandler {
	return &DocumentHandler{db: db, provider: provider, svc: documents.New(db, provider)}
}

// NewDocumentHandlerWithImages builds a handler backed by external chart-image
// storage. A nil store is valid and selects the database column.
func NewDocumentHandlerWithImages(db *sql.DB, provider *credentials.Resolver, images external.ChartImageStore) *DocumentHandler {
	return &DocumentHandler{db: db, provider: provider, svc: documents.New(db, provider), images: images}
}

// resolveClient builds a model client for the requesting user and writes the
// error response itself, returning false when the caller should stop. Every
// endpoint that reaches a model goes through here, so an unconfigured key
// produces the same answer everywhere instead of a per-handler variant.
func (h *DocumentHandler) resolveClient(w http.ResponseWriter, r *http.Request) (*external.LLM, bool) {
	client, err := h.provider.For(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		writeCredentialError(w, err)
		return nil, false
	}
	return client, true
}

// RequireDocAccess rejects a request before its handler runs unless the caller
// may access the document named by the :id route parameter. It backs the
// section and relationship routes so their ownership rule is the one Get uses
// instead of none (Wave G found these routes unguarded).
func (h *DocumentHandler) RequireDocAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := services.DocumentAccessAllowed(h.db, chi.URLParam(r, "id"), UserIDFromContext(r.Context()))
		if err != nil {
			writeDocumentAccessError(w, err, "document access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeDocumentAccessError maps an ownership failure to 403 and a missing
// document to 404, so a cross-user write never leaks whether the id exists.
func writeDocumentAccessError(w http.ResponseWriter, err error, msg string) {
	switch {
	case errors.Is(err, apperr.Permission):
		writeError(w, http.StatusForbidden, "forbidden")
	case services.IsDocumentNotFound(err):
		writeError(w, http.StatusNotFound, "not_found")
	default:
		slog.Error(msg, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
