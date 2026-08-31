package mcp

import (
	"sync"

	"golang.org/x/time/rate"
)

// keyLimiters holds per-operation rate limiters for a single API key.
type keyLimiters struct {
	analyze *rate.Limiter
	read    *rate.Limiter
	compare *rate.Limiter
}

// RateLimiter manages per-key, per-operation rate limiting.
type RateLimiter struct {
	mu    sync.Mutex
	keys  map[string]*keyLimiters
}

// NewRateLimiter creates a RateLimiter with default limits.
// analyze: 5/min, read: 30/min, compare: 2/min.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		keys: make(map[string]*keyLimiters),
	}
}

// getOrCreate returns the limiter set for an API key, creating if absent.
func (rl *RateLimiter) getOrCreate(apiKey string) *keyLimiters {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	k, ok := rl.keys[apiKey]
	if !ok {
		k = &keyLimiters{
			analyze: rate.NewLimiter(rate.Every(60/5*60/1000*1000*1000000000/1000000000), 5), // ~5/min burst
			read:    rate.NewLimiter(rate.Every(2), 30),                                     // 30/min burst
			compare: rate.NewLimiter(rate.Every(30), 2),                                     // 2/min burst
		}
		k.analyze = rate.NewLimiter(rate.Limit(5.0/60.0), 5)
		k.read = rate.NewLimiter(rate.Limit(30.0/60.0), 30)
		k.compare = rate.NewLimiter(rate.Limit(2.0/60.0), 2)
		rl.keys[apiKey] = k
	}
	return k
}

// AllowAnalyze checks whether the API key can perform an analyze operation.
func (rl *RateLimiter) AllowAnalyze(apiKey string) bool {
	return rl.getOrCreate(apiKey).analyze.Allow()
}

// AllowRead checks whether the API key can perform a read operation.
func (rl *RateLimiter) AllowRead(apiKey string) bool {
	return rl.getOrCreate(apiKey).read.Allow()
}

// AllowCompare checks whether the API key can perform a compare operation.
func (rl *RateLimiter) AllowCompare(apiKey string) bool {
	return rl.getOrCreate(apiKey).compare.Allow()
}
