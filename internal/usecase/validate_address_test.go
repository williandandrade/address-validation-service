package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
)

func TestValidateAddressUsecase_Execute(t *testing.T) {
	usecase := NewValidateAddressUsecase()

	t.Run("should return empty response and no error", func(t *testing.T) {
		ctx := context.Background()

		request := &dto.ValidateRequest{}
		response, err := usecase.Execute(ctx, request)

		assert.Empty(t, response)
		assert.NoError(t, err)
	})
}
