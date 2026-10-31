package security

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"treeblog/internal/platform/httpserver"
)

// RateLimiter defines the interface for rate limiting checks.
type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (bool, time.Duration)
}

// MemoryRateLimiter implements an in-memory sliding window rate limiter.
type MemoryRateLimiter struct {
	mu      sync.Mutex
	buckets map[string][]time.Time
}

func NewMemoryRateLimiter() *MemoryRateLimiter {
	limiter := &MemoryRateLimiter{
		buckets: make(map[string][]time.Time),
	}
	go limiter.cleanupLoop(1 * time.Minute)
	return limiter
}

func (m *MemoryRateLimiter) Allow(key string, limit int, window time.Duration) (bool, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	timestamps := m.buckets[key]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		oldest := valid[0]
		retryAfter := oldest.Add(window).Sub(now)
		if retryAfter <= 0 {
			retryAfter = time.Second
		}
		m.buckets[key] = valid
		return false, retryAfter
	}

	m.buckets[key] = append(valid, now)
	return true, 0
}

func (m *MemoryRateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for key, times := range m.buckets {
			if len(times) == 0 || times[len(times)-1].Before(cutoff) {
				delete(m.buckets, key)
			}
		}
		m.mu.Unlock()
	}
}

// ClientIP extracts the real client IP from request headers or remote address.
func ClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(parts[0])
		if ip := net.ParseIP(clientIP); ip != nil {
			return clientIP
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// RateLimitMiddleware creates an HTTP middleware enforcing rate limits based on client IP.
func RateLimitMiddleware(limiter RateLimiter, limit int, window time.Duration, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r)
			key := fmt.Sprintf("%s:%s", keyPrefix, ip)

			allowed, retryAfter := limiter.Allow(key, limit, window)
			if !allowed {
				retrySeconds := int(retryAfter.Seconds())
				if retrySeconds <= 0 {
					retrySeconds = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retrySeconds))
				httpserver.WriteError(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded. Try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

