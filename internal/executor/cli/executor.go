package cli

import (
	"github.com/esxi-manager/esxi-manager/internal/commands"
	"github.com/esxi-manager/esxi-manager/internal/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/connection"
)

// Executor handles CLI command execution
type Executor struct {
	params     *Params
	dispatcher *commands.Dispatcher
}

// NewExecutor creates a new CLI executor
func NewExecutor(params *Params) *Executor {
	return &Executor{
		params:     params,
		dispatcher: commands.NewDispatcher(),
	}
}

// Execute executes the CLI command
func (e *Executor) Execute() error {
	common.Info("executing command", "command", e.params.Command, "host", e.params.ESXiHostURI)

	// Create connection manager
	host := e.params.ToESXiHost()
	manager, err := connection.NewManager(host)
	if err != nil {
		return common.WrapError(err, "failed to create connection manager")
	}
	defer manager.Close()

	// Connect to host
	if err := manager.Connect(); err != nil {
		return common.WrapError(err, "failed to connect to ESXi host")
	}
	common.Info("connected to ESXi host", "host", e.params.ESXiHostURI)

	// Dispatch command
	cmdParams := e.params.ToCommandParams()
	cmd, err := e.dispatcher.Dispatch(e.params.Command, cmdParams)
	if err != nil {
		return common.WrapError(err, "failed to dispatch command")
	}

	// Validate command parameters
	if err := cmd.Validate(); err != nil {
		return common.WrapError(err, "command validation failed")
	}
	common.Info("command validated", "command", e.params.Command)

	// Get client and execute
	client, err := manager.GetClient()
	if err != nil {
		return common.WrapError(err, "failed to get SSH client")
	}

	if err := cmd.Execute(client); err != nil {
		return common.WrapError(err, "command execution failed")
	}

	common.Info("command executed successfully", "command", e.params.Command)
	return nil
}
