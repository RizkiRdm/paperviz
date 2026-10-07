package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateDocument_PartialParseErrors asserts a bad multipart request is
// reported as what it is. Every parse failure used to return file_too_large,
// so a caller who sent plain JSON was told their file was too big.
func TestCreateDocument_PartialParseErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		setCT     bool
		ctype     string
		wantError string
	}{
		{
			name:      "json body is not multipart",
			body:      `{"text":"some paper"}`,
			setCT:     true,
			ctype:     "application/json",
			wantError: "invalid_content_type",
		},
		{
			name:      "form encoded body is not multipart",
			body:      "text=some+paper",
			setCT:     true,
			ctype:     "application/x-www-form-urlencoded",
			wantError: "invalid_content_type",
		},
		{
			name:      "multipart with malformed boundary is a bad request",
			body:      "--broken\r\nnot a real part\r\n",
			setCT:     true,
			ctype:     "multipart/form-data; boundary=broken",
			wantError: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/documents", strings.NewReader(tt.body))
			if tt.setCT {
				req.Header.Set("Content-Type", tt.ctype)
			}
			rec := httptest.NewRecorder()

			// A nil DB is safe here: every case is rejected during parsing,
			// before the handler reaches the service layer.
			NewDocumentHandler(nil, nil).Create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %s)", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.wantError) {
				t.Errorf("body %q does not report %q", rec.Body.String(), tt.wantError)
			}
			if strings.Contains(rec.Body.String(), "file_too_large") {
				t.Errorf("body %q wrongly reports file_too_large", rec.Body.String())
			}
		})
	}
}

// zeroReader streams filler bytes without allocating them.
type zeroReader struct {
	remaining int64
}

func (z *zeroReader) Read(p []byte) (int, error) {
	if z.remaining <= 0 {
		return 0, io.EOF
	}
	n := int64(len(p))
	if n > z.remaining {
		n = z.remaining
	}
	for i := int64(0); i < n; i++ {
		p[i] = 'a'
	}
	z.remaining -= n
	return int(n), nil
}

// TestCreateDocument_OversizedBodyStillReportsTooLarge guards the other
// direction: the split must not turn a genuinely oversized body into a generic
// invalid_request.
//
// The declared length is what the handler reads, because the multipart parse
// error for an oversized body is whatever the reader hit first ("bufio: buffer
// full" in practice) and is indistinguishable from a malformed one.
func TestCreateDocument_OversizedBodyStillReportsTooLarge(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		&zeroReader{remaining: maxUploadBytes + 4096},
	)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=zzz")
	req.ContentLength = maxUploadBytes + 4096
	rec := httptest.NewRecorder()

	NewDocumentHandler(nil, nil).Create(rec, req)

	if !strings.Contains(rec.Body.String(), "file_too_large") {
		t.Errorf("body %q does not report file_too_large for an oversized upload", rec.Body.String())
	}
}
