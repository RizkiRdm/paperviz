package services

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	TierFree     = "free"
	TierPro      = "pro"
	TierResearch = "research"

	LimitFree     = 5
	LimitPro      = 50
	LimitResearch = 500
)

// TierService wraps tier and usage logic for handler injection.
type TierService struct {
	db *sql.DB
}

// NewTierService creates a TierService with the given database connection.
func NewTierService(db *sql.DB) *TierService {
	return &TierService{db: db}
}

// GetTier returns the current tier for a fingerprint, defaulting to "free".
func (s *TierService) GetTier(fingerprint string) (string, error) {
	var tier string
	err := s.db.QueryRow(`SELECT tier FROM user_tiers WHERE fingerprint = ?`, fingerprint).Scan(&tier)
	if err == sql.ErrNoRows {
		return TierFree, nil
	}
	if err != nil {
		return "", fmt.Errorf("get tier: %w", err)
	}
	return tier, nil
}

// CheckUsage returns whether the user can create a paper and how many they've used.
func (s *TierService) CheckUsage(fingerprint string) (bool, int, error) {
	tier, err := s.GetTier(fingerprint)
	if err != nil {
		return false, 0, fmt.Errorf("check usage: %w", err)
	}

	limit := tierLimit(tier)

	var papersUsed int
	err = s.db.QueryRow(`SELECT papers_used FROM user_tiers WHERE fingerprint = ?`, fingerprint).Scan(&papersUsed)
	if err == sql.ErrNoRows {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("check usage: %w", err)
	}

	return papersUsed < limit, papersUsed, nil
}

// IncrementUsage bumps the papers_used count, creating the record if it doesn't exist.
func (s *TierService) IncrementUsage(fingerprint string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(`
		INSERT INTO user_tiers (fingerprint, papers_used, last_reset)
		VALUES (?, 1, ?)
		ON CONFLICT(fingerprint) DO UPDATE SET papers_used = papers_used + 1
	`, fingerprint, now)
	if err != nil {
		return fmt.Errorf("increment usage: %w", err)
	}
	return nil
}

// GetTier returns the current tier for a fingerprint (package-level, legacy).
func GetTier(db *sql.DB, fingerprint string) (string, error) {
	ts := &TierService{db: db}
	return ts.GetTier(fingerprint)
}

// CheckUsage returns whether the user can create a paper (package-level, legacy).
func CheckUsage(db *sql.DB, fingerprint string) (bool, int, error) {
	ts := &TierService{db: db}
	return ts.CheckUsage(fingerprint)
}

// IncrementUsage bumps the papers_used count (package-level, legacy).
func IncrementUsage(db *sql.DB, fingerprint string) error {
	ts := &TierService{db: db}
	return ts.IncrementUsage(fingerprint)
}

// ResetMonthlyUsage resets all counts when the month has changed since last_reset.
func ResetMonthlyUsage(db *sql.DB) error {
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	result, err := db.Exec(`
		UPDATE user_tiers
		SET papers_used = 0, last_reset = ?
		WHERE last_reset < ?
	`, now.Unix(), currentMonthStart)
	if err != nil {
		return fmt.Errorf("reset monthly usage: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reset monthly usage rows affected: %w", err)
	}
	if n > 0 {
		fmt.Printf("tier: reset monthly usage for %d users\n", n)
	}
	return nil
}

func tierLimit(tier string) int {
	switch tier {
	case TierPro:
		return LimitPro
	case TierResearch:
		return LimitResearch
	default:
		return LimitFree
	}
}
