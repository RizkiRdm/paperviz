package mcp

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxHTTPBodyBytes bounds any incoming StreamableHTTP request body. The
// ingest tool already caps paper text at 500KB; this is the looser
// transport-level sanity bound (well above worst-case JSON escaping).
const maxHTTPBodyBytes = 8 << 20

// requestUserKey is the context key carrying the authenticated user ID.
type requestUserKey struct{}

// withUser derives a request context carrying the authenticated user ID.
func withUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, requestUserKey{}, userID)
}

// HTTPServer is the remote MCP transport: stateless StreamableHTTP where
// every request authenticates by API key and resolves to a user identity.
// Rate and job limiters are shared across requests so limits actually bind.
type HTTPServer struct {
	db          *sql.DB
	rateLimiter *RateLimiter
	jobLimiter  *JobLimiter
	maxBytes    int64
}

// NewHTTPServer creates the remote transport over shared limiters.
func NewHTTPServer(db *sql.DB) *HTTPServer {
	return &HTTPServer{
		db:          db,
		rateLimiter: NewRateLimiter(),
		jobLimiter:  NewJobLimiter(3),
		maxBytes:    maxHTTPBodyBytes,
	}
}

// Handler returns the HTTP handler: auth middleware over the StreamableHTTP
// transport. Each request builds a user-scoped server sharing limiters.
func (h *HTTPServer) Handler() http.Handler {
	inner := sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		userID, _ := r.Context().Value(requestUserKey{}).(string)
		return newServerWithDeps(h.db, userID, userID, true, h.rateLimiter, h.jobLimiter).Server()
	}, &sdk.StreamableHTTPOptions{
		Stateless:           true,
		MaxRequestBodyBytes: h.maxBytes,
	})
	return http.HandlerFunc(h.authenticate(inner.ServeHTTP))
}

// authenticate resolves the presented key to a user once per request.
// Missing/invalid keys get a generic 401 — the key is never echoed or
// logged, and a failed lookup is indistinguishable from a missing key.
func (h *HTTPServer) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		presented := presentedKey(r)
		if presented == "" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		userID, err := h.lookupUser(presented)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if userID == "" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r.WithContext(withUser(r.Context(), userID)))
	}
}

// presentedKey extracts the service key from Authorization, Bearer-prefixed
// or raw. The value is only ever hashed for lookup, never logged.
func presentedKey(r *http.Request) string {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return ""
	}
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return auth
}

// hashPresentedKey returns the hex SHA-256 digest stored in
// users.api_key_hash — the raw presented key never touches storage.
func hashPresentedKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// lookupUser resolves a presented service key to its owner's user ID by
// digest. An empty result means no match — indistinguishable from a
// missing key to the caller.
func (h *HTTPServer) lookupUser(presented string) (string, error) {
	var userID string
	err := h.db.QueryRow(
		`SELECT id FROM users WHERE api_key_hash = ? AND api_key_hash != ''`, hashPresentedKey(presented),
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}
