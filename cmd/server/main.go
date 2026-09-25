// Command server is PaperViz's entrypoint. It loads configuration from
// environment variables, opens the SQLite database, wires the HTTP router,
// starts the background expiry sweep, and serves. This file intentionally
// contains almost no logic of its own — everything it calls lives in
// repository/, external/, handlers/, and services/, per ARCHITECTURE.md's
// layered design. Read this file to understand *what starts up*, then go
// read the layer packages to understand *how each piece works*.
package main

import (
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"os"

	"paperviz/internal/app/credentials"
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

	// No AI vendor credential is required here. PaperViz holds no model
	// account: every call runs on a key the user supplied, decrypted per
	// request by the resolver. GEMINI_API_KEY is only read by cmd/mcp, which
	// runs on the user's own machine.

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

	// CREDENTIAL_ENCRYPTION_KEY fails loud, unlike the peripheral integrations
	// that used to be required here. A wrong key does not disable a feature:
	// it makes every stored BYOK credential permanently unreadable while the
	// server still appears healthy, so it must never be allowed to boot.
	encKeyRaw := os.Getenv("CREDENTIAL_ENCRYPTION_KEY")
	if encKeyRaw == "" {
		slog.Error("CREDENTIAL_ENCRYPTION_KEY environment variable is required")
		os.Exit(1)
	}
	encKey, err := base64.StdEncoding.DecodeString(encKeyRaw)
	if err != nil {
		slog.Error("CREDENTIAL_ENCRYPTION_KEY must be base64-encoded", "error", err)
		os.Exit(1)
	}
	cipher, err := external.NewCipher(encKey)
	if err != nil {
		slog.Error("failed to build credential cipher", "error", err)
		os.Exit(1)
	}

	// The resolver is process-wide; the clients it produces are not. It holds
	// the shared transport and cipher and builds a per-user client from a
	// credential decrypted at request time.
	resolver := credentials.NewResolver(db, tr, cipher)

	// Start the expiry sweep in the background — runs once immediately,
	// then hourly (see services/expiry.go). It never returns, so it must
	// run in its own goroutine, not block main().
	go services.RunExpirySweepLoop(repository.NewDocumentRepo(db))

	router := handlers.NewRouter(db, resolver, staticDir)

	slog.Info("paperviz server starting", "port", port, "database_path", dbPath)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
