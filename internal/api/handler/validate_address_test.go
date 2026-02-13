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
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

func TestValidateAddressHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		setupMocks       func(*usecase.MockValidateAddressUsecaseInterface)
		expectedResponse *dto.ValidateResponse
		expectedErr      bool
	}{
		{
			name:        "UseCase returns empty result",
			requestBody: `{"address":"123 Main St"}`,
			setupMocks: func(mockUsecase *usecase.MockValidateAddressUsecaseInterface) {
				mockUsecase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(&dto.ValidateResponse{}, nil).
					Times(1)
			},
			expectedResponse: &dto.ValidateResponse{},
			expectedErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock usecase
			mockUsecase := usecase.NewMockValidateAddressUsecaseInterface(ctrl)
			tt.setupMocks(mockUsecase)

			// Create handler with mocked usecase
			handler := NewValidateAddressHandler(mockUsecase)

			// Create HTTP request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/validate-address",
				bytes.NewBuffer([]byte(tt.requestBody)),
			)
			req.Header.Set("Content-Type", "application/json")

			// Create gofr context
			ctx := &gofr.Context{
				Context:   req.Context(),
				Request:   gofrHttp.NewRequest(req),
				Container: nil,
			}

			// Call handler
			result, err := handler.Handle(ctx)

			// Assert results
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, result)
			}
		})
	}
}
