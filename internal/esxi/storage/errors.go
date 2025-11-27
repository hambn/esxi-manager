package storage

import "fmt"

// ValidationError represents a storage command validation error
type ValidationError struct {
	message string
}

// NewValidationError creates a new validation error
func NewValidationError(msg string) *ValidationError {
	return &ValidationError{message: msg}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("storage validation error: %s", e.message)
}

// ExecutionError represents a storage command execution error
type ExecutionError struct {
	message string
}

// NewExecutionError creates a new execution error
func NewExecutionError(msg string) *ExecutionError {
	return &ExecutionError{message: msg}
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("storage execution error: %s", e.message)
}
