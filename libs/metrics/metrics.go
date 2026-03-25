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

// Database Metrics
var (
	// DBQueriesTotal counts all database queries by operation and status.
	// Labels: operation (e.g., "GetUser", "CreateAlbum"), status (success/error)
	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// DBQueryDuration measures database query latency in seconds (from app perspective).
	// Labels: operation
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency in seconds (from application perspective)",
			Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	// DBConnectionPoolSize tracks the current size of the DB connection pool.
	DBConnectionPoolSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_size",
			Help: "Current number of connections in the database pool",
		},
	)

	// DBConnectionPoolInUse tracks how many connections are currently in use.
	DBConnectionPoolInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_in_use",
			Help: "Number of database connections currently in use",
		},
	)

	// DBConnectionPoolWaitCount counts how many times we waited for a connection.
	DBConnectionPoolWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_connection_pool_wait_count_total",
			Help: "Total number of times waited for a database connection",
		},
	)

	// DBConnectionPoolWaitDuration measures time spent waiting for a connection.
	DBConnectionPoolWaitDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_connection_pool_wait_duration_seconds",
			Help:    "Time spent waiting for a database connection",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
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
