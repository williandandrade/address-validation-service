package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
)

func TestIntegration_EmptyAddressError(t *testing.T) {
	uc := newTestUsecase()

	_, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: ""})

	require.Error(t, err)
	var ve *domainerrors.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "address", ve.Field)
	assert.Contains(t, ve.Reason, "required")
	assert.NotEmpty(t, ve.Suggestion)
}

func TestIntegration_WhitespaceOnlyAddress(t *testing.T) {
	uc := newTestUsecase()

	_, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: "   "})

	require.Error(t, err)
	var ve *domainerrors.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "address", ve.Field)
}

func TestIntegration_ErrorResponseConsistency(t *testing.T) {
	uc := newTestUsecase()

	// Same invalid input should always produce the same error
	for i := 0; i < 3; i++ {
		_, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: ""})
		require.Error(t, err)

		var ve *domainerrors.ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "address", ve.Field)
		assert.Equal(t, "address field is required and cannot be empty", ve.Reason)
	}
}

func TestIntegration_CorrectionsTracking(t *testing.T) {
	uc := newTestUsecase()

	resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{
		Address: "123 main st, new york, ny 10001",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	// Lowercase input should trigger capitalization correction
	assert.NotEmpty(t, resp.CorrectionsApplied)
}
