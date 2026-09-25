// Package credentials resolves a user's stored BYOK credential into a ready
// model client.
//
// PaperViz holds no AI vendor account. Every model call runs on a key the user
// supplied, so there is no process-wide client to inject: the correct client
// depends on who is asking, and it is built per request from a credential
// decrypted moments earlier.
package credentials

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// Errors a caller maps to an HTTP status. They are deliberately coarse: the
// detail of a decryption failure tells an attacker whether a blob exists.
var (
	// ErrNoCredential means the user has not added a model key yet.
	ErrNoCredential = errors.New("no model credential configured")

	// ErrCredentialUnreadable means a credential exists but could not be
	// opened, almost always because CREDENTIAL_ENCRYPTION_KEY changed.
	ErrCredentialUnreadable = errors.New("model credential could not be read")
)

// Resolver turns a user id into a model client.
type Resolver struct {
	db     *sql.DB
	tr     *external.Transport
	cipher *external.Cipher
}

// NewResolver creates a Resolver. The Transport and Cipher are process-wide;
// the clients they produce are not.
func NewResolver(db *sql.DB, tr *external.Transport, cipher *external.Cipher) *Resolver {
	return &Resolver{db: db, tr: tr, cipher: cipher}
}

// Cipher returns the cipher this resolver was built with, so handlers that
// manage credentials share the exact same one. Passing a different cipher
// would make rows unreadable in one direction only.
func (r *Resolver) Cipher() *external.Cipher { return r.cipher }

// For returns a client bound to the user's default credential. It returns
// ErrNoCredential when the user has not configured a key, which callers must
// surface rather than paper over: there is no server-side fallback key, by
// design.
func (r *Resolver) For(ctx context.Context, userID string) (*external.LLM, error) {
	if userID == "" {
		return nil, ErrNoCredential
	}

	stored, err := repository.NewCredentialRepo(r.db).GetDefault(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoCredential
		}
		return nil, fmt.Errorf("load default credential: %w", err)
	}

	provider := external.Provider(stored.Provider)
	plaintext, err := r.cipher.Decrypt(
		stored.Ciphertext, stored.Nonce,
		external.CredentialAAD(stored.UserID, provider, stored.Model),
	)
	if err != nil {
		// Logged without the ciphertext, nonce, or key so a rotation accident
		// is diagnosable without putting secrets in the log.
		slog.Error("decrypt model credential failed",
			"credential_id", stored.ID, "provider", stored.Provider, "error", err)
		return nil, ErrCredentialUnreadable
	}

	client, err := r.tr.For(provider, string(plaintext), stored.Model)
	if err != nil {
		slog.Error("build model client failed",
			"credential_id", stored.ID, "provider", stored.Provider, "error", err)
		return nil, fmt.Errorf("build model client: %w", err)
	}
	return client, nil
}
