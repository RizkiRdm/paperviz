package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"paperviz/internal/repository"
)

// AuthHandler handles signup, login, and session management.
type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// signupRequest is the wire shape for POST /api/auth/signup.
type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest is the wire shape for POST /api/auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// userResponse is the wire shape for user data in responses.
type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Signup handles POST /api/auth/signup. Creates a new user and session.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !isValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "invalid_email")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password_too_short")
		return
	}

	// Hash password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	// Generate user ID
	userID, err := repository.NewID()
	if err != nil {
		slog.Error("generate user id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	// Insert user
	userRepo := repository.NewUserRepo(h.db)
	if err := userRepo.Insert(repository.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().Unix(),
	}); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "email_taken")
			return
		}
		slog.Error("insert user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	// Create session
	if err := h.createSessionAndSetCookie(w, userID); err != nil {
		slog.Error("create session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{ID: userID, Email: req.Email})
}

// Login handles POST /api/auth/login. Verifies credentials and creates session.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	userRepo := repository.NewUserRepo(h.db)
	user, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		slog.Error("get user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	// Create session
	if err := h.createSessionAndSetCookie(w, user.ID); err != nil {
		slog.Error("create session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Email: user.Email})
}

// Me handles GET /api/auth/me. Returns current user from session cookie.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	sessionRepo := repository.NewSessionRepo(h.db)
	session, err := sessionRepo.Get(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	userRepo := repository.NewUserRepo(h.db)
	user, err := userRepo.GetByID(session.UserID)
	if err != nil {
		slog.Error("get user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Email: user.Email})
}

// Logout handles POST /api/auth/logout. Deletes session and clears cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionRepo := repository.NewSessionRepo(h.db)
		_ = sessionRepo.Delete(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, nil)
}

// createSessionAndSetCookie creates a session and sets the httpOnly cookie.
func (h *AuthHandler) createSessionAndSetCookie(w http.ResponseWriter, userID string) error {
	token, err := generateSessionToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix() // 7 days

	sessionRepo := repository.NewSessionRepo(h.db)
	if err := sessionRepo.Insert(repository.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// generateSessionToken creates a cryptographically secure random token.
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	atIndex := strings.LastIndex(email, "@")
	if atIndex < 1 || atIndex >= len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	return strings.Contains(domain, ".")
}

// hasMinComplexity checks password has uppercase, lowercase, digit.
func hasMinComplexity(password string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}
