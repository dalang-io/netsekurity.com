package main

import (
	"net/http"
	"sync"
	"time"
)

// A fixed-window rate limiter keyed by client IP and route. This is a single
// process (no replicas), so in-memory state is the whole truth.
//
// It exists because the write endpoints had no ceiling at all: one account could
// create Xendit invoices, domains, or scan jobs in a tight loop. The limits are
// generous for a human and tight for a script.
type rateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	buckets map[string]*rlBucket
}

type rlBucket struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{window: window, limit: limit, buckets: map[string]*rlBucket{}}
}

// allow reports whether key may proceed, counting this call.
func (rl *rateLimiter) allow(key string) bool {
	return rl.allowAt(key, time.Now())
}

func (rl *rateLimiter) allowAt(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b := rl.buckets[key]
	if b == nil || now.After(b.reset) {
		// Opportunistic sweep so the map cannot grow without bound.
		if len(rl.buckets) > 10000 {
			for k, v := range rl.buckets {
				if now.After(v.reset) {
					delete(rl.buckets, k)
				}
			}
		}
		rl.buckets[key] = &rlBucket{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.limit {
		return false
	}
	b.count++
	return true
}

// rateLimit wraps a handler. Only POST/PUT/DELETE requests count; a GET that
// renders a fragment must never be throttled by its own polling.
func rateLimit(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next(w, r)
			return
		}
		if !rl.allow(clientIP(r) + "|" + r.URL.Path) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "too many requests — try again in a minute", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// rateLimitAll is rateLimit for a read endpoint that scripts poll (the CI/CD API's
// assessment status). It counts GETs too, against one bucket per client address
// rather than per URL, so a loop over many ids cannot dodge the limit.
func rateLimitAll(rl *rateLimiter, bucket string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r) + "|" + bucket) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests; try again in a minute"}` + "\n"))
			return
		}
		next(w, r)
	}
}
