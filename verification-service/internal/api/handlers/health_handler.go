package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles basic health check requests
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// DetailedHealthCheck handles detailed health check requests
func (h *HealthHandler) DetailedHealthCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check the health of all dependencies
	// For this example, we'll return a simplified response
	
	services := map[string]interface{}{
		"database": map[string]interface{}{
			"status":  "healthy",
			"latency": 12,
		},
		"storage": map[string]interface{}{
			"status":  "healthy",
			"latency": 48,
		},
		"bedrock": map[string]interface{}{
			"status":  "healthy",
			"latency": 125,
		},
		"lambda": map[string]interface{}{
			"status": "healthy",
		},
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
		"services":  services,
	})
}