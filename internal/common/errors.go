package common

import (
	"errors"
	"fmt"
)

// Define custom error types for common scenarios
var (
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrTimeout       = errors.New("operation timeout")
	ErrInternal      = errors.New("internal server error")
)

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted context
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// Is checks if an error is of a specific type (wrapper around errors.Is)
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in the chain matching a type (wrapper around errors.As)
func As(err error, target any) bool {
	return errors.As(err, target)
}
