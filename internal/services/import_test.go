package services

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"paperviz/internal/external"
)

func TestFetchByDOI(t *testing.T) {
	pdfContent := []byte("%PDF-1.4 fake pdf content for testing")
	nonPDFContent := []byte("This is not a PDF file")

	tests := []struct {
		name          string
		doi           string
		doiInfo       *external.DOIInfo
		doiErr        error
		oaURL         *string
		oaErr         error
		serverHandler http.HandlerFunc
		wantBytes     []byte
		wantErr       bool
		wantErrIs     error
	}{
		{
			name: "valid DOI with open access PDF",
			doi:  "10.1234/test.2024.001",
			doiInfo: &external.DOIInfo{
				Title:  "Test Paper",
				PDFUrl: "https://example.com/old.pdf",
			},
			doiErr: nil,
			oaURL:  strPtr("https://example.com/oa.pdf"),
			oaErr:  nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(pdfContent)
			},
			wantBytes: pdfContent,
		},
		{
			name:    "valid DOI with CrossRef PDF URL fallback",
			doi:     "10.1234/test.2024.002",
			doiInfo: &external.DOIInfo{PDFUrl: "https://example.com/crossref.pdf"},
			doiErr:  nil,
			oaURL:   nil,
			oaErr:   nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(pdfContent)
			},
			wantBytes: pdfContent,
		},
		{
			name:      "DOI not found in CrossRef",
			doi:       "10.9999/nonexistent",
			doiInfo:   nil,
			doiErr:    fmt.Errorf("crossref DOI 10.9999/nonexistent: %w", external.ErrDOINotFound),
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "DOI not found via Unpaywall",
			doi:       "10.1234/test.2024.003",
			doiInfo:   &external.DOIInfo{Title: "Paywalled Paper"},
			doiErr:    nil,
			oaURL:     nil,
			oaErr:     fmt.Errorf("unpaywall DOI 10.1234/test.2024.003: %w", external.ErrDOINotFound),
			wantErr:   true,
			wantErrIs: ErrDOINotFound,
		},
		{
			name:      "paywall — no PDF URL from any source",
			doi:       "10.1234/test.2024.004",
			doiInfo:   &external.DOIInfo{Title: "Paywalled Paper"},
			doiErr:    nil,
			oaURL:     nil,
			oaErr:     nil,
			wantErr:   true,
			wantErrIs: ErrPaywallBlocked,
		},
		{
			name:      "invalid DOI format",
			doi:       "not-a-doi",
			wantErr:   true,
			wantErrIs: ErrInvalidDOI,
		},
		{
			name:      "empty DOI",
			doi:       "",
			wantErr:   true,
			wantErrIs: ErrInvalidDOI,
		},
		{
			name: "downloaded file is not PDF",
			doi:  "10.1234/test.2024.005",
			doiInfo: &external.DOIInfo{
				Title:  "HTML Paper",
				PDFUrl: "https://example.com/notpdf",
			},
			doiErr: nil,
			oaURL:  nil,
			oaErr:  nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write(nonPDFContent)
			},
			wantErr:   true,
			wantErrIs: ErrNotPDF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origDOI := fetchDOIInfo
			origOA := fetchOpenAccessPDF
			origClient := importHTTPClient
			defer func() {
				fetchDOIInfo = origDOI
				fetchOpenAccessPDF = origOA
				importHTTPClient = origClient
			}()

			fetchDOIInfo = func(doi string) (*external.DOIInfo, error) {
				return tt.doiInfo, tt.doiErr
			}
			fetchOpenAccessPDF = func(doi string) (*string, error) {
				return tt.oaURL, tt.oaErr
			}

			if tt.serverHandler != nil {
				ts := httptest.NewTLSServer(tt.serverHandler)
				defer ts.Close()
				importHTTPClient = ts.Client()

				if tt.doiInfo != nil && tt.doiInfo.PDFUrl != "" {
					tt.doiInfo.PDFUrl = ts.URL
				}
				if tt.oaURL != nil {
					oa := ts.URL
					tt.oaURL = &oa
				}
			}

			got, err := FetchByDOI(tt.doi)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("want errors.Is(err, %v), got: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytesEqual(got, tt.wantBytes) {
				t.Errorf("got %d bytes, want %d bytes", len(got), len(tt.wantBytes))
			}
		})
	}
}

func TestFetchByURL(t *testing.T) {
	pdfContent := []byte("%PDF-1.4 fake pdf content")
	nonPDFContent := []byte("This is not a PDF")

	tests := []struct {
		name          string
		rawURL        string
		serverHandler http.HandlerFunc
		wantBytes     []byte
		wantErr       bool
		wantErrIs     error
	}{
		{
			name:   "generic HTTPS PDF",
			rawURL: "https://example.com/paper.pdf",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(pdfContent)
			},
			wantBytes: pdfContent,
		},
		{
			name:   "non-PDF content-type rejected",
			rawURL: "https://example.com/notpdf",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write(nonPDFContent)
			},
			wantErr:   true,
			wantErrIs: ErrNotPDF,
		},
		{
			name:      "HTTP scheme rejected (SSRF)",
			rawURL:    "http://example.com/paper.pdf",
			wantErr:   true,
			wantErrIs: ErrInvalidURL,
		},
		{
			name:      "loopback IP blocked",
			rawURL:    "https://127.0.0.1/paper.pdf",
			wantErr:   true,
			wantErrIs: ErrSSRFBlocked,
		},
		{
			name:      "private IP 10.x blocked",
			rawURL:    "https://10.0.0.1/paper.pdf",
			wantErr:   true,
			wantErrIs: ErrSSRFBlocked,
		},
		{
			name:      "private IP 192.168.x blocked",
			rawURL:    "https://192.168.1.1/paper.pdf",
			wantErr:   true,
			wantErrIs: ErrSSRFBlocked,
		},
		{
			name:      "private IP 172.16.x blocked",
			rawURL:    "https://172.16.0.1/paper.pdf",
			wantErr:   true,
			wantErrIs: ErrSSRFBlocked,
		},
		{
			name:      "empty URL rejected",
			rawURL:    "",
			wantErr:   true,
			wantErrIs: ErrInvalidURL,
		},
		{
			name:    "network error on download",
			rawURL:  "https://nonexistent.invalid/paper.pdf",
			wantErr: true,
		},
		{
			name:   "404 response rejected",
			rawURL: "https://example.com/missing.pdf",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			},
			wantErr:   true,
			wantErrIs: ErrNotPDF,
		},
		{
			name:   "valid PDF via direct download",
			rawURL: "https://example.com/direct.pdf",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(pdfContent)
			},
			wantBytes: pdfContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origClient := importHTTPClient
			origIsPrivate := isPrivateIP
			defer func() {
				importHTTPClient = origClient
				isPrivateIP = origIsPrivate
			}()

			if tt.serverHandler != nil {
				ts := httptest.NewTLSServer(tt.serverHandler)
				defer ts.Close()
				importHTTPClient = ts.Client()
				isPrivateIP = func(hostname string) error { return nil } // skip SSRF for test server

				if tt.rawURL != "" {
					parsed, err := url.Parse(tt.rawURL)
					if err == nil {
						parsed.Host = ts.Listener.Addr().String()
						tt.rawURL = parsed.String()
					}
				}
			}

			got, err := FetchByURL(tt.rawURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("want errors.Is(err, %v), got: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytesEqual(got, tt.wantBytes) {
				t.Errorf("got %d bytes, want %d bytes", len(got), len(tt.wantBytes))
			}
		})
	}
}

func TestBlockPrivateIP(t *testing.T) {
	tests := []struct {
		hostname string
		wantErr  bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"172.16.0.1", true},
		{"localhost", true}, // resolves to ::1 (IPv6 loopback)
		{"example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			err := blockPrivateIP(tt.hostname)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %s, got nil", tt.hostname)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %s: %v", tt.hostname, err)
			}
		})
	}
}

func TestIsPDFContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{"application/pdf", true},
		{"application/pdf; charset=utf-8", true},
		{"TEXT/PDF", true},
		{"text/html", false},
		{"image/png", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ct, func(t *testing.T) {
			got := isPDFContentType(tt.ct)
			if got != tt.want {
				t.Errorf("isPDFContentType(%q) = %v, want %v", tt.ct, got, tt.want)
			}
		})
	}
}

func TestResolveArxivURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://arxiv.org/abs/2301.00001", "https://arxiv.org/pdf/2301.00001.pdf"},
		{"https://arxiv.org/pdf/2301.00001", "https://arxiv.org/pdf/2301.00001.pdf"},
		{"https://arxiv.org/pdf/2301.00001.pdf", "https://arxiv.org/pdf/2301.00001.pdf"},
		{"https://arxiv.org/html/2301.00001v1", "https://arxiv.org/pdf/2301.00001v1.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got, err := resolveArxivURL(parsed)
			if err != nil {
				t.Fatalf("resolveArxivURL: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
