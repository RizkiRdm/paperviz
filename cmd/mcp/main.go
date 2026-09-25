// Command mcp is PaperViz's MCP server entrypoint. It exposes 5 tools
// (ingest_document, search_documents, get_document, get_figures, get_evidence)
// over stdio transport. No LLM calls in any tool path.
// Configuration comes from environment variables — same DB and Gemini key as
// the main server, so both can share one database.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/external"
	papervizMCP "paperviz/internal/mcp"
	"paperviz/internal/repository"
)

func main() {
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-2.5-flash-lite"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "paperviz.db"
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	papervizAPIKey := os.Getenv("PAPERVIZ_API_KEY")
	if papervizAPIKey == "" {
		log.Fatal("PAPERVIZ_API_KEY environment variable is required")
	}

	migrations, err := repository.LoadMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("failed to load migrations: %v", err)
	}

	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	tr := external.NewTransport()
	gemini, err := tr.For(external.ProviderGemini, geminiAPIKey, geminiModel)
	if err != nil {
		log.Fatalf("failed to build llm client: %v", err)
	}

	srv := papervizMCP.NewMCPServer(db, gemini, papervizAPIKey)

	if err := srv.Server().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
