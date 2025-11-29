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
