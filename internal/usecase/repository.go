package usecase

import (
	"context"

	"github.com/williandandrade/address-validation-service/internal/domain/entity"
)

// ValidateAddressRepository defines the contract for address parsing.
type ValidateAddressRepository interface {
	ParseAddress(ctx context.Context, rawAddress string) (*entity.Address, []*entity.Address, error)
}
