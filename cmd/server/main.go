// Command server is PaperViz's entrypoint. It loads configuration from
// environment variables, opens the SQLite database, wires the HTTP router,
// starts the background expiry sweep, and serves. This file intentionally
// contains almost no logic of its own — everything it calls lives in
// repository/, external/, handlers/, and services/, per ARCHITECTURE.md's
// layered design. Read this file to understand *what starts up*, then go
// read the layer packages to understand *how each piece works*.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"paperviz/internal/external"
	"paperviz/internal/handlers"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// loadMigrations reads every migration SQL file into a versioned map, in
// declaration order. Kept separate from main so tests can assert the full
// migration chain (including 004) is registered — a missing registration
// silently ships a schema the repository code already depends on.
// migrationsDir is relative to the process working directory (repo root in
// production; tests pass an absolute path since `go test` runs per-package).
func loadMigrations(migrationsDir string) (map[int]string, error) {
	migrations := make(map[int]string)

	paths := map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		3:  "003_chapters.sql",
		4:  "004_chapter_charts.sql",
		5:  "005_evidence.sql",
		6:  "006_document_title.sql",
		7:  "007_saved_papers.sql",
		8:  "008_research_collections.sql",
		9:  "009_share_tokens.sql",
		10: "010_document_share.sql",
		11: "011_share_referrals.sql",
		12: "012_usage_analytics.sql",
	}

	for version, file := range paths {
		sql, err := repository.ReadMigration(filepath.Join(migrationsDir, file))
		if err != nil {
			return nil, fmt.Errorf("read migration %03d: %w", version, err)
		}
		migrations[version] = sql
	}
	return migrations, nil
}

func main() {
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "paperviz.log.jsonl"
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	slog.SetDefault(slog.New(external.NewJSONLHandler(io.MultiWriter(os.Stdout, f))))

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		// PROHIBITED per ARCHITECTURE.md Section 5: no hardcoded keys, no
		// keys in committed config files. If this is missing, fail loudly
		// at startup rather than silently running with a broken LLM client.
		slog.Error("GEMINI_API_KEY environment variable is required")
		os.Exit(1)
	}

	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		// Current fast + cheap default, good for text-only tasks. Valid
		// stable models as of 2026-07: gemini-3.5-flash, gemini-2.5-flash-lite.
		geminiModel = "gemini-2.5-flash-lite"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "paperviz.db"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend/dist"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	migrations, err := loadMigrations("migrations")
	if err != nil {
		slog.Error("failed to load migrations", "error", err)
		os.Exit(1)
	}

	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	gemini := external.NewGeminiClient(geminiAPIKey, geminiModel)

	// Start the expiry sweep in the background — runs once immediately,
	// then hourly (see services/expiry.go). It never returns, so it must
	// run in its own goroutine, not block main().
	go services.RunExpirySweepLoop(repository.NewDocumentRepo(db))

	router := handlers.NewRouter(db, gemini, staticDir)

	slog.Info("paperviz server starting", "port", port, "database_path", dbPath)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
