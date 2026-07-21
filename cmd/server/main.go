// Command server is PaperViz's entrypoint. It loads configuration from
// environment variables, opens the SQLite database, wires the HTTP router,
// starts the background expiry sweep, and serves. This file intentionally
// contains almost no logic of its own — everything it calls lives in
// repository/, external/, handlers/, and services/, per ARCHITECTURE.md's
// layered design. Read this file to understand *what starts up*, then go
// read the layer packages to understand *how each piece works*.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"paperviz/external"
	"paperviz/handlers"
	"paperviz/repository"
	"paperviz/services"
)

func main() {
	// Structured JSON logging to stdout, per ARCHITECTURE.md Section 4
	// Logging Policy. Using this as the default logger (not just for one
	// package) so every log line across the app has consistent shape.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

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

	migrationSQL, err := repository.ReadMigration("migrations/001_init.sql")
	if err != nil {
		// A failed DB connection/schema load at startup is one of the few
		// cases AGENTS.md allows a panic-equivalent (os.Exit) for — see
		// Coding Conventions: "Panics reserved for truly unrecoverable
		// states (e.g., failed DB connection at startup)."
		slog.Error("failed to read migration file", "error", err)
		os.Exit(1)
	}

	db, err := repository.Open(dbPath, migrationSQL)
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
