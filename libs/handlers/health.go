package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheckResponse represents the health check response structure
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// NewHealthCheckHandler returns an HTTP handler function for health checks
// with the specified service name
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
