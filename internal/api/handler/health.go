package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
)

// Version is the service version.
const Version = "1.0.0"

// HealthHandler handles health check endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles the health check endpoint.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := dto.HealthResponse{
		Status:    "healthy",
		Version:   Version,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		// Log error and return internal server error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Ready handles the readiness check endpoint.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	response := dto.ReadinessResponse{
		Ready: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		// Log error and return internal server error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
