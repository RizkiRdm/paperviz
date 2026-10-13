package repository

import (
	"fmt"
)

// User mirrors the users table.
//
// PasswordHash is always "" — Clerk owns passwords, so no code may trust this
// field. It is kept only because the column still exists.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    int64
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
		`SELECT id, email, password_hash, created_at FROM users WHERE id = ?`, id,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpsertClerkUser creates or refreshes the local profile row for a verified
// Clerk identity. A UNIQUE(email) collision means two Clerk identities claim
// one email, so it is returned rather than swallowed: swallowing it would leave
// the caller authenticated against an account it does not own.
func (r *UserRepo) UpsertClerkUser(id, email string, createdAt int64) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at) VALUES (?, ?, '', ?)
		 ON CONFLICT(id) DO UPDATE SET email = excluded.email`,
		id, email, createdAt,
	)
	if err != nil {
		return fmt.Errorf("upsert clerk user: %w", err)
	}
	return nil
}
