package handler

import (
	"errors"

	"gofr.dev/pkg/gofr"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

// ValidateAddressHandler handles POST /api/v1/validate-address requests.
type ValidateAddressHandler struct {
	validateAddressUsecase usecase.ValidateAddressUsecaseInterface
}

// NewValidateAddressHandler creates a new ValidateAddressHandler.
func NewValidateAddressHandler(validateAddressUsecase usecase.ValidateAddressUsecaseInterface) *ValidateAddressHandler {
	return &ValidateAddressHandler{
		validateAddressUsecase: validateAddressUsecase,
	}
}

// Register registers the validate-address route with the GoFr app.
func (v *ValidateAddressHandler) Register(app *gofr.App) {
	app.POST("/api/v1/validate-address", func(ctx *gofr.Context) (any, error) {
		return v.Handle(ctx)
	})
}

// Handle processes the validate-address request.
func (v *ValidateAddressHandler) Handle(ctx *gofr.Context) (any, error) {
	request := new(dto.ValidateRequest)
	if err := ctx.Bind(request); err != nil {
		return &dto.ValidateResponse{
			Success: false,
			Errors: []dto.ErrorDTO{
				{
					Field:      "address",
					Reason:     "Invalid request format",
					Suggestion: "Provide a JSON body with an 'address' field",
				},
			},
			Message: "Request validation failed",
		}, nil
	}

	resp, err := v.validateAddressUsecase.Execute(ctx, request)
	if err != nil {
		return handleUsecaseError(err), nil
	}

	return resp, nil
}

func handleUsecaseError(err error) *dto.ValidateResponse {
	var validationErr *domainerrors.ValidationError
	if errors.As(err, &validationErr) {
		return &dto.ValidateResponse{
			Success: false,
			Errors: []dto.ErrorDTO{
				{
					Field:      validationErr.Field,
					Reason:     validationErr.Reason,
					Value:      validationErr.Value,
					Suggestion: validationErr.Suggestion,
				},
			},
			Message: "Request validation failed",
		}
	}

	var parsingErr *domainerrors.ParsingError
	if errors.As(err, &parsingErr) {
		return &dto.ValidateResponse{
			Success: false,
			Errors: []dto.ErrorDTO{
				{
					Field:      parsingErr.Field,
					Reason:     parsingErr.Reason,
					Value:      parsingErr.Value,
					Suggestion: parsingErr.Suggestion,
				},
			},
			Message: "Address could not be normalized",
		}
	}

	return &dto.ValidateResponse{
		Success: false,
		Message: "Internal server error",
	}
}
