package repository

import (
	"fmt"
	"path/filepath"
)

// migrationPaths maps version numbers to SQL filenames for all 20 migrations.
// Single source of truth — both cmd/server and cmd/mcp call LoadMigrations
// to prevent drift (Chunk 12.1).
var migrationPaths = map[int]string{
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
	17: "017_oauth_columns.sql",
	18: "018_api_key_column.sql",
	19: "019_billing_columns.sql",
	20: "020_user_credentials.sql",
}

// LoadMigrations reads every migration SQL file into a versioned map.
// Single source of truth for migration registration — prevents the DRY
// violation that caused migrations 18-19 to be missing on deploy.
func LoadMigrations(migrationsDir string) (map[int]string, error) {
	migrations := make(map[int]string)

	for version, file := range migrationPaths {
		sql, err := ReadMigration(filepath.Join(migrationsDir, file))
		if err != nil {
			return nil, fmt.Errorf("read migration %03d: %w", version, err)
		}
		migrations[version] = sql
	}
	return migrations, nil
}
