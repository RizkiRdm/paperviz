package external

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// Credential encryption errors. Callers map these to opaque HTTP responses —
// the underlying detail tells an attacker whether a blob exists.
var (
	ErrEmptyPlaintext  = errors.New("plaintext is empty")
	ErrEmptyCiphertext = errors.New("ciphertext is empty")
	ErrBadNonceSize    = errors.New("nonce has wrong size")
)

// Cipher seals user-supplied model API keys at rest using AES-256-GCM.
//
// The database is not the trust boundary. A stolen dump is inert without the
// key, which is read from CREDENTIAL_ENCRYPTION_KEY and never persisted.
type Cipher struct {
	aead cipher.AEAD
}

// NewCipher builds a Cipher from a 32-byte key. Any other length is rejected
// rather than stretched: no derivation, no silent padding, no surprises.
// aes.NewCipher alone would accept 16 and 24 bytes, which would let a truncated
// CREDENTIAL_ENCRYPTION_KEY silently downgrade the store to AES-128.
func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt seals plaintext against aad, which must identify the owning row
// (user, provider, model). A fresh random nonce per call: GCM nonce reuse
// destroys both confidentiality and forgery resistance, so it is never
// derived from the plaintext.
func (c *Cipher) Encrypt(plaintext []byte, aad string) (ciphertext, nonce []byte, err error) {
	if len(plaintext) == 0 {
		return nil, nil, ErrEmptyPlaintext
	}
	nonce = make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("read nonce: %w", err)
	}
	return c.aead.Seal(nil, nonce, plaintext, []byte(aad)), nonce, nil
}

// Decrypt opens ciphertext. A wrong aad, a wrong key, and a tampered blob all
// fail here indistinguishably — that indistinguishability is the property we
// are buying, so the error is wrapped once and never branched on by callers.
func (c *Cipher) Decrypt(ciphertext, nonce []byte, aad string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, ErrEmptyCiphertext
	}
	if len(nonce) != c.aead.NonceSize() {
		return nil, ErrBadNonceSize
	}
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, []byte(aad))
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// CredentialAAD builds the additional authenticated data for a credential
// blob. Both the write and read paths must call this: if they disagree, GCM
// authentication fails and the credential becomes unreadable. Binding the
// tuple means a row cannot be replayed under a different user, provider, or
// model.
func CredentialAAD(userID string, p Provider, model string) string {
	return userID + "|" + string(p) + "|" + model
}

// KeyHint returns at most the last 4 characters, for display only. Never
// enough to reconstruct a key, and never a prefix — a prefix is both a
// recognisable fingerprint and a searchable leak.
func KeyHint(key string) string {
	if len(key) <= 4 {
		return key
	}
	return key[len(key)-4:]
}
