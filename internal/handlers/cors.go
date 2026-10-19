package handlers

import (
	"net/http"
	"os"
	"strings"
)

// corsMiddleware lets the separately deployed frontend (Cloudflare Pages) call
// this API cross-origin. A browser sends Origin on cross-origin XHR and Go adds
// nothing, so without this every call from the Pages origin fails preflight.
// Origins are allowlisted rather than reflected: ALLOWED_ORIGINS is
// comma-separated, plus the local Vite dev server.
func corsMiddleware(next http.Handler) http.Handler {
	allowed := map[string]bool{"http://localhost:5173": true}
	for _, origin := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowed[origin] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && allowed[origin] {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
