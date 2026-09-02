// Command mcp is PaperViz's MCP server entrypoint. It exposes research
// operations (analyze, summarize, compare) as MCP tools over stdio transport.
// Configuration comes from environment variables — same DB and Gemini key as
// the main server, so both can share one database.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/external"
	papervizMCP "paperviz/internal/mcp"
	"paperviz/internal/repository"
)

// loadMigrations reads every migration SQL file into a versioned map.
// Replicated from cmd/server/main.go since loadMigrations is unexported.
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
		13: "013_usage_tiers.sql",
		14: "014_structured_research_objects.sql",
		15: "015_evidence_graph.sql",
		16: "016_annotations.sql",
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

	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("failed to load migrations: %v", err)
	}

	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	gemini := external.NewGeminiClient(geminiAPIKey, geminiModel)

	srv := papervizMCP.NewMCPServer(db, gemini, papervizAPIKey)

	if err := srv.Server().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
