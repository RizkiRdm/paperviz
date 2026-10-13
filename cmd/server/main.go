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

	"github.com/clerk/clerk-sdk-go/v2"

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

	// Migrations are read from disk at startup, so the directory has to be
	// reachable. It defaults to "migrations" next to the binary's working
	// directory; the container image sets MIGRATIONS_DIR=/app/migrations.
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	// Chart images go to external object storage when configured, keeping them
	// out of the SQLite file. Unconfigured is supported and falls back to the
	// database column.
	images, err := external.ChartImageStoreFromEnv()
	if err != nil {
		slog.Error("chart image storage misconfigured; images will stay in the database", "error", err)
		images = nil
	}
	if images != nil {
		services.SetChartImageStore(images)
		slog.Info("chart images will be stored externally")
	}

	db, err := repository.Open(dbPath, migrationsDir)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	tr := external.NewTransport()

	// CLERK_SECRET_KEY fails loud for the same reason: a wrong or missing key
	// does not disable one feature, it makes every authenticated request fail
	// while healthz still reports ok. Booting in that state would look healthy
	// and lock every user out.
	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		slog.Error("CLERK_SECRET_KEY environment variable is required")
		os.Exit(1)
	}
	clerk.SetKey(clerkSecretKey)

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
