package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:  "address",
		Reason: "address field is required and cannot be empty",
	}
	assert.Equal(t, "validation error on field 'address': address field is required and cannot be empty", err.Error())
}

func TestParsingError_Error(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		err := &ParsingError{
			Field:  "state",
			Reason: "Unknown state code",
		}
		assert.Equal(t, "parsing error on field 'state': Unknown state code", err.Error())
	})

	t.Run("without field", func(t *testing.T) {
		err := &ParsingError{
			Reason: "Could not extract components",
		}
		assert.Equal(t, "parsing error: Could not extract components", err.Error())
	})
}

func TestAmbiguousAddressError_Error(t *testing.T) {
	err := &AmbiguousAddressError{
		Message: "Multiple valid interpretations found",
	}
	assert.Equal(t, "Multiple valid interpretations found", err.Error())
}
