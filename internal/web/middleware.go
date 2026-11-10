package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type requestIDKey struct{}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed.",
		},
		[]string{"method", "path", "code"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [8]byte
	// crypto/rand.Read only fails when the OS CSPRNG is broken; ignore and still return hex (possibly weak).
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// withObserve logs each request and updates Prometheus counters after the handler finishes.
func withObserve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		elapsed := time.Since(start)
		path := r.URL.Path
		code := fmt.Sprintf("%d", rec.status)
		httpRequestsTotal.WithLabelValues(r.Method, path, code).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(elapsed.Seconds())

		attrs := []any{
			"method", r.Method,
			"path", path,
			"status", rec.status,
			"duration", elapsed,
		}
		if id, ok := r.Context().Value(requestIDKey{}).(string); ok {
			attrs = append(attrs, "request_id", id)
		}
		slog.Info("request", attrs...)
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				attrs := []any{"panic", v}
				if id, ok := r.Context().Value(requestIDKey{}).(string); ok {
					attrs = append(attrs, "request_id", id)
				}
				slog.Error("request panic", attrs...)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func chain(h http.Handler) http.Handler {
	return withRequestID(withObserve(withRecover(h)))
}
