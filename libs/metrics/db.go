package metrics

import (
	"context"
	"database/sql"
	"time"
)

// TrackDBQuery wraps a database query function and tracks metrics.
// Use this to instrument individual database queries.
func TrackDBQuery(operation string, queryFn func() error) error {
	start := time.Now()

	err := queryFn()

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	DBQueriesTotal.WithLabelValues(operation, status).Inc()
	DBQueryDuration.WithLabelValues(operation).Observe(duration)

	return err
}

// TrackDBQueryContext wraps a database query function that requires context.
// Identical to TrackDBQuery but for functions that need context.
func TrackDBQueryContext(ctx context.Context, operation string, queryFn func(context.Context) error) error {
	start := time.Now()

	err := queryFn(ctx)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	DBQueriesTotal.WithLabelValues(operation, status).Inc()
	DBQueryDuration.WithLabelValues(operation).Observe(duration)

	return err
}

// UpdateDBPoolStats updates database connection pool metrics.
// Call this periodically (e.g., every 10 seconds) to track pool health.
func UpdateDBPoolStats(db *sql.DB) {
	stats := db.Stats()

	DBConnectionPoolSize.Set(float64(stats.OpenConnections))
	DBConnectionPoolInUse.Set(float64(stats.InUse))

	DBConnectionPoolWaitCount.Add(float64(stats.WaitCount))
	if stats.WaitDuration > 0 {
		DBConnectionPoolWaitDuration.Observe(stats.WaitDuration.Seconds())
	}
}
