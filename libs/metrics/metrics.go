package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP Metrics
var (
	// HTTPRequestsTotal counts all HTTP requests by method, endpoint, and status code.
	// Labels: method, endpoint, status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration measures HTTP request latency in seconds.
	// Labels: method, endpoint
	// Buckets optimized for web service latencies (5ms to 10s)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)
)

// Downstream Service Metrics
var (
	// DownstreamCallsTotal counts all calls to downstream services.
	// Labels: service (e.g., "user_database", "bandit"), endpoint, status
	// Track this manually in service client methods (e.g., UserDatabaseClient.GetUser)
	DownstreamCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "downstream_calls_total",
			Help: "Total number of calls to downstream services",
		},
		[]string{"service", "endpoint", "status"},
	)

	// DownstreamCallDuration measures latency of calls to downstream services.
	// Labels: service, endpoint
	// This includes network latency, downstream processing, and serialization overhead
	// Track this manually in service client methods (e.g., UserDatabaseClient.GetUser)
	DownstreamCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "downstream_call_duration_seconds",
			Help:    "Latency of calls to downstream services (full round-trip)",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "endpoint"},
	)
)

// Business Metrics
var (
	// TracksStreamedTotal counts total tracks streamed.
	TracksStreamedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tracks_streamed_total",
			Help: "Total number of tracks streamed",
		},
	)

	// RecommendationsServedTotal counts recommendations served.
	// Labels: algorithm (e.g., "bandit", "popularity")
	RecommendationsServedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recommendations_served_total",
			Help: "Total number of recommendations served",
		},
		[]string{"algorithm"},
	)

	// UserActionsTotal counts user actions.
	// Labels: action (e.g., "like", "skip", "playlist_add")
	UserActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_actions_total",
			Help: "Total number of user actions",
		},
		[]string{"action"},
	)
)
