package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	"github.com/williandandrade/address-validation-service/internal/domain/entity"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
)

type mockRepo struct {
	parseFn func(ctx context.Context, raw string) (*entity.Address, []*entity.Address, error)
}

func (m *mockRepo) ParseAddress(ctx context.Context, raw string) (primary *entity.Address, candidates []*entity.Address, err error) {
	return m.parseFn(ctx, raw)
}

func TestValidateAddressUsecase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		input     *dto.ValidateRequest
		mockAddr  *entity.Address
		mockErr   error
		expectErr bool
		errType   string
		checkResp func(t *testing.T, resp *dto.ValidateResponse)
	}{
		{
			name:  "valid complete address",
			input: &dto.ValidateRequest{Address: "123 Main St New York NY 10001"},
			mockAddr: &entity.Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
				AddressType:   "standard_street",
			},
			expectErr: false,
			checkResp: func(t *testing.T, resp *dto.ValidateResponse) {
				assert.True(t, resp.Success)
				assert.Equal(t, "123 Main St", resp.Address.StreetAddress)
				assert.Equal(t, "New York", resp.Address.City)
				assert.Equal(t, "NY", resp.Address.State)
				assert.Equal(t, "10001", resp.Address.PostalCode)
				assert.NotNil(t, resp.Confidence)
				assert.Equal(t, "direct", resp.Confidence.StateConfidence)
			},
		},
		{
			name:      "empty address returns validation error",
			input:     &dto.ValidateRequest{Address: ""},
			expectErr: true,
			errType:   "ValidationError",
		},
		{
			name:      "whitespace-only address returns validation error",
			input:     &dto.ValidateRequest{Address: "   "},
			expectErr: true,
			errType:   "ValidationError",
		},
		{
			name:  "parser error propagated",
			input: &dto.ValidateRequest{Address: "gibberish"},
			mockErr: &domainerrors.ParsingError{
				Field:  "address",
				Reason: "Could not extract required address components",
			},
			expectErr: true,
			errType:   "ParsingError",
		},
		{
			name:  "address missing required components returns parsing error",
			input: &dto.ValidateRequest{Address: "somewhere"},
			mockAddr: &entity.Address{
				City: "Somewhere",
			},
			expectErr: true,
			errType:   "ParsingError",
		},
		{
			name:  "address without postal code gets inferred confidence",
			input: &dto.ValidateRequest{Address: "123 Main St New York NY"},
			mockAddr: &entity.Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				AddressType:   "standard_street",
			},
			expectErr: false,
			checkResp: func(t *testing.T, resp *dto.ValidateResponse) {
				assert.True(t, resp.Success)
				assert.Equal(t, "inferred", resp.Confidence.PostalConfidence)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				parseFn: func(_ context.Context, _ string) (*entity.Address, []*entity.Address, error) {
					if tt.mockErr != nil {
						return nil, nil, tt.mockErr
					}
					return tt.mockAddr, nil, nil
				},
			}

			uc := NewValidateAddressUsecase(repo)
			resp, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr {
				require.Error(t, err)
				switch tt.errType {
				case "ValidationError":
					var ve *domainerrors.ValidationError
					assert.ErrorAs(t, err, &ve)
				case "ParsingError":
					var pe *domainerrors.ParsingError
					assert.ErrorAs(t, err, &pe)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			}
		})
	}
}
