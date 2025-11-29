package config

import "fmt"

var globalRegistry = &CommandRegistry{
	commands: make(map[string]CommandFactory),
}

// Register registers a command factory with a name
// This is called by command packages in their init() functions
func Register(name string, factory CommandFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.commands[name] = factory
}

// GetCommand retrieves and instantiates a command by name
// Returns an error if the command is not registered
func GetCommand(name string, params CommandParams, host *ESXiHost) (CommandInterface, error) {
	globalRegistry.mu.RLock()
	factory, ok := globalRegistry.commands[name]
	globalRegistry.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}

	return factory(params, host), nil
}

// Dispatcher routes commands by name using the centralized registry
type Dispatcher struct{}

// NewDispatcher creates a new command dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch creates a command from a command name and parameters
// Commands are looked up in the centralized registry
// Each command receives the ESXi host config and manages its own connections
func (d *Dispatcher) Dispatch(commandName string, params CommandParams, host *ESXiHost) (CommandInterface, error) {
	return GetCommand(commandName, params, host)
}
