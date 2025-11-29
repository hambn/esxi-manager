package cli

import (
	"github.com/esxi-manager/esxi-manager/internal/common"
	"github.com/esxi-manager/esxi-manager/internal/framework/router"
)

// Executor handles CLI command execution
// It just orchestrates: parses flags, dispatches to commands, and calls execute
// Commands manage their own connections and all internal details
type Executor struct {
	params     *Params
	dispatcher *router.Dispatcher
}

// NewExecutor creates a new CLI executor
func NewExecutor(params *Params) *Executor {
	return &Executor{
		params:     params,
		dispatcher: router.NewDispatcher(),
	}
}

// Execute executes the CLI command
// Executor is simple: just dispatch and execute the command
// The command manages all its own concerns (connections, parameters, execution)
func (e *Executor) Execute() error {
	// Convert CLI params to ESXiHost config
	host := e.params.ToESXiHost()
	cmdParams := e.params.ToCommandParams()

	// Dispatch command (returns configured command instance)
	cmd, err := e.dispatcher.Dispatch(e.params.Command, cmdParams, host)
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
