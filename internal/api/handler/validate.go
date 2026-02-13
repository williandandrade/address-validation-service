package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	domainErrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

// ValidateHandler handles single address validation requests.
type ValidateHandler struct {
	validateUsecase *usecase.ValidateUsecase
}

// NewValidateHandler creates a new ValidateHandler.
func NewValidateHandler(validateUsecase *usecase.ValidateUsecase) *ValidateHandler {
	return &ValidateHandler{
		validateUsecase: validateUsecase,
	}
}

// Validate handles the POST /api/v1/addresses/validate endpoint.
func (h *ValidateHandler) Validate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req dto.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Call usecase
	result, err := h.validateUsecase.Validate(r.Context(), req)
	if err != nil {
		h.handleUsecaseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		// Log error and return internal server error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ValidateHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(dto.APIErrorResponse{
		Code:    code,
		Message: message,
	})
	if err != nil {
		// Log error and return internal server error
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

func (h *ValidateHandler) handleUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domainErrors.ErrServiceUnavailable):
		w.Header().Set("Retry-After", "60")
		h.writeErrorWithRetry(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"Address validation service is temporarily unavailable", 60)
	case errors.Is(err, domainErrors.ErrRateLimitExceeded):
		w.Header().Set("Retry-After", "60")
		h.writeErrorWithRetry(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
			"Rate limit exceeded", 60)
	case errors.Is(err, domainErrors.ErrInvalidRequest):
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
	default:
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
	}
}

func (h *ValidateHandler) writeErrorWithRetry(w http.ResponseWriter, status int, code, message string, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(dto.APIErrorResponse{
		Code:              code,
		Message:           message,
		RetryAfterSeconds: retryAfter,
	})
	if err != nil {
		// Log error and return internal server error
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
