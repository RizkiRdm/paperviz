package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"paperviz/internal/repository"
	"paperviz/internal/services"
)

type contextKey string

const userIDKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(userIDKey).(string); ok {
		return uid
	}
	return ""
}

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

func (w *responseWriterWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func slogRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		start := time.Now()
		ww := &responseWriterWrapper{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration", duration.String(),
		}
		if reqID != "" {
			args = append([]any{"id", reqID}, args...)
		}

		level := slog.LevelInfo
		if ww.status >= 500 {
			level = slog.LevelError
		} else if ww.status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "request", args...)
	})
}

// AuthMiddleware groups session-based auth middleware that needs DB access.
type AuthMiddleware struct {
	db *sql.DB
}

func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// RequireAuth rejects requests without a valid session cookie.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}

		sessionRepo := repository.NewSessionRepo(m.db)
		session, err := sessionRepo.Get(cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth attaches user_id to context if session cookie is valid.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		sessionRepo := repository.NewSessionRepo(m.db)
		session, err := sessionRepo.Get(cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// usageLimitResponse is the wire format returned when the per-fingerprint
// submission cap is reached. There is no plan, tier, or upgrade path here:
// PaperViz is free and BYOK, so this is abuse protection on an unauthenticated
// endpoint, not a quota sold to a paying customer.
type usageLimitResponse struct {
	Error      string `json:"error"`
	PapersUsed int    `json:"papers_used"`
	Limit      int    `json:"limit"`
}

// UsageLimitMiddleware rejects document creation requests when the
// fingerprint's submission count exceeds the cap.
func (m *AuthMiddleware) UsageLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp := GetFingerprint(r)
		canCreate, papersUsed, err := services.CheckUsage(m.db, fp)
		if err != nil {
			slog.Error("usage check failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		if !canCreate {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(usageLimitResponse{
				Error:      "submission limit reached for this device",
				PapersUsed: papersUsed,
				Limit:      services.LimitFree,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
