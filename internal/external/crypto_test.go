package external

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"
)

// testKey returns a deterministic 32-byte key for tests.
func testKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	return k
}

func TestNewCipher(t *testing.T) {
	// 16 and 24 bytes are valid AES key sizes, so aes.NewCipher accepts them.
	// We reject them anyway: a truncated base64 CREDENTIAL_ENCRYPTION_KEY must
	// fail loudly rather than silently downgrade to AES-128.
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"valid 32-byte key", 32, false},
		{"16-byte AES-128 key rejected", 16, true},
		{"24-byte AES-192 key rejected", 24, true},
		{"31-byte key rejected", 31, true},
		{"33-byte key rejected", 33, true},
		{"empty key rejected", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCipher(make([]byte, tt.keyLen))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatal("expected non-nil cipher")
			}
		})
	}
}

func TestCipherRoundTrip(t *testing.T) {
	c, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	plaintext := []byte("AIzaSyD-EXAMPLE-user-supplied-key")
	aad := "user-123|gemini|gemini-2.5-flash"

	ct, nonce, err := c.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Contains(ct, plaintext) {
		t.Fatal("ciphertext contains plaintext")
	}
	if len(nonce) == 0 {
		t.Fatal("expected non-empty nonce")
	}

	got, err := c.Decrypt(ct, nonce, aad)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plaintext)
	}
}

// TestCipherNonceIsRandom guards against nonce reuse, which breaks GCM outright.
func TestCipherNonceIsRandom(t *testing.T) {
	c, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	plaintext := []byte("same-input")
	aad := "user-123|gemini|gemini-2.5-flash"

	seen := make(map[string]bool)
	for i := 0; i < 32; i++ {
		ct, nonce, err := c.Encrypt(plaintext, aad)
		if err != nil {
			t.Fatalf("encrypt %d: %v", i, err)
		}
		if seen[string(nonce)] {
			t.Fatal("nonce reused across encryptions")
		}
		seen[string(nonce)] = true
		if seen[string(ct)] {
			t.Fatal("ciphertext repeated across encryptions")
		}
	}
}

// TestDecryptRejectsWrongAAD proves ciphertext is bound to its row: moving a
// credential blob onto another user's account must fail authentication.
func TestDecryptRejectsWrongAAD(t *testing.T) {
	c, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	plaintext := []byte("AIzaSyD-EXAMPLE")
	ct, nonce, err := c.Encrypt(plaintext, "user-123|gemini|gemini-2.5-flash")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	wrongAADs := []string{
		"user-456|gemini|gemini-2.5-flash",    // different user
		"user-123|anthropic|gemini-2.5-flash", // different provider
		"user-123|gemini|gemini-3.5-flash",    // different model
		"",                                    // empty
	}
	for _, aad := range wrongAADs {
		t.Run(aad, func(t *testing.T) {
			if _, err := c.Decrypt(ct, nonce, aad); err == nil {
				t.Fatal("expected authentication failure, got nil")
			}
		})
	}
}

func TestDecryptRejectsTamperedInput(t *testing.T) {
	c, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	plaintext := []byte("AIzaSyD-EXAMPLE")
	aad := "user-123|gemini|gemini-2.5-flash"
	ct, nonce, err := c.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	t.Run("tampered ciphertext", func(t *testing.T) {
		bad := append([]byte(nil), ct...)
		bad[0] ^= 0xff
		if _, err := c.Decrypt(bad, nonce, aad); err == nil {
			t.Fatal("expected error on tampered ciphertext")
		}
	})

	t.Run("tampered nonce", func(t *testing.T) {
		bad := append([]byte(nil), nonce...)
		bad[0] ^= 0xff
		if _, err := c.Decrypt(ct, bad, aad); err == nil {
			t.Fatal("expected error on tampered nonce")
		}
	})

	t.Run("empty nonce", func(t *testing.T) {
		if _, err := c.Decrypt(ct, []byte{}, aad); err == nil {
			t.Fatal("expected error on empty nonce")
		}
	})

	t.Run("nil ciphertext", func(t *testing.T) {
		if _, err := c.Decrypt(nil, nonce, aad); err == nil {
			t.Fatal("expected error on nil ciphertext")
		}
	})
}

// TestDecryptRejectsWrongKey proves one user's key cannot be read with another key.
func TestDecryptRejectsWrongKey(t *testing.T) {
	a, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher a: %v", err)
	}
	b, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher b: %v", err)
	}
	ct, nonce, err := a.Encrypt([]byte("AIzaSyD-EXAMPLE"), "user-123|gemini|m")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := b.Decrypt(ct, nonce, "user-123|gemini|m"); err == nil {
		t.Fatal("expected decryption under wrong key to fail")
	}
}

func TestEncryptRejectsEmptyPlaintext(t *testing.T) {
	c, err := NewCipher(testKey(t))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	if _, _, err := c.Encrypt([]byte{}, "aad"); err == nil {
		t.Fatal("expected error on empty plaintext")
	}
	if _, _, err := c.Encrypt(nil, "aad"); err == nil {
		t.Fatal("expected error on nil plaintext")
	}
}

func TestKeyHint(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"long key returns last four", "AIzaSyD-1234567890", "7890"},
		{"exactly four chars", "abcd", "abcd"},
		{"shorter than four", "abc", "abc"},
		{"empty key", "", ""},
		{"single char", "x", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeyHint(tt.key)
			if got != tt.want {
				t.Fatalf("KeyHint(%q) = %q; want %q", tt.key, got, tt.want)
			}
			if tt.key != "" && len(got) > 4 {
				t.Fatalf("hint %q leaks more than 4 chars", got)
			}
		})
	}
}

// TestKeyHintDoesNotLeakPrefix guards the UI contract: the hint must never
// contain the leading characters of the real key.
func TestKeyHintDoesNotLeakPrefix(t *testing.T) {
	key := "sk-ant-api03-SUPERSECRET-TAIL"
	hint := KeyHint(key)
	if strings.HasPrefix(key, hint) {
		t.Fatalf("hint %q is a prefix of the key", hint)
	}
	if strings.Contains(hint, "SUPERSECRET") {
		t.Fatalf("hint %q leaks key body", hint)
	}
}
