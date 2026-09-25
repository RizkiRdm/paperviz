package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// defaultPipelineConcurrency is the number of documents processed at once when
// PIPELINE_MAX_CONCURRENCY is unset.
//
// The bound is about CPU, not money. Under BYOK the model bill belongs to the
// user, so nothing here limits spend, but PDF extraction, chart generation, and
// SQLite writes are all ours. README notes SQLite runs on a single connection,
// so unbounded concurrent pipelines degrade the whole service, not just the
// offending request.
const defaultPipelineConcurrency = 2

// pipelineGate bounds concurrent document processing.
//
// It is package-level because it is process-wide infrastructure, not
// request-scoped state: every upload funnels through it regardless of which
// handler or import path started the work. The precedent is the single-slot
// semaphore on external.Transport, which solves the same problem for model
// calls.
type pipelineGate struct {
	slots chan struct{}
}

// pipelineConcurrencyFromEnv reads PIPELINE_MAX_CONCURRENCY, falling back to the
// default on unset or unparseable input rather than refusing to start.
func pipelineConcurrencyFromEnv() int {
	raw := os.Getenv("PIPELINE_MAX_CONCURRENCY")
	if raw == "" {
		return defaultPipelineConcurrency
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		slog.Warn("ignoring invalid PIPELINE_MAX_CONCURRENCY, using default",
			"value", raw, "default", defaultPipelineConcurrency)
		return defaultPipelineConcurrency
	}
	return n
}

// newPipelineGate builds a gate allowing n concurrent pipelines.
func newPipelineGate(n int) *pipelineGate {
	return &pipelineGate{slots: make(chan struct{}, n)}
}

// acquire takes a slot, waiting until one frees or ctx expires. A queued
// pipeline that times out gives up rather than running unbounded.
func (g *pipelineGate) acquire(ctx context.Context) error {
	select {
	case g.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("waiting for a pipeline slot: %w", ctx.Err())
	}
}

// release returns a slot.
func (g *pipelineGate) release() {
	select {
	case <-g.slots:
	default:
		// Unreachable unless release is called without a matching acquire.
		// Swallowing it keeps a double-release from deadlocking the gate.
		slog.Error("pipeline gate released without a held slot")
	}
}

// pipelines is the process-wide gate. Sized once at startup.
var pipelines = newPipelineGate(pipelineConcurrencyFromEnv())

// IngestionEnabled reports whether document creation is allowed. It backs the
// INGESTION_ENABLED kill switch, which exists so ingestion can be stopped
// without a deploy: an abuse spike, a runaway queue, or an upstream incident.
func IngestionEnabled() bool {
	raw := os.Getenv("INGESTION_ENABLED")
	if raw == "" {
		return true
	}
	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("ignoring invalid INGESTION_ENABLED, treating as enabled", "value", raw)
		return true
	}
	return enabled
}
