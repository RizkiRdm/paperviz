package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
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

// errUnauthenticated covers every "no usable identity" case: no token, an
// unverifiable token, or a Clerk user PaperViz could not resolve.
var errUnauthenticated = errors.New("unauthenticated")

// AuthMiddleware resolves a verified Clerk identity onto a local users row.
type AuthMiddleware struct {
	db *sql.DB
}

func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// RequireAuth rejects requests without a valid Clerk session token.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.resolveUserID(r.Context())
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth attaches user_id when a valid Clerk session token is present.
// An invalid token is treated as anonymous rather than as an error, because
// every route behind it is also reachable without a session.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.resolveUserID(r.Context())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveUserID turns verified Clerk claims into a local user id, creating the
// local profile row the first time an identity is seen.
//
// The local row is created on first sight rather than at signup because Clerk
// owns the signup flow; PaperViz never learns that an account exists until the
// account makes an authenticated request. Steady-state requests are one indexed
// lookup with no call to Clerk.
func (m *AuthMiddleware) resolveUserID(ctx context.Context) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok || claims == nil || claims.Subject == "" {
		return "", errUnauthenticated
	}

	users := repository.NewUserRepo(m.db)
	local, err := users.GetByID(claims.Subject)
	if err == nil {
		return local.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("get local user: %w", err)
	}

	remote, err := clerkuser.Get(ctx, claims.Subject)
	if err != nil {
		return "", fmt.Errorf("get clerk user: %w", err)
	}
	email := primaryEmail(remote)
	if email == "" {
		return "", fmt.Errorf("clerk user %s has no primary email", claims.Subject)
	}
	if err := users.UpsertClerkUser(claims.Subject, email, time.Now().Unix()); err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// primaryEmail picks the address Clerk marks primary, falling back to the first
// non-empty one. The fallback covers an instance whose primary id is unset but
// whose addresses are still usable as the users.email UNIQUE key.
func primaryEmail(u *clerk.User) string {
	for _, addr := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && addr.ID == *u.PrimaryEmailAddressID && addr.EmailAddress != "" {
			return addr.EmailAddress
		}
	}
	for _, addr := range u.EmailAddresses {
		if addr.EmailAddress != "" {
			return addr.EmailAddress
		}
	}
	return ""
}

// usageLimitResponse is the wire format returned when monthly paper limit is reached.
type usageLimitResponse struct {
	Error      string `json:"error"`
	Tier       string `json:"tier"`
	PapersUsed int    `json:"papers_used"`
	Limit      int    `json:"limit"`
	UpgradeCTA string `json:"upgrade_cta"`
}

// UsageLimitMiddleware rejects document creation requests when the
// fingerprint's monthly paper count exceeds the tier limit.
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
				Error:      "monthly limit reached",
				Tier:       "free",
				PapersUsed: papersUsed,
				Limit:      services.LimitFree,
				UpgradeCTA: "Upgrade to Pro for more papers",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
