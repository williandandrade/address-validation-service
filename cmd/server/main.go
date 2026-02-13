package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/williandandrade/address-validation-service/internal/api/handler"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

const (
	BaseURL = "/api/v1"
)

func main() {
	app := gofr.New()

	// Usecases
	validateAddressUsecase := usecase.NewValidateAddressUsecase()

	// Handlers
	validateAddressHandler := handler.NewValidateAddressHandler(validateAddressUsecase)
	validateAddressHandler.Register(app)

	app.Run()
}
