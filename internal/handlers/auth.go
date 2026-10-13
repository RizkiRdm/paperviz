package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"paperviz/internal/repository"
)

// AuthHandler serves the endpoints that still belong to PaperViz after Clerk
// took over identity: only /api/auth/me, which returns the local profile for an
// already-verified Clerk subject.
//
// Signup, login, logout and session storage are Clerk's. They are absent here
// on purpose — a local credential path is one Clerk cannot revoke, so leaving a
// second one behind would reintroduce the exact failure the pivot removed.
type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// userResponse is the wire shape for user data in responses.
type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Me handles GET /api/auth/me. RequireAuth has already verified the Clerk token
// and provisioned the local row, so this only reads it back.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	user, err := repository.NewUserRepo(h.db).GetByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		slog.Error("get user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Email: user.Email})
}
