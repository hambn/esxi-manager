package command

import (
	"fmt"
	"sync"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
)

// CommandFactory is a function that creates a command instance
// It receives both the command parameters and the ESXi host configuration
type CommandFactory func(params Params, host *config.ESXiHost) Interface

// Registry holds all registered commands
type Registry struct {
	mu       sync.RWMutex
	commands map[string]CommandFactory
}

var globalRegistry = &Registry{
	commands: make(map[string]CommandFactory),
}

// Register registers a command factory with a name
// This is called by command packages in their init() functions
func Register(name string, factory CommandFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.commands[name] = factory
}

// Get retrieves and instantiates a command by name
// Commands receive both parameters and ESXi host configuration
// Returns an error if the command is not registered
func Get(name string, params Params, host *config.ESXiHost) (Interface, error) {
	globalRegistry.mu.RLock()
	factory, ok := globalRegistry.commands[name]
	globalRegistry.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}

	return factory(params, host), nil
}
