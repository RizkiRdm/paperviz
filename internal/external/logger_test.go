package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"
)

// TestJSONLHandlerErrorRendering guards against the defect where an error
// attribute is JSON-marshaled as `{}` (error types have no exported fields),
// burying the actual failure message. See cmd/server/main.go:86 — startup
// errors were being logged as `"error":{}` which made "failed to open
// database" impossible to diagnose.
func TestJSONLHandlerErrorRendering(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "wrapped error", err: errors.New("apply migration 1: table documents already exists")},
		{name: "sentinel error", err: errors.New("boom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			h := NewJSONLHandler(&buf)

			rec := slog.NewRecord(time.Now(), slog.LevelError, "failed to open database", 0)
			rec.AddAttrs(slog.Any("error", tt.err))
			if err := h.Handle(context.Background(), rec); err != nil {
				t.Fatalf("Handle returned error: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
				t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
			}

			got, ok := m["error"]
			if !ok {
				t.Fatalf("output missing 'error' key: %s", buf.String())
			}

			switch v := got.(type) {
			case string:
				if v != tt.err.Error() {
					t.Errorf("error = %q; want %q", v, tt.err.Error())
				}
			default:
				t.Errorf("error field rendered as %T (%v); want error string %q", got, got, tt.err.Error())
			}
		})
	}
}

// TestJSONLHandlerWritesOtherAttrKinds ensures the handler still serializes
// non-error attrs (strings, ints, bools) without regression.
func TestJSONLHandlerWritesOtherAttrKinds(t *testing.T) {
	var buf bytes.Buffer
	h := NewJSONLHandler(&buf)

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "server starting", 0)
	rec.AddAttrs(
		slog.String("port", "8080"),
		slog.Int("documents_deleted", 1),
		slog.Bool("wal_enabled", true),
	)
	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
	}

	for key, want := range map[string]any{
		"port":              "8080",
		"documents_deleted": float64(1),
		"wal_enabled":       true,
	} {
		if m[key] != want {
			t.Errorf("key %q = %v (%T); want %v (%T)", key, m[key], m[key], want, want)
		}
	}
}
