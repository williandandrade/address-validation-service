package handler

import (
	"gofr.dev/pkg/gofr"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

type ValidateAddressHandler struct {
	validateAddressUsecase usecase.ValidateAddressUsecaseInterface
}

func NewValidateAddressHandler(validateAddressUsecase usecase.ValidateAddressUsecaseInterface) *ValidateAddressHandler {
	return &ValidateAddressHandler{
		validateAddressUsecase: validateAddressUsecase,
	}
}

func (v *ValidateAddressHandler) Register(app *gofr.App) {
	app.POST("/api/v1/validate-address", func(ctx *gofr.Context) (any, error) {
		return v.Handle(ctx)
	})
}

func (v *ValidateAddressHandler) Handle(ctx *gofr.Context) (any, error) {
	request := new(dto.ValidateRequest)
	ctx.Request.Bind(request)

	return v.validateAddressUsecase.Execute(ctx, request)
}
