package config

import (
	"fmt"
)

var globalRegistry = &CommandRegistry{
	commands: make(map[string]CommandFactory),
}

// Register registers a command factory with a name
func Register(name string, factory CommandFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.commands[name] = factory
}

// SimpleCommand is a simple command function that takes params and returns (string, error)
type SimpleCommand func(params *Params) (string, error)

// RegisterFunc registers a simple function as a command with an explicit name
func RegisterFunc(name string, fn SimpleCommand) {
	Register(name, func(params *Params) CommandInterface {
		return &simpleCommandWrapper{fn: fn, params: params}
	})
}

// simpleCommandWrapper wraps a SimpleCommand function into CommandInterface
type simpleCommandWrapper struct {
	fn     SimpleCommand
	params *Params
}

// Execute implements CommandInterface by calling the wrapped function
func (s *simpleCommandWrapper) Execute() (string, error) {
	return s.fn(s.params)
}

// Dispatch retrieves and executes a command by name
func Dispatch(name string, params *Params) (CommandInterface, error) {
	globalRegistry.mu.RLock()
	factory, ok := globalRegistry.commands[name]
	globalRegistry.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}

	return factory(params), nil
}
