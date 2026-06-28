package middleware

import (
	"net/http"
	"sync"
	"time"
)

type limiter struct {
	tokens   int
	max      int
	lastFill time.Time
	mu       sync.Mutex
}

var limiters sync.Map

func getOrCreate(key string, max int) *limiter {
	val, _ := limiters.LoadOrStore(key, &limiter{tokens: max, max: max, lastFill: time.Now()})
	return val.(*limiter)
}

func RateLimit(subdomain string, rpm int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rpm == 0 {
				next.ServeHTTP(w, r)
				return
			}
			l := getOrCreate(subdomain, rpm)
			l.mu.Lock()
			elapsed := time.Since(l.lastFill)
			refill := int(elapsed.Minutes()) * l.max
			if refill > 0 {
				l.tokens = min(l.max, l.tokens+refill)
				l.lastFill = time.Now()
			}
			if l.tokens <= 0 {
				l.mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			l.tokens--
			l.mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
