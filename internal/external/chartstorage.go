package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ChartImageStore persists chart images outside the database.
//
// Image bytes used to live in charts.image_blob, which meant every chart image
// grew the SQLite file that gets wiped on each deploy anyway. External storage
// keeps the database small and lets images outlive one container.
//
// The interface exists so a deployment with no external storage configured
// keeps working against the database column. It is not an abstraction for its
// own sake: two real backends, and the nil case is a supported configuration.
type ChartImageStore interface {
	// Put stores image under key and returns the object path. Replacing an
	// existing key must succeed.
	Put(ctx context.Context, key string, image []byte, contentType string) (string, error)

	// Get returns the stored image, or an error matching ErrImageNotFound.
	Get(ctx context.Context, key string) ([]byte, string, error)

	// Delete removes an object. Deleting a missing object is not an error.
	Delete(ctx context.Context, key string) error
}

// ErrImageNotFound reports a missing object, distinct from a transport failure.
var ErrImageNotFound = fmt.Errorf("chart image not found")

// supabaseStore implements ChartImageStore against Supabase Storage.
type supabaseStore struct {
	baseURL string
	bucket  string
	apiKey  string
	http    *http.Client
}

// NewSupabaseChartImageStore builds a store from Supabase Storage credentials.
//
// secretKey must be the project's secret key, never the publishable one. The
// bucket is private and every read is authorized with the secret key, because
// chart images can expose a figure from an unpublished paper.
func NewSupabaseChartImageStore(projectRef, bucket, secretKey string) (ChartImageStore, error) {
	if projectRef == "" || bucket == "" || secretKey == "" {
		return nil, fmt.Errorf("supabase: project ref, bucket, and secret key are all required")
	}
	if strings.HasPrefix(secretKey, "sb_publishable_") {
		// A publishable key would silently produce 401s on every read.
		return nil, fmt.Errorf("supabase: refusing to build a store with a publishable key; the secret key is required")
	}
	if strings.HasPrefix(projectRef, "http://") || strings.HasPrefix(projectRef, "https://") {
		return nil, fmt.Errorf("supabase: project ref must be a bare host, not a URL")
	}
	return &supabaseStore{
		baseURL: "https://" + projectRef + ".supabase.co",
		bucket:  bucket,
		apiKey:  secretKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ChartImageStoreFromEnv returns a store built from the environment, or nil when
// unconfigured.
//
// A nil store is the supported "no external storage" configuration: callers
// fall back to the database column rather than failing. That keeps local
// development and any deployment without Supabase working unchanged.
func ChartImageStoreFromEnv() (ChartImageStore, error) {
	projectRef := os.Getenv("SUPABASE_PROJECT_REF")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	secretKey := os.Getenv("SECRET_KEY_SUPABASE")
	if projectRef == "" || bucket == "" || secretKey == "" {
		return nil, nil
	}
	return NewSupabaseChartImageStore(projectRef, bucket, secretKey)
}

// Put uploads an image, overwriting any object already at key.
func (s *supabaseStore) Put(ctx context.Context, key string, image []byte, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	path := s.bucket + "/" + key
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.baseURL+"/storage/v1/object/"+path, bytes.NewReader(image))
	if err != nil {
		return "", fmt.Errorf("supabase: build upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase: upload %s: %w", path, err)
	}
	defer drainAndClose(resp)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("supabase: upload %s: %s", path, statusError(resp))
	}
	return path, nil
}

// Get downloads an image and its content type.
func (s *supabaseStore) Get(ctx context.Context, key string) ([]byte, string, error) {
	path := s.bucket + "/" + key
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/storage/v1/object/"+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("supabase: build download request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("apikey", s.apiKey)

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("supabase: download %s: %w", path, err)
	}
	defer drainAndClose(resp)

	// Supabase answers 400 for a missing object, not 404. Verified against the
	// live bucket: without this, a deleted image would surface as a 500 rather
	// than a 404, and the handler could not tell "gone" from "broken".
	if resp.StatusCode == http.StatusNotFound || isObjectMissing(resp) {
		return nil, "", ErrImageNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("supabase: download %s: %s", path, statusError(resp))
	}

	// Bound the read so a corrupt or unexpectedly large object cannot exhaust
	// memory in a process that also holds the database.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, "", fmt.Errorf("supabase: read %s: %w", path, err)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	return body, ct, nil
}

// Delete removes an object. A missing object is not an error, because cleanup
// must be idempotent.
func (s *supabaseStore) Delete(ctx context.Context, key string) error {
	path := s.bucket + "/" + key
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.baseURL+"/storage/v1/object/"+path, nil)
	if err != nil {
		return fmt.Errorf("supabase: build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("apikey", s.apiKey)

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: delete %s: %w", path, err)
	}
	defer drainAndClose(resp)

	if resp.StatusCode == http.StatusNotFound || isObjectMissing(resp) {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase: delete %s: %s", path, statusError(resp))
	}
	return nil
}

// isObjectMissing reports whether a Supabase error response means the object
// simply does not exist.
//
// Supabase Storage returns 400 with this message for a missing key, which is
// not obvious from the status code alone, so the body is inspected rather than
// trusting the status. The body is consumed here; callers must call this before
// any other read.
func isObjectMissing(resp *http.Response) bool {
	if resp.StatusCode != http.StatusBadRequest {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return false
	}
	var parsed struct {
		StatusCode string `json:"statusCode"`
		Error      string `json:"error"`
		Message    string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) != nil {
		return false
	}
	haystack := strings.ToLower(parsed.StatusCode + " " + parsed.Error + " " + parsed.Message)
	return strings.Contains(haystack, "not found") || strings.Contains(haystack, "does not exist")
}

// statusError renders a non-2xx response without leaking the response body
// verbatim, which can echo request headers.
func statusError(resp *http.Response) string {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		if parsed.Message != "" {
			return fmt.Sprintf("%s: %s", resp.Status, parsed.Message)
		}
		if parsed.Error != "" {
			return fmt.Sprintf("%s: %s", resp.Status, parsed.Error)
		}
	}
	return resp.Status
}

// drainAndClose lets the connection return to the pool.
func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	_ = resp.Body.Close()
}
