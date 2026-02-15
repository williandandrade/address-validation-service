//go:build !gopostal

package address_parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGopostalParser_ParseAddress(t *testing.T) {
	parser := NewGopostalParser()

	tests := []struct {
		name        string
		input       string
		expectErr   bool
		checkResult func(t *testing.T, result *struct{ street, city, state, postal, addrType string })
	}{
		{
			name:  "comma-separated full address",
			input: "123 Main St, New York, NY 10001",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "123 Main St", r.street)
				assert.Equal(t, "New York", r.city)
				assert.Equal(t, "NY", r.state)
				assert.Equal(t, "10001", r.postal)
				assert.Equal(t, "standard_street", r.addrType)
			},
		},
		{
			name:  "lowercase address normalized",
			input: "123 main st, new york, ny 10001",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "123 Main St", r.street)
				assert.Equal(t, "New York", r.city)
				assert.Equal(t, "NY", r.state)
				assert.Equal(t, "10001", r.postal)
			},
		},
		{
			name:  "address without zip code",
			input: "456 Oak Ave, Los Angeles, CA",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "456 Oak Ave", r.street)
				assert.Equal(t, "Los Angeles", r.city)
				assert.Equal(t, "CA", r.state)
				assert.Empty(t, r.postal)
			},
		},
		{
			name:  "PO Box detected as po_box type",
			input: "PO Box 123, Springfield, IL 62701",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "po_box", r.addrType)
				assert.Equal(t, "IL", r.state)
				assert.Equal(t, "62701", r.postal)
			},
		},
		{
			name:  "ZIP+4 format",
			input: "789 Pine Rd, Chicago, IL 60601-1234",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "60601-1234", r.postal)
				assert.Equal(t, "IL", r.state)
			},
		},
		{
			name:      "empty address returns error",
			input:     "",
			expectErr: true,
		},
		{
			name:  "space-separated address (no commas)",
			input: "123 Main St New York NY 10001",
			checkResult: func(t *testing.T, r *struct{ street, city, state, postal, addrType string }) {
				assert.Equal(t, "NY", r.state)
				assert.Equal(t, "10001", r.postal)
				assert.NotEmpty(t, r.street)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, _, err := parser.ParseAddress(context.Background(), tt.input)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, addr)

			if tt.checkResult != nil {
				result := &struct{ street, city, state, postal, addrType string }{
					street:   addr.StreetAddress,
					city:     addr.City,
					state:    addr.State,
					postal:   addr.PostalCode,
					addrType: addr.AddressType,
				}
				tt.checkResult(t, result)
			}
		})
	}
}

func TestDetectAddressType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123 Main St, New York, NY", "standard_street"},
		{"PO Box 456, Springfield, IL", "po_box"},
		{"P.O. Box 789, Chicago, IL", "po_box"},
		{"APO AE 09012", "apo_fpo"},
		{"FPO AP 96261", "apo_fpo"},
		{"Rural Route 1 Box 123, Nowhere, KS", "rural_route"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectAddressType(tt.input))
		})
	}
}

func TestTrackCorrections(t *testing.T) {
	t.Run("detects whitespace normalization", func(t *testing.T) {
		corrections := TrackCorrections("  123 Main St  ", map[string]string{"road": "Main St"})
		assert.Contains(t, corrections, "Normalized whitespace")
	})

	t.Run("detects capitalization correction", func(t *testing.T) {
		corrections := TrackCorrections("123 main st", map[string]string{"road": "main st"})
		assert.Contains(t, corrections, "Standardized capitalization")
	})

	t.Run("no corrections for clean input", func(t *testing.T) {
		corrections := TrackCorrections("123 Main St", map[string]string{"road": "Main St"})
		assert.Empty(t, corrections)
	})
}
