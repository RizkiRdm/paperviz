package mcp

import "sync"

// JobLimiter tracks concurrent in-flight jobs per API key.
type JobLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	maxJobs int
}

// NewJobLimiter creates a JobLimiter with the given per-key concurrency cap.
func NewJobLimiter(maxJobs int) *JobLimiter {
	return &JobLimiter{
		counts:  make(map[string]int),
		maxJobs: maxJobs,
	}
}

// Acquire attempts to reserve a job slot for the API key. Returns false if at capacity.
func (jl *JobLimiter) Acquire(apiKey string) bool {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	if jl.counts[apiKey] >= jl.maxJobs {
		return false
	}
	jl.counts[apiKey]++
	return true
}

// Release frees one job slot for the API key.
func (jl *JobLimiter) Release(apiKey string) {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	if jl.counts[apiKey] > 0 {
		jl.counts[apiKey]--
	}
}

// Current returns the number of in-flight jobs for the API key.
func (jl *JobLimiter) Current(apiKey string) int {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	return jl.counts[apiKey]
}

// Max returns the configured per-key concurrency limit.
func (jl *JobLimiter) Max() int {
	return jl.maxJobs
}
