package router

import (
	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	// Import all command packages to trigger their init() functions
	// which register commands in the central registry
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/host"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/virtual-machines"
)

// Dispatcher routes commands by name using the centralized registry
type Dispatcher struct{}

// NewDispatcher creates a new command dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch creates a command from a command name and parameters
// Commands are looked up in the centralized registry
// (registered by init() functions in command packages)
// Each command receives the ESXi host config and manages its own connections
func (d *Dispatcher) Dispatch(commandName string, params command.Params, host *config.ESXiHost) (command.Interface, error) {
	return command.Get(commandName, params, host)
}
