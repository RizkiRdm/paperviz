package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"paperviz/internal/external"
	"paperviz/internal/services"
)

// noindexMiddleware sets X-Robots-Tag so crawlers never index or follow ephemeral share links.
func noindexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

// spaNotFound dispatches unmatched paths per plan chunk 5.1 §1.7: API prefix gets 404 JSON, existing static files are served, GET/HEAD fall back to the SPA shell, other methods get 404.
func spaNotFound(staticDir string, fileServer http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/explain/") {
			w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		cleaned := filepath.Clean("/" + r.URL.Path)
		if strings.Contains(r.URL.Path, "..") {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		full := filepath.Join(staticDir, cleaned)
		if full == staticDir || strings.HasPrefix(full, staticDir+string(os.PathSeparator)) {
			if info, err := os.Stat(full); err == nil && info.Mode().IsRegular() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		writeError(w, http.StatusNotFound, "not_found")
	}
}

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
	analyticsHandler := NewAnalyticsHandler(db)
	usageHandler := NewUsageHandler(services.NewTierService(db))

	r.Route("/api/documents", func(r chi.Router) {
		r.With(rateLimitDocumentCreate, authMiddleware.OptionalAuth, authMiddleware.UsageLimitMiddleware).Post("/", docHandler.Create)
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

	r.With(noindexMiddleware).Get("/share/fig/{shareToken}", shareHandler.GetSharedFigure)
	r.With(noindexMiddleware).Head("/share/fig/{shareToken}", shareHandler.GetSharedFigure)
	r.With(noindexMiddleware).Get("/share/doc/{shareToken}", shareHandler.GetSharedPaper)
	r.With(noindexMiddleware).Head("/share/doc/{shareToken}", shareHandler.GetSharedPaper)
	r.Post("/share-referrals", shareHandler.TrackReferral)

	r.With(authMiddleware.RequireAuth).Get("/analytics", analyticsHandler.GetSummary)
	r.Post("/api/analytics/pricing-view", analyticsHandler.TrackPricingView)
	r.Post("/api/analytics/upgrade-intent", analyticsHandler.TrackUpgradeIntent)
	r.Get("/api/usage", usageHandler.GetUsage)

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

	// spaNotFound serves API 404 JSON, real static files, and the SPA fallback for deep links.
	fileServer := http.FileServer(http.Dir(staticDir))
	r.NotFound(spaNotFound(staticDir, fileServer))

	return r
}
