package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequireIngestionEnabled(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		wantStatus int
	}{
		{"enabled by default", "", http.StatusCreated},
		{"explicitly enabled", "true", http.StatusCreated},
		{"kill switch off", "false", http.StatusServiceUnavailable},
		{"zero is off", "0", http.StatusServiceUnavailable},
		// An unparseable value must not silently stop intake, or a typo takes
		// the product down with no obvious cause.
		{"garbage stays enabled", "maybe", http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("INGESTION_ENABLED", tt.value)

			reached := false
			h := requireIngestionEnabled(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
				w.WriteHeader(http.StatusCreated)
			}))

			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/documents", nil))

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if reached != (tt.wantStatus == http.StatusCreated) {
				t.Fatalf("downstream reached = %v, want %v", reached, tt.wantStatus == http.StatusCreated)
			}
		})
	}
}

// TestRequireIngestionEnabledSendsRetryAfter proves a client told to wait has a
// hint for how long, rather than an open-ended 503 to hammer.
func TestRequireIngestionEnabledSendsRetryAfter(t *testing.T) {
	t.Setenv("INGESTION_ENABLED", "false")

	h := requireIngestionEnabled(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("downstream must not run when ingestion is disabled")
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/documents", nil))

	if got := w.Header().Get("Retry-After"); got == "" {
		t.Fatal("missing Retry-After on a 503")
	}
	if body := w.Body.String(); !strings.Contains(body, "ingestion_disabled") {
		t.Fatalf("body = %q, want the ingestion_disabled code", body)
	}
}
