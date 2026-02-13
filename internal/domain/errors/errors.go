package errors

import "errors"

// Domain errors for address validation operations.
var (
	// ErrValidation indicates a validation error with the input data.
	ErrValidation = errors.New("validation error")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrServiceUnavailable indicates the external validation service is unavailable.
	ErrServiceUnavailable = errors.New("service unavailable")

	// ErrRateLimitExceeded indicates the rate limit has been exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrUnauthorized indicates invalid or missing authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidRequest indicates the request is malformed.
	ErrInvalidRequest = errors.New("invalid request")
)

// ValidationError represents a specific validation issue with details.
type ValidationError struct {
	Code    string
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewValidationError creates a new ValidationError with the given details.
func NewValidationError(code, field, message string) *ValidationError {
	return &ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
	}
}

// APIError represents an error returned by the API.
type APIError struct {
	Code              string
	Message           string
	Details           map[string]any
	RetryAfterSeconds int
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new APIError with the given code and message.
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the API error.
func (e *APIError) WithDetails(details map[string]any) *APIError {
	e.Details = details
	return e
}

// WithRetryAfter sets the retry-after seconds for the error.
func (e *APIError) WithRetryAfter(seconds int) *APIError {
	e.RetryAfterSeconds = seconds
	return e
}
