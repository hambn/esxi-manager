package esxi

// This file re-exports key types and functions from internal packages
// to provide a clean public API for the esxi package

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// Params is the unified parameter type for all ESXi operations
// Re-exported from config package for cleaner imports
// Use Params.RegisterFlags() to register all CLI flags
type Params = config.Params

// CommandInterface defines the contract for a command
// Re-exported from config package for cleaner imports
type CommandInterface = config.CommandInterface

// CommandFactory is a function type that creates command instances
// Re-exported from config package for cleaner imports
type CommandFactory = config.CommandFactory

// SimpleCommand is a simple command function that takes params and returns (string, error)
// Re-exported from config package for cleaner imports
type SimpleCommand = config.SimpleCommand

// CommandRegistry manages command registration and execution
// Re-exported from config package for cleaner imports
type CommandRegistry = config.CommandRegistry

// Register registers a command factory with a name
// Re-exported from config package for cleaner imports
var Register = config.Register

// RegisterFunc registers a simple function as a command with auto-generated name
// Re-exported from config package for cleaner imports
// Command name is automatically derived from the file path
var RegisterFunc = config.RegisterFunc

// Error logs an error message with structured fields
// Re-exported from common package for cleaner imports
var Error = common.Error

// Warn logs a warning message with structured fields
// Re-exported from common package for cleaner imports
var Warn = common.Warn

// Info logs an info message with structured fields
// Re-exported from common package for cleaner imports
var Info = common.Info

// Debug logs a debug message with structured fields
// Re-exported from common package for cleaner imports
var Debug = common.Debug

// Logger returns the singleton logger instance
// Re-exported from common package for cleaner imports
var Logger = common.Logger

// WrapError wraps an error with additional context
// Re-exported from common package for cleaner imports
var WrapError = common.WrapError

// Dispatch finds and instantiates a command by name
// Re-exported from config package for cleaner imports
var Dispatch = config.Dispatch

// ValidateConnectionParams validates that all required connection parameters are set
// Re-exported from utils package for cleaner imports
var ValidateConnectionParams = utils.ValidateConnectionParams

// ValidationError represents a parameter validation failure
// Re-exported from common package for cleaner imports
type ValidationError = common.ValidationError

// ConnectionError represents a connection failure
// Re-exported from common package for cleaner imports
type ConnectionError = common.ConnectionError

// NewValidationError creates a validation error with field context
// Re-exported from common package for cleaner imports
var NewValidationError = common.NewValidationError

// NewValidationErrorWithValue creates a validation error with the invalid value shown
// Re-exported from common package for cleaner imports
var NewValidationErrorWithValue = common.NewValidationErrorWithValue

// NewConnectionError creates a connection error with host context
// Re-exported from common package for cleaner imports
var NewConnectionError = common.NewConnectionError

// IsValidationError checks if an error is a ValidationError
// Re-exported from common package for cleaner imports
var IsValidationError = common.IsValidationError

// IsConnectionError checks if an error is a ConnectionError
// Re-exported from common package for cleaner imports
var IsConnectionError = common.IsConnectionError
