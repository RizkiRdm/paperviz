package credentials

import (
	"context"
	"errors"
	"strings"
	"testing"

	"paperviz/internal/external"
)

// TestValidateCredential proves a stored key is checked live and failures are
// classified without ever returning provider internals.
func TestValidateCredential(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	ownID := addCredential(t, db, cipher, "u1", "gemini", "", testKey, true)

	pingErr := func(err error) PingFunc {
		return func(context.Context, *external.LLM) error { return err }
	}

	tests := []struct {
		name       string
		userID     string
		id         string
		ping       PingFunc
		wantValid  bool
		wantReason string
		wantErr    error
	}{
		{name: "working key", userID: "u1", id: ownID, ping: pingErr(nil), wantValid: true},
		{name: "rejected key is invalid_key", userID: "u1", id: ownID, ping: pingErr(errors.New("provider returned status 401: bad key")), wantReason: "invalid_key"},
		{name: "forbidden key is invalid_key", userID: "u1", id: ownID, ping: pingErr(errors.New("provider returned status 403")), wantReason: "invalid_key"},
		{name: "exhausted quota is quota_exceeded", userID: "u1", id: ownID, ping: pingErr(errors.New("daily quota exhausted (429)")), wantReason: "quota_exceeded"},
		{name: "network failure is unavailable", userID: "u1", id: ownID, ping: pingErr(errors.New("dial: connection refused")), wantReason: "unavailable"},
		{name: "unknown id is not found", userID: "u1", id: "ghost", ping: pingErr(nil), wantErr: ErrNotFound},
		{name: "cross-user id is not found", userID: "u2", id: ownID, ping: pingErr(nil), wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(r, tt.ping)
			got, err := v.Validate(context.Background(), tt.userID, tt.id)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if got.Valid != tt.wantValid || got.Reason != tt.wantReason {
				t.Fatalf("got %+v, want valid=%v reason=%q", got, tt.wantValid, tt.wantReason)
			}
			if tt.wantValid && got.Message != "" {
				t.Fatalf("success must not carry a message: %+v", got)
			}
			if !tt.wantValid && got.Message == "" {
				t.Fatal("failure must carry a UI-safe message")
			}
		})
	}
}

// TestValidateNeverEchoesKey proves the result cannot leak the plaintext key
// even when the provider echoes it in its error body.
func TestValidateNeverEchoesKey(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	id := addCredential(t, db, cipher, "u1", "gemini", "", testKey, true)

	v := NewValidator(r, func(context.Context, *external.LLM) error {
		return errors.New("provider returned status 401: key " + testKey + " rejected")
	})
	got, err := v.Validate(context.Background(), "u1", id)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if got.Valid || got.Reason != "invalid_key" {
		t.Fatalf("got %+v, want invalid_key", got)
	}
	if got.Message == "" || len(got.Message) > 120 {
		t.Fatalf("message must be a fixed UI string, got %q", got.Message)
	}
	if strings.Contains(got.Message, testKey) {
		t.Fatalf("message echoed the key: %q", got.Message)
	}
}
