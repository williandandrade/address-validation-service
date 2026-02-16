package usecase

import (
	"context"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
)

//go:generate mockgen -destination=validate_address_mock.go -package=usecase github.com/williandandrade/address-validation-service/internal/usecase ValidateAddressUsecaseInterface
type ValidateAddressUsecaseInterface interface {
	Execute(ctx context.Context, input *dto.ValidateRequest) (*dto.ValidateResponse, error)
}

// ValidateAddressUsecase handles single address validation logic.
type ValidateAddressUsecase struct {
}

// NewValidateAddressUsecase creates a new ValidateAddressUsecase.
func NewValidateAddressUsecase() *ValidateAddressUsecase {
	return &ValidateAddressUsecase{}
}

// Validate validates a single address.
func (v *ValidateAddressUsecase) Execute(ctx context.Context, input *dto.ValidateRequest) (*dto.ValidateResponse, error) {
	return &dto.ValidateResponse{}, nil
}
