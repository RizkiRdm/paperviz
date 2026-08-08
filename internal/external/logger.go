package external

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type jsonlHandler struct {
	mu      sync.Mutex
	writers []io.Writer
	attrs   []slog.Attr
}

func NewJSONLHandler(writers ...io.Writer) slog.Handler {
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	return &jsonlHandler{writers: writers}
}

func (h *jsonlHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *jsonlHandler) Handle(_ context.Context, r slog.Record) error {
	m := make(map[string]any, 4+r.NumAttrs()+len(h.attrs))

	m["time"] = r.Time.Format(time.RFC3339Nano)
	m["message"] = r.Message
	m["severity"] = strings.ToLower(r.Level.String())

	for _, a := range h.attrs {
		m[a.Key] = a.Value.Any()
	}

	r.Attrs(func(a slog.Attr) bool {
		// Resolve LogValuer attrs, then render errors as their Error() string.
		// JSON-marshaling an error value directly produces {} because error
		// types expose no exported fields — symptoms like "failed to open
		// database" lost their cause entirely (see cmd/server/main.go).
		v := a.Value.Resolve()
		if err, ok := v.Any().(error); ok {
			m[a.Key] = err.Error()
			return true
		}
		m[a.Key] = v.Any()
		return true
	})

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, w := range h.writers {
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (h *jsonlHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &jsonlHandler{
		writers: h.writers,
		attrs:   append(h.attrs, attrs...),
	}
}

func (h *jsonlHandler) WithGroup(_ string) slog.Handler {
	return h
}
