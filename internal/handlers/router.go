package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"paperviz/internal/external"
)

// NewRouter builds the full chi router: the two unversioned API endpoints
// (ARCHITECTURE.md Section E) plus static file serving for the built React
// SPA. staticDir is the frontend's built assets directory (frontend/dist),
// served directly by this Go binary — no separate frontend server in
// production, per ARCHITECTURE.md's "single binary" architecture style.
func NewRouter(db *sql.DB, gemini *external.GeminiClient, staticDir string) http.Handler {
	r := chi.NewRouter()

	// RequestID injects a unique request ID into every request context so
	// log lines belonging to the same request can be correlated.
	r.Use(middleware.RequestID)

	// Recoverer converts a panic in any handler into a 500 response instead
	// of crashing the whole server process — this is standard chi
	// middleware, not a new dependency (chi ships it).
	r.Use(middleware.Recoverer)

	// SecurityHeaders sets X-Content-Type-Options, X-Frame-Options, and
	// CSP on every response. Must run before logging so headers are set
	// when log captures response.
	r.Use(SecurityHeaders)

	// slogRequestLogger logs every request with method, path, status,
	// duration, and request ID — structured JSON output via slog.
	r.Use(slogRequestLogger)

	docHandler := NewDocumentHandler(db, gemini)
	authMiddleware := NewAuthMiddleware(db)
	shareHandler := NewShareHandler(db)

	r.Route("/api/documents", func(r chi.Router) {
		r.With(rateLimitDocumentCreate, authMiddleware.OptionalAuth).Post("/", docHandler.Create)
		r.Get("/{id}", docHandler.Get)
		r.Get("/{id}/charts/{chartId}/image", docHandler.GetChartImage)
		r.With(authMiddleware.RequireAuth).Get("/stats", docHandler.Stats)
		r.With(authMiddleware.RequireAuth).Get("/", docHandler.List)
		r.With(authMiddleware.RequireAuth).Put("/{id}/save", docHandler.ToggleSaved)
		r.With(authMiddleware.RequireAuth).Patch("/{id}", docHandler.UpdateTitle)
		r.With(authMiddleware.RequireAuth).Delete("/{id}", docHandler.Delete)
		r.With(authMiddleware.RequireAuth).Post("/{id}/charts/{chartId}/share", shareHandler.GenerateToken)
		r.With(authMiddleware.RequireAuth).Delete("/{id}/charts/{chartId}/share", shareHandler.RevokeToken)
		r.With(authMiddleware.RequireAuth).Post("/{id}/share", shareHandler.GenerateDocToken)
		r.With(authMiddleware.RequireAuth).Delete("/{id}/share", shareHandler.RevokeDocToken)
		r.With(authMiddleware.RequireAuth).Patch("/{id}/visibility", shareHandler.UpdateVisibility)
		r.Post("/compare", docHandler.Compare)
	})

	r.Get("/share/fig/{shareToken}", shareHandler.GetSharedFigure)
	r.Get("/share/doc/{shareToken}", shareHandler.GetSharedPaper)
	r.Post("/share-referrals", shareHandler.TrackReferral)

	authHandler := NewAuthHandler(db)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.Me)
	})

	collectionHandler := NewCollectionHandler(db)
	r.Route("/api/collections", func(r chi.Router) {
		r.With(authMiddleware.RequireAuth).Post("/", collectionHandler.Create)
		r.With(authMiddleware.RequireAuth).Get("/", collectionHandler.List)
		r.With(authMiddleware.RequireAuth).Get("/{id}", collectionHandler.Get)
		r.With(authMiddleware.RequireAuth).Patch("/{id}", collectionHandler.Rename)
		r.With(authMiddleware.RequireAuth).Delete("/{id}", collectionHandler.Delete)
		r.With(authMiddleware.RequireAuth).Post("/{id}/documents", collectionHandler.AddDocument)
		r.With(authMiddleware.RequireAuth).Delete("/{id}/documents/{docId}", collectionHandler.RemoveDocument)
	})

	// Serve the built SPA for everything else. Any unmatched path falls
	// back to index.html so client-side routing (if the frontend adds any)
	// still works on a hard refresh.
	fileServer := http.FileServer(http.Dir(staticDir))
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		fileServer.ServeHTTP(w, req)
	})

	return r
}
