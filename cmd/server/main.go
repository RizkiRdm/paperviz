// Command server is PaperViz's entrypoint. It loads configuration from
// environment variables, opens the SQLite database, wires the HTTP router,
// starts the background expiry sweep, and serves. This file intentionally
// contains almost no logic of its own — everything it calls lives in
// repository/, external/, handlers/, and services/, per ARCHITECTURE.md's
// layered design. Read this file to understand *what starts up*, then go
// read the layer packages to understand *how each piece works*.
package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"paperviz/internal/external"
	"paperviz/internal/handlers"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

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

	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		slog.Error("GOOGLE_CLIENT_ID environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("GOOGLE_CLIENT_SECRET") == "" {
		slog.Error("GOOGLE_CLIENT_SECRET environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("GOOGLE_REDIRECT_URL") == "" {
		slog.Error("GOOGLE_REDIRECT_URL environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		slog.Error("STRIPE_SECRET_KEY environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("STRIPE_WEBHOOK_SECRET") == "" {
		slog.Error("STRIPE_WEBHOOK_SECRET environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("FRONTEND_URL") == "" {
		slog.Error("FRONTEND_URL environment variable is required")
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

	migrations, err := repository.LoadMigrations("migrations")
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

	sessionRepo := repository.NewSessionRepo(db)
	if _, err := sessionRepo.DeleteExpired(); err != nil {
		slog.Error("failed to delete expired sessions on startup", "error", err)
	}

	tr := external.NewTransport()
	gemini, err := tr.For(external.ProviderGemini, geminiAPIKey, geminiModel)
	if err != nil {
		slog.Error("failed to build llm client", "error", err)
		os.Exit(1)
	}

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
