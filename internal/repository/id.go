package repository

import (
	"crypto/rand"
)

// alphabet avoids visually ambiguous characters (0/O, 1/l/I) to keep share
// links easy to read aloud/retype without changing entropy meaningfully.
const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// NewID returns a cryptographically random, non-guessable, non-sequential
// identifier of at least 12 characters, per ARCHITECTURE.md Section 5.
func NewID() (string, error) {
	const length = 16
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := make([]byte, length)
	for i, b := range buf {
		id[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(id), nil
}
