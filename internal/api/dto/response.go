package dto

import (
	"time"
)

// ValidateResponse represents the input address in a response.
type ValidateResponse struct {
	Street  string `json:"street"`
	Number  string `json:"number,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

// APIErrorResponse represents an API error response.
type APIErrorResponse struct {
	Code              string         `json:"code"`
	Message           string         `json:"message"`
	Details           map[string]any `json:"details,omitempty"`
	RetryAfterSeconds int            `json:"retry_after_seconds,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// ReadinessResponse represents the readiness check response.
type ReadinessResponse struct {
	Ready bool `json:"ready"`
}
