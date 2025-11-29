package esxi

// This file re-exports key types and functions from internal packages
// to provide a clean public API for the esxi package

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
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

// CommandRegistry manages command registration and execution
// Re-exported from config package for cleaner imports
type CommandRegistry = config.CommandRegistry

// Register registers a command factory with a name
// Re-exported from config package for cleaner imports
var Register = config.Register

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
