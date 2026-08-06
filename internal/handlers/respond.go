package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// errorResponse is the wire format for all error replies: { "error": "<snake_case_code>" }.
// STRICT: never include raw Go error strings or stack traces (ARCHITECTURE.md Internal Contracts).
type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response failed", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, errorResponse{Error: code})
}

// readJSON decodes the request body into dst. Returns an error if the body
// is not valid JSON or exceeds 1MB.
func readJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1MB limit
	defer func() {
		// Drain and close the body to allow connection reuse
		_, _ = io.Copy(io.Discard, r.Body)
		r.Body.Close()
	}()
	return json.NewDecoder(r.Body).Decode(dst)
}
