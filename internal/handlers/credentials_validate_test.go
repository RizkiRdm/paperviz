package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
)

// newValidateHandler builds a credential handler whose provider ping is
// stubbed, so validate is exercised without any network call.
func newValidateHandler(t *testing.T, ping credentials.PingFunc) (*CredentialHandler, string) {
	t.Helper()
	db := openCredentialHandlerDB(t)
	cipher := testCipher(t)
	resolver := credentials.NewResolver(db, external.NewTransport(), cipher)
	h := newCredentialHandlerWithValidator(db, cipher, resolver, ping)
	w := doJSON(t, h.Create, http.MethodPost, `{"provider":"gemini","api_key":"`+testModelKey+`"}`, "u1")
	if w.Code != http.StatusCreated {
		t.Fatalf("seed credential: status %d; body %s", w.Code, w.Body.String())
	}
	var created credentialResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode seed: %v", err)
	}
	return h, created.ID
}

// doValidate calls the validate endpoint for a user and credential id.
func doValidate(t *testing.T, h *CredentialHandler, userID, id string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/credentials/"+id+"/validate", nil)
	if userID != "" {
		r = asUser(r, userID)
	}
	rec := httptest.NewRecorder()
	h.Validate(rec, withChiID(r, id))
	return rec
}

// TestCredentialValidate covers the live key check: unauthenticated, foreign
// id, working key, and rejected key.
func TestCredentialValidate(t *testing.T) {
	pingErr := func(err error) credentials.PingFunc {
		return func(context.Context, *external.LLM) error { return err }
	}

	t.Run("unauthenticated is rejected", func(t *testing.T) {
		h, id := newValidateHandler(t, pingErr(nil))
		rec := doValidate(t, h, "", id)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("foreign id is not found", func(t *testing.T) {
		h, id := newValidateHandler(t, pingErr(nil))
		rec := doValidate(t, h, "u2", id)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404; body %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), testModelKey) {
			t.Fatalf("404 leaked the key: %s", rec.Body.String())
		}
	})

	tests := []struct {
		name       string
		ping       credentials.PingFunc
		wantValid  bool
		wantReason string
	}{
		{name: "working key", ping: pingErr(nil), wantValid: true},
		{name: "rejected key", ping: pingErr(errors.New("provider returned status 401")), wantReason: "invalid_key"},
		{name: "provider down", ping: pingErr(errors.New("dial: connection refused")), wantReason: "unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, id := newValidateHandler(t, tt.ping)
			rec := doValidate(t, h, "u1", id)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), testModelKey) {
				t.Fatalf("response leaked the key: %s", rec.Body.String())
			}
			var got validateResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.Valid != tt.wantValid || got.Reason != tt.wantReason {
				t.Fatalf("got %+v, want valid=%v reason=%q", got, tt.wantValid, tt.wantReason)
			}
		})
	}
}
