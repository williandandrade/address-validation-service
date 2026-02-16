package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/williandandrade/address-validation-service/internal/api/handler"
	"github.com/williandandrade/address-validation-service/internal/infrastructure/address_parser"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

func main() {
	app := gofr.New()

	// Infrastructure
	parser := address_parser.NewGopostalParser()

	// Usecases
	validateAddressUsecase := usecase.NewValidateAddressUsecase(parser)

	// Handlers
	validateAddressHandler := handler.NewValidateAddressHandler(validateAddressUsecase)
	validateAddressHandler.Register(app)

	app.Run()
}
