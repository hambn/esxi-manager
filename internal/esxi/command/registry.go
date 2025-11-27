package command

import (
	"fmt"
	"sync"
)

// CommandFactory is a function that creates a command instance
type CommandFactory func(params Params) Interface

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
// Returns an error if the command is not registered
func Get(name string, params Params) (Interface, error) {
	globalRegistry.mu.RLock()
	factory, ok := globalRegistry.commands[name]
	globalRegistry.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}

	return factory(params), nil
}
