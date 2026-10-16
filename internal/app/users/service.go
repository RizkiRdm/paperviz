// Package users owns the local profile rows that mirror Clerk identities.
// Clerk holds signup and passwords; PaperViz holds only what its own data
// model needs, provisioned on first authenticated request.
package users

import (
	"database/sql"

	"paperviz/internal/repository"
)

// User is the local profile row. PasswordHash is always empty — Clerk owns
// passwords and no code may trust that field.
type User = repository.User

// Service provides user-row reads and first-sight provisioning.
type Service struct {
	db *sql.DB
}

// New creates a Service bound to the given database.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetByID returns the local profile row, or sql.ErrNoRows when the identity
// has never made an authenticated request. The sentinel is passed through
// untouched so callers keep their existing missing-row handling.
func (s *Service) GetByID(id string) (*User, error) {
	return repository.NewUserRepo(s.db).GetByID(id)
}

// ProvisionClerk creates or refreshes the local profile row for a verified
// Clerk identity. A UNIQUE(email) collision is returned rather than swallowed:
// swallowing it would leave the caller authenticated against an account it
// does not own.
func (s *Service) ProvisionClerk(subject, email string, createdAt int64) error {
	return repository.NewUserRepo(s.db).UpsertClerkUser(subject, email, createdAt)
}
