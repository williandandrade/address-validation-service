package usecase

import (
	"context"
	"strings"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	"github.com/williandandrade/address-validation-service/internal/domain/entity"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
)

//go:generate mockgen -destination=validate_address_mock.go -package=usecase github.com/williandandrade/address-validation-service/internal/usecase ValidateAddressUsecaseInterface
type ValidateAddressUsecaseInterface interface {
	Execute(ctx context.Context, input *dto.ValidateRequest) (*dto.ValidateResponse, error)
}

// ValidateAddressUsecase handles address validation business logic.
type ValidateAddressUsecase struct {
	repo ValidateAddressRepository
}

// NewValidateAddressUsecase creates a new ValidateAddressUsecase.
func NewValidateAddressUsecase(repo ValidateAddressRepository) *ValidateAddressUsecase {
	return &ValidateAddressUsecase{repo: repo}
}

// Execute validates and normalizes a raw address string.
func (uc *ValidateAddressUsecase) Execute(ctx context.Context, input *dto.ValidateRequest) (*dto.ValidateResponse, error) {
	rawAddress := strings.TrimSpace(input.Address)
	if rawAddress == "" {
		return nil, &domainerrors.ValidationError{
			Field:      "address",
			Reason:     "address field is required and cannot be empty",
			Suggestion: "Provide a valid US address",
		}
	}

	addr, candidates, err := uc.repo.ParseAddress(ctx, rawAddress)
	if err != nil {
		return nil, err
	}

	if err := addr.Validate(); err != nil {
		return nil, &domainerrors.ParsingError{
			Field:      "address",
			Reason:     err.Error(),
			Suggestion: "Ensure address contains at least street address, city, and state",
		}
	}

	uc.assignConfidence(addr)
	addr.FormatAddress()

	for _, cand := range candidates {
		cand.FormatAddress()
	}

	resp := &dto.ValidateResponse{
		Success: true,
		Address: mapAddressToDTO(addr),
		Message: "Address validated successfully",
	}

	if addr.Confidence != nil {
		resp.Confidence = &dto.ConfidenceDTO{
			StateConfidence:  addr.Confidence.StateConfidence,
			CityConfidence:   addr.Confidence.CityConfidence,
			PostalConfidence: addr.Confidence.PostalConfidence,
		}
	}

	if len(addr.CorrectionsApplied) > 0 {
		resp.CorrectionsApplied = addr.CorrectionsApplied
	}

	if len(candidates) > 0 {
		for _, cand := range candidates {
			resp.Candidates = append(resp.Candidates, mapAddressToDTO(cand))
		}
		resp.Message = "Multiple valid interpretations found; returning most populous match"
	}

	return resp, nil
}

func (uc *ValidateAddressUsecase) assignConfidence(addr *entity.Address) {
	if addr.Confidence == nil {
		addr.Confidence = &entity.Confidence{}
	}

	if addr.State != "" {
		addr.Confidence.StateConfidence = "direct"
	}
	if addr.City != "" {
		addr.Confidence.CityConfidence = "direct"
	}
	if addr.PostalCode != "" {
		addr.Confidence.PostalConfidence = "direct"
	} else {
		addr.Confidence.PostalConfidence = "inferred"
	}
}

func mapAddressToDTO(addr *entity.Address) *dto.AddressDTO {
	return &dto.AddressDTO{
		StreetAddress:    addr.StreetAddress,
		City:             addr.City,
		State:            addr.State,
		PostalCode:       addr.PostalCode,
		AddressType:      addr.AddressType,
		FormattedAddress: addr.FormattedAddress,
	}
}
