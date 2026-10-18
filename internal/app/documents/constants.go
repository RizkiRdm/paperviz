package documents

import (
	"paperviz/internal/models"
	"paperviz/internal/repository"
)

// ErrNotFound is the document-missing sentinel handlers map to 404. It is the
// repository's error value re-exported here so the transport layer never
// imports the repository to compare against it.
var ErrNotFound = repository.ErrNotFound

// Domain constants re-exported from the canonical models package so transport
// and service layers validate against the same values without importing the
// repository.
const (
	ReadingLevelSimplified = models.ReadingLevelSimplified
	ReadingLevelELI5       = models.ReadingLevelELI5
	StatusProcessing       = models.StatusProcessing
)

// ValidReadingLevel reports whether level is one of the two supported
// simplification depths accepted at intake.
func ValidReadingLevel(level string) bool {
	return level == ReadingLevelSimplified || level == ReadingLevelELI5
}
