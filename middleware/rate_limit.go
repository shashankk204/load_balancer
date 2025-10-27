package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	rate       int
	burst      int
	tokens     map[string]int
	lastRefill map[string]time.Time
	mu         sync.Mutex
}

func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		rate:      rate,
		burst:     burst,
		tokens:    make(map[string]int),
		lastRefill: make(map[string]time.Time),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	last := rl.lastRefill[key]
	elapsed := now.Sub(last).Seconds()

	// Refill tokens
	newTokens := int(elapsed * float64(rl.rate))
	if newTokens > 0 {
		rl.tokens[key] = min(rl.burst, rl.tokens[key]+newTokens)
		rl.lastRefill[key] = now
	}

	if rl.tokens[key] > 0 {
		rl.tokens[key]--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func RateLimitMiddleware(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !rl.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}