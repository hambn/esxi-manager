package config

import "sync"

// Params holds all parameters - connection and command parameters unified
type Params struct {
	// Connection parameters
	URI      string
	Username string
	Password string
	Port     int

	// Command parameters
	VMName        string
	SourceVMName  string
	SourceVMID    string
	DestVMName    string
	DestDiskStore string
	DestRAM       int
	DestCPU       int
	DestNetwork   string
	VSwitchName   string
	PortgroupName string
	VLAN          int
	MTU           int
	Uplinks       []string
	DatastoreName string
}

// CommandInterface defines the contract all commands must implement
type CommandInterface interface {
	Validate() error
	Execute() error
}

// CommandFactory creates a command instance
type CommandFactory func(params *Params) CommandInterface

// CommandRegistry manages registered commands
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]CommandFactory
}
