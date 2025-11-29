package common

import (
	"errors"
	"fmt"
)

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// ValidationError represents a parameter validation failure
type ValidationError struct {
	Field   string // Which field failed validation
	Message string // Why it failed
	Value   string // Optional: the invalid value
}

func (e ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation error: %s (%s) - %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// NewValidationError creates a validation error with field context
func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}

// NewValidationErrorWithValue creates a validation error with the invalid value shown
func NewValidationErrorWithValue(field, message, value string) ValidationError {
	return ValidationError{Field: field, Message: message, Value: value}
}

// ConnectionError represents a connection failure
type ConnectionError struct {
	Host    string // Which host
	Reason  string // Why it failed
	Details error  // Underlying error
}

func (e ConnectionError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("connection error: cannot connect to %s - %s (%v)", e.Host, e.Reason, e.Details)
	}
	return fmt.Sprintf("connection error: cannot connect to %s - %s", e.Host, e.Reason)
}

func (e ConnectionError) Unwrap() error {
	return e.Details
}

// NewConnectionError creates a connection error with host context
func NewConnectionError(host, reason string, details error) ConnectionError {
	return ConnectionError{Host: host, Reason: reason, Details: details}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

// IsConnectionError checks if an error is a ConnectionError
func IsConnectionError(err error) bool {
	var ce ConnectionError
	return errors.As(err, &ce)
}
