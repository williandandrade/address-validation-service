package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_Validate(t *testing.T) {
	tests := []struct {
		name      string
		address   Address
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid complete address",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
			},
			expectErr: false,
		},
		{
			name: "valid without postal code",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
			},
			expectErr: false,
		},
		{
			name: "valid with street and city only",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
			},
			expectErr: false,
		},
		{
			name: "invalid - only street",
			address: Address{
				StreetAddress: "123 Main St",
			},
			expectErr: true,
			errMsg:    "at least 2 of street_address, city, state must be present",
		},
		{
			name: "invalid state code",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "Faketown",
				State:         "XX",
			},
			expectErr: true,
			errMsg:    "invalid state code: XX",
		},
		{
			name: "invalid postal code format",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "1234",
			},
			expectErr: true,
			errMsg:    "postal_code must be 5 or 9-digit format",
		},
		{
			name: "valid ZIP+4 format",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001-1234",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAddress_FormatAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  Address
		expected string
	}{
		{
			name: "full address",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
			},
			expected: "123 Main St, New York, NY 10001",
		},
		{
			name: "without postal code",
			address: Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
			},
			expected: "123 Main St, New York, NY",
		},
		{
			name: "preserves existing formatted address",
			address: Address{
				StreetAddress:    "123 Main St",
				City:             "New York",
				State:            "NY",
				FormattedAddress: "Already Formatted",
			},
			expected: "Already Formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.address.FormatAddress()
			assert.Equal(t, tt.expected, result)
		})
	}
}
