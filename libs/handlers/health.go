package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheckResponse represents the health check response structure.
// It contains the service status and name.
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// NewHealthCheckHandler returns an HTTP handler function for health checks.
// The returned handler responds with a JSON object containing the service status ("healthy")
// and the provided service name. This is typically used for Kubernetes liveness/readiness probes.
//
// Example response:
//
//	{"status": "healthy", "service": "my-service"}
func NewHealthCheckHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthCheckResponse{
			Status:  "healthy",
			Service: serviceName,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
