package errors

import "fmt"

// ValidationError represents a request validation failure (400 Bad Request).
type ValidationError struct {
	Field      string
	Reason     string
	Value      interface{}
	Suggestion string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Reason)
}

// ParsingError represents an address parsing failure (422 Unprocessable Entity).
type ParsingError struct {
	Field      string
	Reason     string
	Value      string
	Suggestion string
}

func (e *ParsingError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("parsing error on field '%s': %s", e.Field, e.Reason)
	}
	return fmt.Sprintf("parsing error: %s", e.Reason)
}

// AmbiguousAddressError represents multiple valid interpretations.
type AmbiguousAddressError struct {
	Message string
}

func (e *AmbiguousAddressError) Error() string {
	return e.Message
}
