package dto

import (
	"time"
)

// ValidateResponse represents the API response for address validation.
type ValidateResponse struct {
	Success            bool           `json:"success"`
	Address            *AddressDTO    `json:"address,omitempty"`
	Candidates         []*AddressDTO  `json:"candidates,omitempty"`
	Confidence         *ConfidenceDTO `json:"confidence,omitempty"`
	CorrectionsApplied []string       `json:"corrections_applied,omitempty"`
	Errors             []ErrorDTO     `json:"errors,omitempty"`
	Message            string         `json:"message"`
}

// AddressDTO represents a normalized address in the response.
type AddressDTO struct {
	StreetAddress    string `json:"street_address"`
	City             string `json:"city"`
	State            string `json:"state"`
	PostalCode       string `json:"postal_code"`
	AddressType      string `json:"address_type"`
	FormattedAddress string `json:"formatted_address,omitempty"`
}

// ConfidenceDTO represents confidence levels for address components.
type ConfidenceDTO struct {
	StateConfidence  string `json:"state_confidence"`
	CityConfidence   string `json:"city_confidence"`
	PostalConfidence string `json:"postal_confidence"`
}

// ErrorDTO represents a field-level error in the response.
type ErrorDTO struct {
	Field      string `json:"field"`
	Reason     string `json:"reason"`
	Value      any    `json:"value,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
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
