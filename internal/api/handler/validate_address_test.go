package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gofr.dev/pkg/gofr"
	gofrHttp "gofr.dev/pkg/gofr/http"

	"github.com/williandandrade/address-validation-service/internal/api/dto"
	domainerrors "github.com/williandandrade/address-validation-service/internal/domain/errors"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

func TestValidateAddressHandler_Handle(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   string
		setupMocks    func(*usecase.MockValidateAddressUsecaseInterface)
		checkResponse func(t *testing.T, result any)
		expectedErr   bool
	}{
		{
			name:        "successful address validation",
			requestBody: `{"address":"123 Main St New York NY 10001"}`,
			setupMocks: func(mockUsecase *usecase.MockValidateAddressUsecaseInterface) {
				mockUsecase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(&dto.ValidateResponse{
						Success: true,
						Address: &dto.AddressDTO{
							StreetAddress:    "123 Main St",
							City:             "New York",
							State:            "NY",
							PostalCode:       "10001",
							AddressType:      "standard_street",
							FormattedAddress: "123 Main St, New York, NY 10001",
						},
						Message: "Address validated successfully",
					}, nil).
					Times(1)
			},
			checkResponse: func(t *testing.T, result any) {
				resp, ok := result.(*dto.ValidateResponse)
				require.True(t, ok)
				assert.True(t, resp.Success)
				assert.Equal(t, "123 Main St", resp.Address.StreetAddress)
				assert.Equal(t, "New York", resp.Address.City)
			},
			expectedErr: false,
		},
		{
			name:        "empty address returns validation error",
			requestBody: `{"address":""}`,
			setupMocks: func(mockUsecase *usecase.MockValidateAddressUsecaseInterface) {
				mockUsecase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, &domainerrors.ValidationError{
						Field:      "address",
						Reason:     "address field is required and cannot be empty",
						Suggestion: "Provide a valid US address",
					}).
					Times(1)
			},
			checkResponse: func(t *testing.T, result any) {
				resp, ok := result.(*dto.ValidateResponse)
				require.True(t, ok)
				assert.False(t, resp.Success)
				assert.Equal(t, "Request validation failed", resp.Message)
				require.Len(t, resp.Errors, 1)
				assert.Equal(t, "address", resp.Errors[0].Field)
			},
			expectedErr: false,
		},
		{
			name:        "unparseable address returns parsing error",
			requestBody: `{"address":"gibberish"}`,
			setupMocks: func(mockUsecase *usecase.MockValidateAddressUsecaseInterface) {
				mockUsecase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, &domainerrors.ParsingError{
						Field:      "address",
						Reason:     "Could not extract required address components",
						Suggestion: "Ensure address contains street address, city, and state",
					}).
					Times(1)
			},
			checkResponse: func(t *testing.T, result any) {
				resp, ok := result.(*dto.ValidateResponse)
				require.True(t, ok)
				assert.False(t, resp.Success)
				assert.Equal(t, "Address could not be normalized", resp.Message)
				require.Len(t, resp.Errors, 1)
				assert.Equal(t, "address", resp.Errors[0].Field)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := usecase.NewMockValidateAddressUsecaseInterface(ctrl)
			tt.setupMocks(mockUsecase)

			handler := NewValidateAddressHandler(mockUsecase)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/validate-address",
				bytes.NewBuffer([]byte(tt.requestBody)),
			)
			req.Header.Set("Content-Type", "application/json")

			ctx := &gofr.Context{
				Context:   req.Context(),
				Request:   gofrHttp.NewRequest(req),
				Container: nil,
			}

			result, err := handler.Handle(ctx)

			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, result)
				}
			}
		})
	}
}
