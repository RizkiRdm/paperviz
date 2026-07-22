package services

import (
	"log/slog"
	"time"
)

// expiryWindow is the 7-day inactivity limit from PRD.md ("stored result
// and link expire 7 days after last access") and ARCHITECTURE.md Acceptance
// Scenario 5.
const expiryWindow = 7 * 24 * time.Hour

// sweepInterval is how often the background sweep re-checks for expired
// documents after the initial startup sweep. A low-volume MVP doesn't need
// this to be tight — once an hour is frequent enough that no one waits
// meaningfully long past their 7-day window, without adding needless DB load.
const sweepInterval = 1 * time.Hour

// ExpiryDeleter is the one repository method the sweep needs. Declared as
// an interface here (rather than importing repository.DocumentRepo
// directly) so this file only depends on the shape it actually uses —
// easy to fake in a unit test without a real database.
type ExpiryDeleter interface {
	DeleteExpiredBefore(cutoff int64) (int64, error)
}

// RunExpirySweepLoop deletes documents past the 7-day inactivity window
// once immediately (covering time the server was offline) and then every
// sweepInterval thereafter. Intended to be started once in main.go via
// `go services.RunExpirySweepLoop(...)` — it never returns.
//
// This is the "Expiry Sweep — startup + interval" component from
// ARCHITECTURE.md's architecture diagram. It is deliberately a simple
// ticker loop, not a scheduled job framework — see AGENTS.md: no job queue
// infrastructure for MVP.
func RunExpirySweepLoop(repo ExpiryDeleter) {
	sweepOnce(repo)

	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for range ticker.C {
		sweepOnce(repo)
	}
}

// sweepOnce runs a single expiry pass. charts and claim_diffs rows cascade
// automatically via the ON DELETE CASCADE foreign keys defined in
// migrations/001_init.sql — this function only needs to delete from
// documents.
func sweepOnce(repo ExpiryDeleter) {
	cutoff := time.Now().Add(-expiryWindow).Unix()
	deleted, err := repo.DeleteExpiredBefore(cutoff)
	if err != nil {
		slog.Error("expiry sweep failed", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("expiry sweep", "stage", "expiry_sweep", "documents_deleted", deleted)
	}
}
