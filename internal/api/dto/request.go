package dto

// ValidateRequest represents the request body for single address validation.
type ValidateRequest struct {
	Address string `json:"address"`
}
