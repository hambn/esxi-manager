package config

import (
	"flag"
	"sync"
)

// Params holds all application parameters - both connection and command parameters unified in one place
type Params struct {
	// ESXi Host Connection Parameters
	// These are required for all commands to establish SSH connectivity
	ESXiHostURI      string
	ESXiHostUsername string
	ESXiHostPassword string
	ESXiHostPort     int

	// Virtual Machine Parameters
	// Used by VM-related commands (clone, create, delete, list)
	VMName        string
	SourceVMName  string
	SourceVMID    string
	DestVMName    string
	DestDiskStore string
	DestRAM       int
	DestCPU       int
	DestNetwork   string

	// VM Inspect Parameters
	// Used by vm-inspect command to show all details about a specific VM
	VMInspectID   string
	VMInspectName string

	// Network Parameters
	// Used by networking commands (vswitch, portgroup management)
	VSwitchName   string
	PortgroupName string
	VLAN          int
	MTU           int
	Uplinks       []string

	// Storage Parameters
	// Used by storage commands (datastore operations)
	DatastoreName string
}

// CommandInterface defines the contract that all commands must implement
// Each command must be able to validate its specific parameters and execute its operation
type CommandInterface interface {
	Validate() error
	Execute() error
}

// CommandFactory is a factory function that creates a command instance
// It receives the unified Params and returns a CommandInterface ready to execute
type CommandFactory func(params *Params) CommandInterface

// CommandRegistry manages all registered commands with thread-safe access
// Commands auto-register themselves via init() functions in their packages
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]CommandFactory
}

// RegisterFlags registers all parameter flags with the flag package
// This centralizes all CLI flag definitions in one place
func (p *Params) RegisterFlags() {
	// ESXi Host Connection Flags
	flag.StringVar(&p.ESXiHostURI, "esxi-host-uri", "", "ESXi host URI/IP address")
	flag.StringVar(&p.ESXiHostUsername, "esxi-host-username", "", "ESXi host username")
	flag.StringVar(&p.ESXiHostPassword, "esxi-host-password", "", "ESXi host password")
	flag.IntVar(&p.ESXiHostPort, "esxi-host-port", 22, "ESXi host SSH port")

	// Virtual Machine Operation Flags
	flag.StringVar(&p.VMName, "vm-name", "", "Virtual machine name")
	flag.StringVar(&p.SourceVMName, "source-vm-name", "", "Source VM name for cloning")
	flag.StringVar(&p.SourceVMID, "source-vm-id", "", "Source VM ID for cloning")
	flag.StringVar(&p.DestVMName, "dest-vm-name", "", "Destination VM name")
	flag.StringVar(&p.DestDiskStore, "dest-vm-disk-store", "", "Destination datastore for VM")
	flag.IntVar(&p.DestRAM, "dest-vm-ram", 0, "Destination VM RAM in MB")
	flag.IntVar(&p.DestCPU, "dest-vm-cpu", 0, "Destination VM CPU count")
	flag.StringVar(&p.DestNetwork, "dest-vm-network", "", "Destination VM network/portgroup")

	// VM Inspect Flags
	flag.StringVar(&p.VMInspectID, "vm-inspect-id", "", "VM ID to inspect")
	flag.StringVar(&p.VMInspectName, "vm-inspect-name", "", "VM name to inspect")

	// Network Operation Flags
	flag.StringVar(&p.VSwitchName, "vswitch-name", "", "Virtual switch name")
	flag.StringVar(&p.PortgroupName, "portgroup-name", "", "Port group name")
	flag.IntVar(&p.VLAN, "vlan", 0, "VLAN ID")
	flag.IntVar(&p.MTU, "mtu", 1500, "MTU size")

	// Storage Operation Flags
	flag.StringVar(&p.DatastoreName, "datastore-name", "", "Datastore name")
}

