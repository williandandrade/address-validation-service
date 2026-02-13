package usecase

import (
	"context"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
)

// ValidateUsecase handles single address validation logic.
type ValidateUsecase struct {
}

// NewValidateUsecase creates a new ValidateUsecase.
func NewValidateUsecase() *ValidateUsecase {
	return &ValidateUsecase{}
}

// Validate validates a single address.
func (u *ValidateUsecase) Validate(ctx context.Context, input dto.ValidateRequest) (*dto.ValidateResponse, error) {
	return &dto.ValidateResponse{}, nil
}
