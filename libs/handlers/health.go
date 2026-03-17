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

// MethodNotAllowedHandler returns an HTTP handler that responds with 405 Method Not Allowed.
// This handler is intended to be used with gorilla/mux's MethodNotAllowedHandler setting
// to provide consistent error responses when a client uses an unsupported HTTP method.
//
// Example usage:
//
//	router.MethodNotAllowedHandler = handlers.MethodNotAllowedHandler()
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
	})
}
