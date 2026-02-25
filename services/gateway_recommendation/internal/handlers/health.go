package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthCheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthCheckResponse{
		Status:  "healthy",
		Service: "gateway-recommendation",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
