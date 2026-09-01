package external

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchDOIInfo(t *testing.T) {
	tests := []struct {
		name       string
		doi        string
		status     int
		body       interface{}
		wantErr    bool
		wantErrIs  error
		wantTitle  string
		wantPDFUrl string
	}{
		{
			name:   "valid DOI with PDF link",
			doi:    "10.1234/test.2024",
			status: http.StatusOK,
			body: crossrefEnvelope{
				Status: "ok",
				Message: crossrefWork{
					Title:  []string{"A Test Paper"},
					Author: []crossrefAuthor{{Given: "Jane", Family: "Doe"}},
					Link: []crossrefLink{
						{URL: "https://example.com/paper.pdf", Type: "application/pdf"},
						{URL: "https://example.com/other", Type: "text/html"},
					},
					Abstract: "Test abstract",
				},
			},
			wantTitle:  "A Test Paper",
			wantPDFUrl: "https://example.com/paper.pdf",
		},
		{
			name:   "valid DOI no PDF link",
			doi:    "10.5678/no-pdf.2024",
			status: http.StatusOK,
			body: crossrefEnvelope{
				Status: "ok",
				Message: crossrefWork{
					Title:  []string{"No PDF Paper"},
					Author: []crossrefAuthor{{Given: "John", Family: "Smith"}},
				},
			},
			wantTitle: "No PDF Paper",
		},
		{
			name:      "DOI not found",
			doi:       "10.9999/notfound.2024",
			status:    http.StatusNotFound,
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
			name:      "network error (server unreachable)",
			doi:       "10.1234/server-down.2024",
			wantErr:   true,
			wantErrIs: ErrNetworkError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "network error (server unreachable)" {
				// Point client at a non-existent server
				orig := crossrefHTTPClient
				crossrefHTTPClient = &http.Client{
					Timeout: 10 * time.Second,
					Transport: &http.Transport{
						DialContext: nil, // will fail to dial
					},
				}
				t.Cleanup(func() { crossrefHTTPClient = orig })

				// Actually the simplest way is to point at an invalid URL via a custom round-tripper
				crossrefHTTPClient = &http.Client{
					Timeout: 2 * time.Second,
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("connection refused")
					}),
				}
				t.Cleanup(func() { crossrefHTTPClient = orig })
			}

			// Start test server for non-network-error cases
			if tt.name != "empty DOI" && tt.name != "invalid DOI format" && tt.name != "network error (server unreachable)" {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					if tt.body != nil {
						json.NewEncoder(w).Encode(tt.body)
					}
				}))
				defer ts.Close()
				crossrefHTTPClient = ts.Client()
				// Override the endpoint — we'll set URL directly in the test
				// Since FetchDOIInfo builds URL internally, we need a different approach.
				// Let's use a custom round-tripper that rewrites the host.
				origTransport := crossrefHTTPClient.Transport
				crossrefHTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
					req.URL.Scheme = "http"
					req.URL.Host = ts.Listener.Addr().String()
					return origTransport.RoundTrip(req)
				})
				t.Cleanup(func() {
					ts.Close()
					crossrefHTTPClient = &http.Client{Timeout: 10 * time.Second}
				})
			}

			info, err := FetchDOIInfo(tt.doi)

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
			if info.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", info.Title, tt.wantTitle)
			}
			if tt.wantPDFUrl != "" && info.PDFUrl != tt.wantPDFUrl {
				t.Errorf("PDFUrl = %q, want %q", info.PDFUrl, tt.wantPDFUrl)
			}
		})
	}
}

// roundTripFunc adapts a function to the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
