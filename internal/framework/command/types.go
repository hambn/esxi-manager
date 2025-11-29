package command

// Interface defines the interface all commands must implement
// Specific command types are implemented in their respective packages:
// - internal/esxi/vm for VM operations
// - internal/esxi/network for network operations
// - internal/esxi/storage for storage operations
// - internal/esxi/general for general operations
//
// Commands are responsible for managing their own connections and execution.
// They receive ESXiHost configuration during registration and manage everything internally.
type Interface interface {
	Validate() error
	Execute() error
}

// Params holds all possible command parameters
// Executors populate only the fields relevant to their command
type Params struct {
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
