package handlers

import (
	"net/http"

	"paperviz/internal/services"
)

// requireIngestionEnabled refuses document creation when the INGESTION_ENABLED
// kill switch is off.
//
// This is the panic button. It exists so ingestion can be stopped during an
// abuse spike or a runaway queue without a deploy and without taking down the
// read paths — browsing, sharing, and libraries keep working. It complements,
// and does not replace, the per-IP rate limiter: a limiter sheds one client,
// this stops intake entirely.
func requireIngestionEnabled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !services.IngestionEnabled() {
			w.Header().Set("Retry-After", "3600")
			writeError(w, http.StatusServiceUnavailable, "ingestion_disabled")
			return
		}
		next.ServeHTTP(w, r)
	})
}
