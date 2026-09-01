package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"paperviz/internal/external"
)

// mockImportService implements ImportService for testing.
type mockImportService struct {
	fetchByDOIFunc func(doi string) (string, string, error)
	fetchByURLFunc func(url string) (string, string, error)
}

// FetchByDOI delegates to the mock function.
func (m *mockImportService) FetchByDOI(doi string) (string, string, error) {
	return m.fetchByDOIFunc(doi)
}

// FetchByURL delegates to the mock function.
func (m *mockImportService) FetchByURL(url string) (string, string, error) {
	return m.fetchByURLFunc(url)
}

func TestImportByDOI(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(doi string) (string, string, error)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "empty_doi",
			body:       ImportByDOIRequest{DOI: "", ReadingLevel: "simplified"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "doi_required",
		},
		{
			name:       "invalid_doi_format",
			body:       ImportByDOIRequest{DOI: "not-a-doi", ReadingLevel: "simplified"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_doi",
		},
		{
			name:       "invalid_reading_level",
			body:       ImportByDOIRequest{DOI: "10.1234/test", ReadingLevel: "advanced"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_reading_level",
		},
		{
			name: "service_error",
			body: ImportByDOIRequest{DOI: "10.1234/test", ReadingLevel: "simplified"},
			mockFunc: func(doi string) (string, string, error) {
				return "", "", external.ErrDOINotFound
			},
			wantStatus: http.StatusBadGateway,
			wantCode:   "fetch_failed",
		},
		{
			name:       "nil_service",
			body:       ImportByDOIRequest{DOI: "10.1234/test", ReadingLevel: "simplified"},
			wantStatus: http.StatusNotImplemented,
			wantCode:   "import_not_available",
		},
		{
			name:       "invalid_json",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/import/doi", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			var handler *ImportHandler
			if tt.name == "nil_service" {
				handler = NewImportHandler(nil, nil)
			} else {
				mock := &mockImportService{fetchByDOIFunc: tt.mockFunc}
				handler = NewImportHandler(nil, nil, mock)
			}

			handler.ImportByDOI(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != "" {
				var resp errorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if resp.Error != tt.wantCode {
					t.Errorf("error = %q; want %q", resp.Error, tt.wantCode)
				}
			}
		})
	}
}

func TestImportByURL(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(url string) (string, string, error)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "empty_url",
			body:       ImportByURLRequest{URL: "", ReadingLevel: "simplified"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "url_required",
		},
		{
			name:       "invalid_url_no_scheme",
			body:       ImportByURLRequest{URL: "example.com/paper.pdf", ReadingLevel: "simplified"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_url",
		},
		{
			name:       "invalid_reading_level",
			body:       ImportByURLRequest{URL: "https://example.com/paper.pdf", ReadingLevel: "advanced"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_reading_level",
		},
		{
			name: "service_error",
			body: ImportByURLRequest{URL: "https://example.com/paper.pdf", ReadingLevel: "simplified"},
			mockFunc: func(url string) (string, string, error) {
				return "", "", external.ErrNoOpenAccess
			},
			wantStatus: http.StatusBadGateway,
			wantCode:   "fetch_failed",
		},
		{
			name:       "nil_service",
			body:       ImportByURLRequest{URL: "https://example.com/paper.pdf", ReadingLevel: "simplified"},
			wantStatus: http.StatusNotImplemented,
			wantCode:   "import_not_available",
		},
		{
			name:       "invalid_json",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/import/url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			var handler *ImportHandler
			if tt.name == "nil_service" {
				handler = NewImportHandler(nil, nil)
			} else {
				mock := &mockImportService{fetchByURLFunc: tt.mockFunc}
				handler = NewImportHandler(nil, nil, mock)
			}

			handler.ImportByURL(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != "" {
				var resp errorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if resp.Error != tt.wantCode {
					t.Errorf("error = %q; want %q", resp.Error, tt.wantCode)
				}
			}
		})
	}
}

func TestImportByDOIWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/import/doi", bytes.NewReader([]byte(`{"doi":"  ","reading_level":"simplified"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mock := &mockImportService{fetchByDOIFunc: func(doi string) (string, string, error) {
		t.Fatal("should not be called")
		return "", "", nil
	}}
	handler := NewImportHandler(nil, nil, mock)
	handler.ImportByDOI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "doi_required" {
		t.Errorf("error = %q; want doi_required", resp.Error)
	}
}

func TestImportByURLWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/import/url", bytes.NewReader([]byte(`{"url":"  ","reading_level":"simplified"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mock := &mockImportService{fetchByURLFunc: func(url string) (string, string, error) {
		t.Fatal("should not be called")
		return "", "", nil
	}}
	handler := NewImportHandler(nil, nil, mock)
	handler.ImportByURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "url_required" {
		t.Errorf("error = %q; want url_required", resp.Error)
	}
}
