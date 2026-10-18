package credentials

import (
	"database/sql"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// Credential is the stored-row shape shared by the store and its callers. It
// aliases the repository type so transport layers can map it to wire shapes
// without importing the repository. The plaintext key is never in this struct.
type Credential = repository.Credential

// ErrNotFound reports a credential lookup that matched no row for that user.
// It is the repository sentinel re-exported so handlers never import it.
var ErrNotFound = repository.ErrCredentialNotFound

// Store owns the lifecycle of a user's BYOK keys: sealed at rest with
// AES-256-GCM, never returned in plaintext, always scoped to one user.
type Store struct {
	db     *sql.DB
	cipher *external.Cipher
}

// NewStore creates a Store. The cipher must be the same one the resolver uses,
// or stored rows become unreadable in one direction only.
func NewStore(db *sql.DB, cipher *external.Cipher) *Store {
	return &Store{db: db, cipher: cipher}
}

// Create seals and stores one key. The first key a user adds becomes their
// default, otherwise they would add a credential and still be told none is
// configured.
func (s *Store) Create(userID, provider, model, apiKey string) (*Credential, error) {
	repo := repository.NewCredentialRepo(s.db)
	existing, err := repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	ciphertext, nonce, err := s.cipher.Encrypt([]byte(apiKey), external.CredentialAAD(userID, external.Provider(provider), model))
	if err != nil {
		return nil, err
	}
	id, err := repository.NewID()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	cred := Credential{
		ID: id, UserID: userID, Provider: provider, Model: model,
		Ciphertext: ciphertext, Nonce: nonce, KeyHint: external.KeyHint(apiKey),
		IsDefault: len(existing) == 0,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Insert(cred); err != nil {
		return nil, err
	}
	return &cred, nil
}

// List returns every credential belonging to userID.
func (s *Store) List(userID string) ([]Credential, error) {
	return repository.NewCredentialRepo(s.db).ListByUser(userID)
}

// GetDefault returns the user's default credential.
func (s *Store) GetDefault(userID string) (*Credential, error) {
	return repository.NewCredentialRepo(s.db).GetDefault(userID)
}

// GetByID returns one of the user's credentials, scoped so a foreign id can
// never be read.
func (s *Store) GetByID(userID, id string) (*Credential, error) {
	return repository.NewCredentialRepo(s.db).GetByID(userID, id)
}

// SetDefault promotes one of the user's credentials to default.
func (s *Store) SetDefault(userID, id string) error {
	return repository.NewCredentialRepo(s.db).SetDefault(userID, id)
}

// Delete removes one of the user's credentials.
func (s *Store) Delete(userID, id string) error {
	return repository.NewCredentialRepo(s.db).Delete(userID, id)
}
