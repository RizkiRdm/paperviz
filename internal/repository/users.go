package repository

import (
	"fmt"
)

// User mirrors the users table.
type User struct {
	ID              string
	Email           string
	PasswordHash    string
	PasswordVersion int64
	CreatedAt       int64
	OAuthProvider   string
	OAuthID         string
}

// UserRepo provides CRUD access to the users table.
type UserRepo struct {
	db dbExecutor
}

func NewUserRepo(db dbExecutor) *UserRepo {
	return &UserRepo{db: db}
}

// Insert creates a new user row.
func (r *UserRepo) Insert(u User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetByEmail retrieves a user by email. Returns sql.ErrNoRows if not found.
func (r *UserRepo) GetByEmail(email string) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ?`, email,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID retrieves a user by ID. Returns sql.ErrNoRows if not found.
func (r *UserRepo) GetByID(id string) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at, oauth_provider, oauth_id FROM users WHERE id = ?`, id,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.OAuthProvider, &u.OAuthID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByOAuth retrieves a user by OAuth provider and ID. Returns sql.ErrNoRows if not found.
func (r *UserRepo) GetByOAuth(provider, oauthID string) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at, oauth_provider, oauth_id FROM users WHERE oauth_provider = ? AND oauth_id = ?`,
		provider, oauthID,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.OAuthProvider, &u.OAuthID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpsertByOAuth(u User) error {
	passwordHash := u.PasswordHash
	if passwordHash == "" {
		passwordHash = "oauth-only"
	}
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at, oauth_provider, oauth_id)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(email) DO UPDATE SET oauth_provider = excluded.oauth_provider, oauth_id = excluded.oauth_id`,
		u.ID, u.Email, passwordHash, u.CreatedAt, u.OAuthProvider, u.OAuthID,
	)
	if err != nil {
		return fmt.Errorf("upsert user by oauth: %w", err)
	}
	return nil
}
