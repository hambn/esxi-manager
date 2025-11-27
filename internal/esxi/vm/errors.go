package vm

import "fmt"

// ValidationError represents a VM command validation error
type ValidationError struct {
	message string
}

// NewValidationError creates a new validation error
func NewValidationError(msg string) *ValidationError {
	return &ValidationError{message: msg}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("vm validation error: %s", e.message)
}

// ExecutionError represents a VM command execution error
type ExecutionError struct {
	message string
}

// NewExecutionError creates a new execution error
func NewExecutionError(msg string) *ExecutionError {
	return &ExecutionError{message: msg}
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("vm execution error: %s", e.message)
}
