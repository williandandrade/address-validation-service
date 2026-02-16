package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	"github.com/williandandrade/address-validation-service/internal/infrastructure/address_parser"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

func newTestUsecase() *usecase.ValidateAddressUsecase {
	parser := address_parser.NewGopostalParser()
	return usecase.NewValidateAddressUsecase(parser)
}

func TestIntegration_ValidAddress(t *testing.T) {
	uc := newTestUsecase()

	tests := []struct {
		name           string
		input          string
		expectedCity   string
		expectedState  string
		expectedPostal string
	}{
		{
			name:           "comma-separated full address",
			input:          "123 Main St, New York, NY 10001",
			expectedCity:   "New York",
			expectedState:  "NY",
			expectedPostal: "10001",
		},
		{
			name:          "lowercase address normalized",
			input:         "456 oak ave, los angeles, ca 90210",
			expectedCity:  "Los Angeles",
			expectedState: "CA",
		},
		{
			name:          "address with extra spaces",
			input:         "789  Pine  Rd,  Chicago,  IL  60601",
			expectedState: "IL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: tt.input})

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.True(t, resp.Success)
			require.NotNil(t, resp.Address)

			if tt.expectedCity != "" {
				assert.Equal(t, tt.expectedCity, resp.Address.City)
			}
			if tt.expectedState != "" {
				assert.Equal(t, tt.expectedState, resp.Address.State)
			}
			if tt.expectedPostal != "" {
				assert.Equal(t, tt.expectedPostal, resp.Address.PostalCode)
			}

			assert.NotEmpty(t, resp.Address.FormattedAddress)
			assert.NotNil(t, resp.Confidence)
		})
	}
}

func TestIntegration_EmptyAddress(t *testing.T) {
	uc := newTestUsecase()

	resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: ""})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestIntegration_ResponseSchema(t *testing.T) {
	uc := newTestUsecase()

	resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{
		Address: "123 Main St, Springfield, IL 62701",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response structure matches spec
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Address)
	assert.NotEmpty(t, resp.Address.StreetAddress)
	assert.NotEmpty(t, resp.Address.State)
	assert.NotEmpty(t, resp.Address.AddressType)
	assert.NotEmpty(t, resp.Address.FormattedAddress)
	assert.NotNil(t, resp.Confidence)
	assert.NotEmpty(t, resp.Confidence.StateConfidence)
	assert.NotEmpty(t, resp.Message)
}

func TestIntegration_AddressFormats(t *testing.T) {
	uc := newTestUsecase()

	tests := []struct {
		name  string
		input string
	}{
		{"lowercase", "123 main st, new york, ny 10001"},
		{"uppercase", "123 MAIN ST, NEW YORK, NY 10001"},
		{"mixed case", "123 Main St, New York, NY 10001"},
		{"extra commas/spaces", "123 Main St , New York , NY 10001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{Address: tt.input})

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.True(t, resp.Success)
			assert.Equal(t, "NY", resp.Address.State)
		})
	}
}

func TestIntegration_POBox(t *testing.T) {
	uc := newTestUsecase()

	resp, err := uc.Execute(context.Background(), &dto.ValidateRequest{
		Address: "PO Box 123, Springfield, IL 62701",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "po_box", resp.Address.AddressType)
}
