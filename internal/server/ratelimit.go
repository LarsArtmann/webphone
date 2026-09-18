package server

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// keyedLimiter is a per-client token-bucket set with opportunistic
// pruning of idle buckets, so the map cannot grow unbounded under a
// spoofed-source flood.
type keyedLimiter struct {
	mu      sync.Mutex
	entries map[string]*keyedBucket
	limit   rate.Limit
	burst   int
	maxIdle time.Duration
	sweptAt time.Time
}

type keyedBucket struct {
	lim  *rate.Limiter
	seen time.Time
}

func newKeyedLimiter(limit rate.Limit, burst int) *keyedLimiter {
	return &keyedLimiter{
		entries: map[string]*keyedBucket{},
		limit:   limit,
		burst:   burst,
		maxIdle: 10 * time.Minute,
	}
}

// Allow consumes one token for key, creating its bucket on first use.
func (l *keyedLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.sweptAt) > time.Minute {
		for id, bucket := range l.entries {
			if now.Sub(bucket.seen) > l.maxIdle {
				delete(l.entries, id)
			}
		}
		l.sweptAt = now
	}
	bucket, ok := l.entries[key]
	if !ok {
		bucket = &keyedBucket{lim: rate.NewLimiter(l.limit, l.burst), seen: now}
		l.entries[key] = bucket
	}
	bucket.seen = now
	return bucket.lim.Allow()
}

// middleware answers 429 once a client's bucket is exhausted.
func (l *keyedLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(clientKey(r)) {
			w.Header().Set("Retry-After", "2")
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientKey identifies the direct peer. Behind the consuming stack's TLS
// terminator every request arrives from the proxy socket — the shared
// secret and the PBX-proven session remain the real boundaries; this
// only throttles flooding.
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
