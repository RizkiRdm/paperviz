package external

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// fakeStore records calls so tests can assert storage behaviour without a network.
type fakeStore struct {
	puts    map[string][]byte
	deleted []string
	putErr  error
	getErr  error
}

func newFakeStore() *fakeStore {
	return &fakeStore{puts: map[string][]byte{}}
}

func (f *fakeStore) Put(_ context.Context, key string, image []byte, _ string) (string, error) {
	if f.putErr != nil {
		return "", f.putErr
	}
	f.puts[key] = image
	return key, nil
}

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, string, error) {
	if f.getErr != nil {
		return nil, "", f.getErr
	}
	b, ok := f.puts[key]
	if !ok {
		return nil, "", ErrImageNotFound
	}
	return b, "image/png", nil
}

func (f *fakeStore) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	delete(f.puts, key)
	return nil
}

// TestChartImageStoreFromEnv covers the nil case, which is a supported
// deployment: no external storage configured means the database column is
// used instead of the service failing to start.
func TestChartImageStoreFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		ref        string
		bucket     string
		secret     string
		wantNil    bool
		wantErrStr string
	}{
		{name: "all unset falls back to the database", wantNil: true},
		{name: "ref missing", bucket: "b", secret: "sb_secret_x", wantNil: true},
		{name: "bucket missing", ref: "abc", secret: "sb_secret_x", wantNil: true},
		{name: "secret missing", ref: "abc", bucket: "b", wantNil: true},
		{name: "fully configured", ref: "abc", bucket: "b", secret: "sb_secret_x"},
		{
			name:       "publishable key is refused",
			ref:        "abc",
			bucket:     "b",
			secret:     "sb_publishable_x",
			wantErrStr: "publishable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, "SUPABASE_PROJECT_REF", tt.ref)
			setEnv(t, "SUPABASE_STORAGE_BUCKET", tt.bucket)
			setEnv(t, "SECRET_KEY_SUPABASE", tt.secret)

			store, err := ChartImageStoreFromEnv()
			if tt.wantErrStr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrStr) {
					t.Fatalf("err = %v, want one containing %q", err, tt.wantErrStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && store != nil {
				t.Errorf("expected nil store, got %#v", store)
			}
			if !tt.wantNil && store == nil {
				t.Error("expected a store, got nil")
			}
		})
	}
}

// TestNewSupabaseChartImageStoreRejectsURLRef guards against a project ref
// pasted as a full URL, which would produce a malformed base URL at request
// time instead of at startup.
func TestNewSupabaseChartImageStoreRejectsURLRef(t *testing.T) {
	_, err := NewSupabaseChartImageStore("https://abc.supabase.co", "b", "sb_secret_x")
	if err == nil {
		t.Fatal("expected an error for a URL-shaped project ref")
	}
}

// TestSupabaseStorePutGetDelete exercises the real request shapes against a
// stub server, including the upsert header Storage requires to replace an
// object. Without x-upsert a replacement silently creates a duplicate instead.
func TestSupabaseStorePutGetDelete(t *testing.T) {
	var (
		gotUpsert   string
		gotAuth     string
		gotMime     string
		gotPath     string
		stored      []byte
		deletedPath string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			gotUpsert = r.Header.Get("x-upsert")
			gotAuth = r.Header.Get("Authorization")
			gotMime = r.Header.Get("Content-Type")
			gotPath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			stored = b
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key":"chart-images/a.png"}`))
		case http.MethodGet:
			if len(stored) == 0 {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message":"Object not found"}`))
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.Write(stored)
		case http.MethodDelete:
			deletedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Deleted"}`))
		}
	}))
	defer srv.Close()

	store := newStoreForTest(srv.URL)
	ctx := context.Background()
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00}

	path, err := store.Put(ctx, "a.png", png, "image/png")
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if path != "chart-images/a.png" {
		t.Errorf("path = %q, want chart-images/a.png", path)
	}
	if gotUpsert != "true" {
		t.Errorf("x-upsert = %q, want true; without it a replacement silently duplicates", gotUpsert)
	}
	if gotAuth != "Bearer sb_secret_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotMime != "image/png" {
		t.Errorf("Content-Type = %q", gotMime)
	}
	if gotPath != "/storage/v1/object/chart-images/a.png" {
		t.Errorf("path = %q", gotPath)
	}

	body, ct, err := store.Get(ctx, "a.png")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !bytes.Equal(body, png) {
		t.Errorf("round-trip body mismatch")
	}
	if ct != "image/png" {
		t.Errorf("content-type = %q", ct)
	}

	// A missing object must report ErrImageNotFound, not a generic failure, so
	// the handler can answer 404 instead of 500.
	stored = nil
	if _, _, err := store.Get(ctx, "a.png"); !errors.Is(err, ErrImageNotFound) {
		t.Errorf("get missing = %v, want ErrImageNotFound", err)
	}

	if err := store.Delete(ctx, "a.png"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if deletedPath != "/storage/v1/object/chart-images/a.png" {
		t.Errorf("delete path = %q", deletedPath)
	}
}

// TestSupabaseStoreDeleteMissingIsNotAnError keeps cleanup idempotent.
func TestSupabaseStoreDeleteMissingIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Object not found"}`))
	}))
	defer srv.Close()

	if err := newStoreForTest(srv.URL).Delete(context.Background(), "gone.png"); err != nil {
		t.Errorf("deleting a missing object returned %v, want nil", err)
	}
}

// TestSupabaseStoreSurfacesErrorMessages confirms a real API failure is
// reported, and that the request headers are not echoed back in the error.
func TestSupabaseStoreSurfacesErrorMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"new row violates row-level security policy"}`))
	}))
	defer srv.Close()

	_, err := newStoreForTest(srv.URL).Put(context.Background(), "a.png", []byte("x"), "image/png")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "row-level security") {
		t.Errorf("error lost the API message: %v", err)
	}
	if strings.Contains(err.Error(), "sb_secret_test") {
		t.Error("error leaked the secret key")
	}
}

func newStoreForTest(baseURL string) *supabaseStore {
	return &supabaseStore{
		baseURL: baseURL,
		bucket:  "chart-images",
		apiKey:  "sb_secret_test",
		http:    &http.Client{},
	}
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if value == "" {
		os.Unsetenv(key)
		return
	}
	t.Setenv(key, value)
}

// TestSupabaseStoreMissingObjectIs400NotFound pins Supabase's real behaviour.
// Storage answers a missing key with 400 and a "not found" body rather than 404,
// which is why isObjectMissing exists. Without a test for this, the handler
// would answer 500 for a deleted image and the distinction between "gone" and
// "broken" would be lost.
func TestSupabaseStoreMissingObjectIs400NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"statusCode":"404","error":"Not found","message":"Object not found"}`))
	}))
	defer srv.Close()

	store := newStoreForTest(srv.URL)
	if _, _, err := store.Get(context.Background(), "gone.png"); !errors.Is(err, ErrImageNotFound) {
		t.Errorf("get missing = %v, want ErrImageNotFound", err)
	}
	if err := store.Delete(context.Background(), "gone.png"); err != nil {
		t.Errorf("delete missing = %v, want nil", err)
	}
}

// TestSupabaseStoreOtherBadRequestIsNotTreatedAsMissing keeps a real fault from
// being reported as a missing image, which would hide a broken configuration.
func TestSupabaseStoreOtherBadRequestIsNotTreatedAsMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid bucket name"}`))
	}))
	defer srv.Close()

	if _, _, err := newStoreForTest(srv.URL).Get(context.Background(), "a.png"); errors.Is(err, ErrImageNotFound) {
		t.Error("a configuration error was reported as a missing object")
	}
}
