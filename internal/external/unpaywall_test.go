package external

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchOpenAccessPDF(t *testing.T) {
	tests := []struct {
		name      string
		doi       string
		status    int
		body      interface{}
		wantURL   *string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "open access with PDF URL",
			doi:    "10.1234/oapaper.2024",
			status: http.StatusOK,
			body: unpaywallEnvelope{
				Doi:  "10.1234/oapaper.2024",
				IsOa: true,
				BestOaLocation: &unpaywallOaLocation{
					UrlForPdf: "https://example.com/oa.pdf",
					Url:       "https://example.com/oa-page",
				},
			},
			wantURL: strPtr("https://example.com/oa.pdf"),
		},
		{
			name:   "open access with page URL only",
			doi:    "10.1234/pageonly.2024",
			status: http.StatusOK,
			body: unpaywallEnvelope{
				Doi:  "10.1234/pageonly.2024",
				IsOa: true,
				BestOaLocation: &unpaywallOaLocation{
					UrlForPdf: "",
					Url:       "https://example.com/oa-page",
				},
			},
			wantURL: strPtr("https://example.com/oa-page"),
		},
		{
			name:   "not open access",
			doi:    "10.1234/closed.2024",
			status: http.StatusOK,
			body: unpaywallEnvelope{
				Doi:  "10.1234/closed.2024",
				IsOa: false,
			},
			wantURL: nil,
		},
		{
			name:      "DOI not found",
			doi:       "10.9999/notfound.2024",
			status:    http.StatusNotFound,
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "DOI gone",
			doi:       "10.1111/gone.2024",
			status:    http.StatusGone,
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "empty DOI",
			doi:       "",
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "invalid DOI format",
			doi:       "not-a-doi",
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "network error",
			doi:       "10.1234/down.2024",
			wantErr:   true,
			wantErrIs: ErrNetworkError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "network error" {
				orig := unpaywallHTTPClient
				unpaywallHTTPClient = &http.Client{
					Timeout: 2 * time.Second,
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("connection refused")
					}),
				}
				t.Cleanup(func() { unpaywallHTTPClient = orig })
			}

			if tt.name != "empty DOI" && tt.name != "invalid DOI format" && tt.name != "network error" {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					if tt.body != nil {
						json.NewEncoder(w).Encode(tt.body)
					}
				}))
				origTransport := http.DefaultTransport
				if unpaywallHTTPClient.Transport != nil {
					origTransport = unpaywallHTTPClient.Transport
				}
				unpaywallHTTPClient = ts.Client()
				unpaywallHTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
					req.URL.Scheme = "http"
					req.URL.Host = ts.Listener.Addr().String()
					return origTransport.RoundTrip(req)
				})
				t.Cleanup(func() {
					ts.Close()
					unpaywallHTTPClient = &http.Client{Timeout: 10 * time.Second}
				})
			}

			pdfUrl, err := FetchOpenAccessPDF(tt.doi)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("expected error wrapping %v, got: %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantURL == nil {
				if pdfUrl != nil {
					t.Errorf("expected nil URL, got %q", *pdfUrl)
				}
			} else {
				if pdfUrl == nil {
					t.Fatalf("expected URL %q, got nil", *tt.wantURL)
				}
				if *pdfUrl != *tt.wantURL {
					t.Errorf("URL = %q, want %q", *pdfUrl, *tt.wantURL)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
