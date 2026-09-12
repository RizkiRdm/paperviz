package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	created  map[string]time.Time
	r        rate.Limit
	burst    int
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		created:  make(map[string]time.Time),
		r:        r,
		burst:    burst,
	}
}

func (l *ipRateLimiter) purgeOldEntries() {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, created := range l.created {
		if now.Sub(created) > 5*time.Minute {
			delete(l.limiters, ip)
			delete(l.created, ip)
		}
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.created[ip]; !ok || time.Since(l.created[ip]) > 5*time.Minute {
		lim := rate.NewLimiter(l.r, l.burst)
		l.limiters[ip] = lim
		l.created[ip] = time.Now()
		return lim
	}
	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.r, l.burst)
		l.limiters[ip] = lim
		l.created[ip] = time.Now()
	}
	return lim
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func rateLimitDocumentCreate(next http.Handler) http.Handler {
	limiter := newIPRateLimiter(rate.Every(30*1000_000_000), 2)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !limiter.getLimiter(ip).Allow() {
			writeError(w, http.StatusTooManyRequests, "rate_limited")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimitAuth limits auth requests to 5 per 60s per IP, burst 3.
func rateLimitAuth(next http.Handler) http.Handler {
	limiter := newIPRateLimiter(rate.Every(12_000_000_000), 3)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !limiter.getLimiter(ip).Allow() {
			writeError(w, http.StatusTooManyRequests, "rate_limited")
			return
		}
		next.ServeHTTP(w, r)
	})
}
