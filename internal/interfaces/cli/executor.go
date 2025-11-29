package cli

import (
	"github.com/esxi-manager/esxi-manager/internal/common"
	"github.com/esxi-manager/esxi-manager/internal/config"
	// Import all command packages to trigger their init() functions
	// which register commands in the central registry
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/host"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/virtual-machines"
)

// Executor handles CLI command execution
type Executor struct {
	params      *config.Params
	commandName string
}

// NewExecutor creates a new CLI executor
func NewExecutor(params *config.Params, commandName string) *Executor {
	return &Executor{params: params, commandName: commandName}
}

// Execute executes the CLI command
func (e *Executor) Execute() error {
	cmd, err := config.Dispatch(e.commandName, e.params)
	if err != nil {
		return common.WrapError(err, "failed to dispatch command")
	}

	// Validate command parameters
	if err := cmd.Validate(); err != nil {
		return common.WrapError(err, "command validation failed")
	}

	// Execute command (command manages its own connections)
	if err := cmd.Execute(); err != nil {
		return common.WrapError(err, "command execution failed")
	}

	return nil
}
