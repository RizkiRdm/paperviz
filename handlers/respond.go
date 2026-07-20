package handlers

import (
	"encoding/json"
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
