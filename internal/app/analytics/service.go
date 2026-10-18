// Package analytics serves the aggregate usage metrics behind the dashboard.
package analytics

import (
	"database/sql"

	"paperviz/internal/repository"
)

// Summary computes all aggregate usage metrics in a single pass. It exists as
// a service boundary so HTTP handlers never construct repositories directly.
func Summary(db *sql.DB) (*repository.AnalyticsSummary, error) {
	return repository.NewAnalyticsRepo(db).GetSummary()
}
