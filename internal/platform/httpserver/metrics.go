package httpserver

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Metrics collects per-route request counts, error counts, and latency
// buckets. It is intentionally dependency-free; scrape the JSON endpoint.
type Metrics struct {
	mu       sync.Mutex
	requests map[string]*routeStats
	started  time.Time
}

type routeStats struct {
	Count      int64   `json:"count"`
	Errors5xx  int64   `json:"errors_5xx"`
	Errors4xx  int64   `json:"errors_4xx"`
	TotalMs    float64 `json:"total_ms"`
	MaxMs      float64 `json:"max_ms"`
	BucketLEMs map[string]int64 `json:"latency_buckets_ms"`
}

var latencyBounds = []float64{5, 25, 100, 500, 2000}

// NewMetrics creates an empty registry.
func NewMetrics() *Metrics {
	return &Metrics{requests: make(map[string]*routeStats), started: time.Now().UTC()}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0
		route := r.Method + " " + routePattern(r)
		m.observe(route, recorder.status, elapsedMs)
	})
}

func routePattern(r *http.Request) string {
	if ctx := chi.RouteContext(r.Context()); ctx != nil && ctx.RoutePattern() != "" {
		return ctx.RoutePattern()
	}
	return r.URL.Path
}

func (m *Metrics) observe(route string, status int, elapsedMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats, ok := m.requests[route]
	if !ok {
		stats = &routeStats{BucketLEMs: make(map[string]int64)}
		m.requests[route] = stats
	}
	stats.Count++
	stats.TotalMs += elapsedMs
	if elapsedMs > stats.MaxMs {
		stats.MaxMs = elapsedMs
	}
	if status >= 500 {
		stats.Errors5xx++
	} else if status >= 400 {
		stats.Errors4xx++
	}
	for _, bound := range latencyBounds {
		if elapsedMs <= bound {
			key := msKey(bound)
			stats.BucketLEMs[key]++
		}
	}
	stats.BucketLEMs["+Inf"]++
}

func msKey(bound float64) string {
	return strconv.FormatFloat(bound, 'f', -1, 64)
}

// Snapshot renders the registry as JSON-ready data.
func (m *Metrics) Snapshot() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	routes := make([]string, 0, len(m.requests))
	for route := range m.requests {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	stats := make(map[string]any, len(routes))
	var total int64
	for _, route := range routes {
		entry := m.requests[route]
		total += entry.Count
		avg := 0.0
		if entry.Count > 0 {
			avg = entry.TotalMs / float64(entry.Count)
		}
		stats[route] = map[string]any{
			"count": entry.Count, "errors_4xx": entry.Errors4xx, "errors_5xx": entry.Errors5xx,
			"avg_ms": avg, "max_ms": entry.MaxMs, "latency_buckets_ms": entry.BucketLEMs,
		}
	}
	return map[string]any{
		"uptime_seconds": int64(time.Since(m.started).Seconds()),
		"total_requests": total,
		"routes":         stats,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
