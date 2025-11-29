package config

import "sync"

// ESXiHost represents the connection parameters for an ESXi host
type ESXiHost struct {
	URI      string
	Username string
	Password string
	Port     int
}

// Validate checks if the ESXi host configuration is valid
func (e *ESXiHost) Validate() error {
	if e.URI == "" {
		return ErrMissingURI
	}
	if e.Username == "" {
		return ErrMissingUsername
	}
	if e.Password == "" {
		return ErrMissingPassword
	}
	if e.Port == 0 {
		e.Port = 22 // Default SSH port
	}
	return nil
}

// CommandInterface defines the interface all commands must implement
type CommandInterface interface {
	Validate() error
	Execute() error
}

// CommandParams holds all possible command parameters
type CommandParams struct {
	// VM operations
	VMName          string
	SourceVMName    string
	SourceVMID      string
	DestVMName      string
	DestDiskStore   string
	DestRAM         int
	DestCPU         int
	DestNetwork     string

	// Network operations
	VSwitchName     string
	PortgroupName   string
	VLAN            int
	MTU             int
	Uplinks         []string

	// Storage operations
	DatastoreName   string
}

// CommandFactory is a function that creates a command instance
type CommandFactory func(params CommandParams, host *ESXiHost) CommandInterface

// CommandRegistry holds all registered commands
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]CommandFactory
}
