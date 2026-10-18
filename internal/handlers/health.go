package handlers

import (
	"database/sql"
	"net/http"
	"runtime/debug"
)

// apiRootHandler answers service-discovery probes on /. This process serves no
// files, so / is the only place an uptime check or a person learns what it is.
func apiRootHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "PaperViz API",
		"version": buildVersion(),
	})
}

// buildVersion reads the version the linker stamped from the git tag, so the
// binary reports one version per release instead of a constant that goes stale.
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

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
