package repository

import (
	"database/sql"
	"fmt"
)

// AnalyticsSummary holds aggregate usage metrics for the dashboard.
type AnalyticsSummary struct {
	TotalPapers           int     `json:"total_papers"`
	TotalFigures          int     `json:"total_figures"`
	AvgProcessingTimeMs   int     `json:"avg_processing_time_ms"`
	ReturningUsers        int     `json:"returning_users"`
	PapersPerUser         float64 `json:"papers_per_user"`
	TotalShares           int     `json:"total_shares"`
	TotalShareVisits      int     `json:"total_share_visits"`
	TotalShareConversions int     `json:"total_share_conversions"`
	TotalComparisons      int     `json:"total_comparisons"`
	SuccessRate           float64 `json:"success_rate"`
}

// AnalyticsRepo reads aggregate usage metrics from the database.
type AnalyticsRepo struct {
	db *sql.DB
}

// NewAnalyticsRepo constructs an AnalyticsRepo backed by the given database.
func NewAnalyticsRepo(db *sql.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

// GetSummary computes all aggregate usage metrics in a single pass.
func (r *AnalyticsRepo) GetSummary() (*AnalyticsSummary, error) {
	s := &AnalyticsSummary{}

	// Total number of processed papers.
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&s.TotalPapers); err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	// Total number of re-visualized figures.
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM charts`).Scan(&s.TotalFigures); err != nil {
		return nil, fmt.Errorf("count charts: %w", err)
	}

	// Average pipeline processing time in milliseconds (0 when none recorded).
	var avgMs float64
	if err := r.db.QueryRow(`SELECT COALESCE(AVG(processing_time_ms), 0) FROM documents WHERE processing_time_ms IS NOT NULL`).Scan(&avgMs); err != nil {
		return nil, fmt.Errorf("avg processing time: %w", err)
	}
	s.AvgProcessingTimeMs = int(avgMs)

	// Distinct users that have submitted at least one paper.
	if err := r.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM documents WHERE user_id IS NOT NULL`).Scan(&s.ReturningUsers); err != nil {
		return nil, fmt.Errorf("count returning users: %w", err)
	}

	// Papers per user, guarded against divide-by-zero.
	if s.ReturningUsers > 0 {
		s.PapersPerUser = float64(s.TotalPapers) / float64(s.ReturningUsers)
	}

	// Total share visits across documents and charts.
	if err := r.db.QueryRow(`SELECT COALESCE((SELECT SUM(share_visits) FROM documents), 0) + COALESCE((SELECT SUM(share_visits) FROM charts), 0)`).Scan(&s.TotalShareVisits); err != nil {
		return nil, fmt.Errorf("sum share visits: %w", err)
	}

	// Total share conversions across documents and charts.
	if err := r.db.QueryRow(`SELECT COALESCE((SELECT SUM(share_conversions) FROM documents), 0) + COALESCE((SELECT SUM(share_conversions) FROM charts), 0)`).Scan(&s.TotalShareConversions); err != nil {
		return nil, fmt.Errorf("sum share conversions: %w", err)
	}

	// Shares tracked as visits for now.
	s.TotalShares = s.TotalShareVisits

	// Total comparison events recorded.
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM analytics_events WHERE event_type = 'comparison'`).Scan(&s.TotalComparisons); err != nil {
		return nil, fmt.Errorf("count comparisons: %w", err)
	}

	// Pipeline success rate (completed documents over all documents).
	var completed, total int
	if err := r.db.QueryRow(`SELECT COUNT(CASE WHEN status = 'complete' THEN 1 END), COUNT(*) FROM documents`).Scan(&completed, &total); err != nil {
		return nil, fmt.Errorf("compute success rate: %w", err)
	}
	if total > 0 {
		s.SuccessRate = float64(completed) / float64(total)
	}

	return s, nil
}
