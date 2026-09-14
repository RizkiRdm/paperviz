package ratelimit

import (
	"context"
	"log/slog"
	"time"
)

// DefaultDelay is Gemini free-tier recovery pause between API calls.
const DefaultDelay = 3 * time.Second

// WaitRateLimit sleeps for the given delay or until ctx is done.
// Returns immediately on context cancellation — pipeline stays responsive.
func WaitRateLimit(ctx context.Context, delay time.Duration) {
	if delay <= 0 {
		return
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		slog.Warn("ratelimit wait interrupted by context", "error", ctx.Err())
	case <-timer.C:
	}
}
