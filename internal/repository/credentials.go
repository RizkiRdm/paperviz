package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrCredentialNotFound reports that a credential lookup matched no row for
// that user. Distinct from the document-scoped ErrNotFound so logs name the
// resource that was actually missing.
var ErrCredentialNotFound = errors.New("credential not found")

// Credential is one user's model API key, sealed with AES-256-GCM. The
// plaintext key never reaches this struct and never leaves the process.
type Credential struct {
	ID         string
	UserID     string
	Provider   string
	Model      string
	Ciphertext []byte
	Nonce      []byte
	KeyHint    string
	IsDefault  bool
	CreatedAt  int64
	UpdatedAt  int64
}

// CredentialRepo provides CRUD access to user_credentials.
//
// Every method takes userID and filters on it. Scoping in the query rather
// than checking after the fetch means a handler bug cannot turn into a
// cross-account read, because there is no unscoped read to call.
type CredentialRepo struct {
	db dbExecutor
}

func NewCredentialRepo(db dbExecutor) *CredentialRepo {
	return &CredentialRepo{db: db}
}

const credentialColumns = `id, user_id, provider, model, ciphertext, nonce, key_hint, is_default, created_at, updated_at`

// scanCredential maps one row into a Credential, translating the SQLite
// integer flag into a bool.
func scanCredential(scan func(dest ...any) error) (*Credential, error) {
	var c Credential
	var isDefault int
	if err := scan(&c.ID, &c.UserID, &c.Provider, &c.Model, &c.Ciphertext, &c.Nonce, &c.KeyHint, &isDefault, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	c.IsDefault = isDefault == 1
	return &c, nil
}

// defaultFlag converts a bool to the integer column representation.
func defaultFlag(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Insert stores one sealed credential.
func (r *CredentialRepo) Insert(c Credential) error {
	_, err := r.db.Exec(
		`INSERT INTO user_credentials (`+credentialColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.UserID, c.Provider, c.Model, c.Ciphertext, c.Nonce, c.KeyHint, defaultFlag(c.IsDefault), c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert credential: %w", err)
	}
	return nil
}

// GetByID returns one of the given user's credentials, or sql.ErrNoRows.
func (r *CredentialRepo) GetByID(userID, id string) (*Credential, error) {
	c, err := scanCredential(r.db.QueryRow(
		`SELECT `+credentialColumns+` FROM user_credentials WHERE user_id = ? AND id = ?`, userID, id,
	).Scan)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListByUser returns every credential belonging to userID, oldest first.
func (r *CredentialRepo) ListByUser(userID string) ([]Credential, error) {
	rows, err := r.db.Query(
		`SELECT `+credentialColumns+` FROM user_credentials WHERE user_id = ? ORDER BY created_at, id`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	defer rows.Close()

	out := []Credential{}
	for rows.Next() {
		c, err := scanCredential(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan credential: %w", err)
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}
	return out, nil
}

// GetDefault returns the credential flagged default for userID, or
// sql.ErrNoRows when the user has not chosen one.
func (r *CredentialRepo) GetDefault(userID string) (*Credential, error) {
	c, err := scanCredential(r.db.QueryRow(
		`SELECT `+credentialColumns+` FROM user_credentials WHERE user_id = ? AND is_default = 1`, userID,
	).Scan)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// SetDefault promotes one of the user's credentials and demotes the rest.
// Ownership is checked before anything is written, so a rejected call cannot
// leave the user with no default at all.
func (r *CredentialRepo) SetDefault(userID, id string) error {
	var exists int
	err := r.db.QueryRow(`SELECT 1 FROM user_credentials WHERE user_id = ? AND id = ?`, userID, id).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCredentialNotFound
	}
	if err != nil {
		return fmt.Errorf("check credential ownership: %w", err)
	}

	now := time.Now().Unix()
	if _, err := r.db.Exec(
		`UPDATE user_credentials SET is_default = 0, updated_at = ? WHERE user_id = ? AND is_default = 1`, now, userID,
	); err != nil {
		return fmt.Errorf("demote previous default: %w", err)
	}
	if _, err := r.db.Exec(
		`UPDATE user_credentials SET is_default = 1, updated_at = ? WHERE user_id = ? AND id = ?`, now, userID, id,
	); err != nil {
		return fmt.Errorf("promote default: %w", err)
	}
	return nil
}

// Delete removes one of the user's credentials. Deleting an id that is absent
// or owned by someone else is a no-op, not an error.
func (r *CredentialRepo) Delete(userID, id string) error {
	if _, err := r.db.Exec(`DELETE FROM user_credentials WHERE user_id = ? AND id = ?`, userID, id); err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}
	return nil
}
