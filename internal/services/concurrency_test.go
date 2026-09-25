package services

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPipelineConcurrencyFromEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"unset uses default", "", defaultPipelineConcurrency},
		{"explicit value", "5", 5},
		{"one", "1", 1},
		{"garbage uses default", "lots", defaultPipelineConcurrency},
		{"zero uses default", "0", defaultPipelineConcurrency},
		{"negative uses default", "-3", defaultPipelineConcurrency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PIPELINE_MAX_CONCURRENCY", tt.value)
			if got := pipelineConcurrencyFromEnv(); got != tt.want {
				t.Fatalf("pipelineConcurrencyFromEnv() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIngestionEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"unset is enabled", "", true},
		{"true", "true", true},
		{"enabled spelled out", "1", true},
		{"false", "false", false},
		{"zero is false", "0", false},
		{"garbage is enabled", "maybe", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("INGESTION_ENABLED", tt.value)
			if got := IngestionEnabled(); got != tt.want {
				t.Fatalf("IngestionEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineGateBoundsConcurrency(t *testing.T) {
	g := newPipelineGate(2)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := g.acquire(ctx); err != nil {
			t.Fatalf("acquire %d: %v", i, err)
		}
	}

	// Gate is full: a third acquire must not succeed immediately.
	short, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := g.acquire(short); err == nil {
		t.Fatal("acquire succeeded while the gate was full")
	} else if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}

	g.release()
	if err := g.acquire(ctx); err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
}

// TestPipelineGateNeverExceedsLimit is the property that matters: no matter how
// many goroutines arrive, the number inside the critical section stays capped.
func TestPipelineGateNeverExceedsLimit(t *testing.T) {
	const limit = 3
	g := newPipelineGate(limit)

	var inflight, maxInflight atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.acquire(context.Background()); err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			cur := inflight.Add(1)
			for {
				prev := maxInflight.Load()
				if cur <= prev || maxInflight.CompareAndSwap(prev, cur) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			inflight.Add(-1)
			g.release()
		}()
	}
	wg.Wait()

	if got := maxInflight.Load(); got > limit {
		t.Fatalf("max concurrent = %d, want at most %d", got, limit)
	}
	if got := inflight.Load(); got != 0 {
		t.Fatalf("leaked %d slots", got)
	}
}

// TestPipelineGateReleaseWithoutAcquire proves a stray release cannot wedge the
// gate for every later caller.
func TestPipelineGateReleaseWithoutAcquire(t *testing.T) {
	g := newPipelineGate(1)

	g.release() // must not panic and must not free a slot it never took
	if err := g.acquire(context.Background()); err != nil {
		t.Fatalf("gate wedged by a stray release: %v", err)
	}
}

func TestPipelineGateRespectsCancelledContext(t *testing.T) {
	g := newPipelineGate(1)
	if err := g.acquire(context.Background()); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := g.acquire(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
