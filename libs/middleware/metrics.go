package middleware

import (
	"libs/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// MetricsMiddleware returns middleware that automatically tracks HTTP request metrics.
// It measures request count and duration.
//
// This middleware tracks:
//   - http_requests_total - Total requests by method, endpoint, and status
//   - http_request_duration_seconds - Request latency
func MetricsMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Init
			start := time.Now()
			tracked := WrapResponseWriter(w)
			method := r.Method

			// Serve
			next.ServeHTTP(tracked, r)

			// Compute
			endpoint := r.URL.Path
			if route := mux.CurrentRoute(r); route != nil {
				if template, err := route.GetPathTemplate(); err == nil {
					logger.Warn("Route doesn't have a path template", zap.String("path", endpoint), zap.String("template", template), zap.String("method", method))
					endpoint = template
				}
			}
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(tracked.StatusCode)

			// Metrics
			metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		})
	}
}
