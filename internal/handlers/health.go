package handlers

import (
	"database/sql"
	"net/http"
)

// healthzHandler returns 200 when DB is reachable, 503 otherwise.
func healthzHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "error": "db_unreachable"})
			return
		}
		if _, err := db.Exec("SELECT 1"); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "error": "db_unreachable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
