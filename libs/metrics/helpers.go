package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler is a wrapper for Prometheus
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordUserAction records a user action (like, skip, playlist_add, etc.).
func RecordUserAction(action string) {
	UserActionsTotal.WithLabelValues(action).Inc()
}

// RecordTrackStreamed increments the counter for tracks streamed.
func RecordTrackStreamed() {
	TracksStreamedTotal.Inc()
}

// RecordRecommendation records a recommendation served by the given algorithm.
func RecordRecommendation(algorithm string) {
	RecommendationsServedTotal.WithLabelValues(algorithm).Inc()
}

// TrackDownstreamCall tracks a downstream service call with metrics.
// Records both the call count (with success/error status) and duration.
//
// This is a convenience function to reduce boilerplate in service clients.
// It automatically determines the status based on whether an error occurred.
func TrackDownstreamCall(service, endpoint string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	DownstreamCallsTotal.WithLabelValues(service, endpoint, status).Inc()
	DownstreamCallDuration.WithLabelValues(service, endpoint).Observe(duration.Seconds())
}
