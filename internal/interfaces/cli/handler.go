package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi"
)

// PrintUsage prints the CLI usage help text
func PrintUsage() {
	fmt.Fprintf(os.Stderr, `ESXi Manager CLI

USAGE:
  esxi-manager [options]

REQUIRED FLAGS:
  --esxi-host-uri         ESXi host URI/IP address
  --esxi-host-username    ESXi host username
  --esxi-host-password    ESXi host password
  --command              Command to execute

OPTIONAL FLAGS:
  --esxi-host-port       ESXi host SSH port (default: 22)

AVAILABLE COMMANDS:
  test-connection       Test SSH connectivity to ESXi host
  get-version          Get ESXi host version information
  list-vms             List all virtual machines on ESXi host
  list-port-groups     List all port groups on ESXi host
  list-datastores      List all datastores on ESXi host

EXAMPLE:
  esxi-manager \
    --esxi-host-uri="192.168.0.186" \
    --esxi-host-port="22" \
    --esxi-host-username="root" \
    --esxi-host-password="H@med1382" \
    --command="test-connection"

NOTE: All flags must come BEFORE any positional arguments.
`)
}

// ParseFlags parses command-line flags and returns parsed parameters
// Returns: (params, commandName, error)
func ParseFlags() (*config.Params, string, error) {
	params := &config.Params{}

	// ESXi Host Connection Flags
	flag.StringVar(&params.ESXiHostURI, "esxi-host-uri", "", "ESXi host URI/IP address")
	flag.StringVar(&params.ESXiHostUsername, "esxi-host-username", "", "ESXi host username")
	flag.StringVar(&params.ESXiHostPassword, "esxi-host-password", "", "ESXi host password")
	flag.IntVar(&params.ESXiHostPort, "esxi-host-port", 22, "ESXi host SSH port")

	// Command Selection
	commandName := flag.String("command", "", "Command to execute (test-connection, get-version, list-vms, etc)")

	// Virtual Machine Operation Flags
	flag.StringVar(&params.VMName, "vm-name", "", "Virtual machine name")
	flag.StringVar(&params.SourceVMName, "source-vm-name", "", "Source VM name for cloning")
	flag.StringVar(&params.SourceVMID, "source-vm-id", "", "Source VM ID for cloning")
	flag.StringVar(&params.DestVMName, "dest-vm-name", "", "Destination VM name")
	flag.StringVar(&params.DestDiskStore, "dest-vm-disk-store", "", "Destination datastore for VM")
	flag.IntVar(&params.DestRAM, "dest-vm-ram", 0, "Destination VM RAM in MB")
	flag.IntVar(&params.DestCPU, "dest-vm-cpu", 0, "Destination VM CPU count")
	flag.StringVar(&params.DestNetwork, "dest-vm-network", "", "Destination VM network/portgroup")

	// Network Operation Flags
	flag.StringVar(&params.VSwitchName, "vswitch-name", "", "Virtual switch name")
	flag.StringVar(&params.PortgroupName, "portgroup-name", "", "Port group name")
	flag.IntVar(&params.VLAN, "vlan", 0, "VLAN ID")
	flag.IntVar(&params.MTU, "mtu", 1500, "MTU size")

	// Storage Operation Flags
	flag.StringVar(&params.DatastoreName, "datastore-name", "", "Datastore name")

	flag.Parse()

	// Validate positional arguments (should be none)
	if flag.NArg() > 0 {
		args := flag.Args()
		return nil, "", fmt.Errorf("unexpected positional argument(s): %v\n\nHint: All flags (including --command) must come at the start, before any positional arguments.\nDid you put '%s' before your flags?", args, args[0])
	}

	// Validate all required parameters
	if err := validateParams(params, *commandName); err != nil {
		return nil, "", err
	}

	return params, *commandName, nil
}

// validateParams checks that all required parameters are provided
func validateParams(params *config.Params, commandName string) error {
	// Connection parameters are required for all commands
	if params.ESXiHostURI == "" {
		return fmt.Errorf("esxi-host-uri is required")
	}
	if params.ESXiHostUsername == "" {
		return fmt.Errorf("esxi-host-username is required")
	}
	if params.ESXiHostPassword == "" {
		return fmt.Errorf("esxi-host-password is required")
	}
	if params.ESXiHostPort <= 0 || params.ESXiHostPort > 65535 {
		return fmt.Errorf("invalid port: %d", params.ESXiHostPort)
	}

	// Command is required
	if commandName == "" {
		return fmt.Errorf("command is required")
	}

	return nil
}

// Executor handles CLI command execution
type Executor struct {
	params      *config.Params
	commandName string
}

// NewExecutor creates a new CLI executor
func NewExecutor(params *config.Params, commandName string) *Executor {
	return &Executor{params: params, commandName: commandName}
}

// Execute dispatches and executes the command
func (e *Executor) Execute() error {
	// Dispatch command and get command instance
	cmd, err := config.Dispatch(e.commandName, e.params)
	if err != nil {
		return common.WrapError(err, "failed to dispatch command")
	}

	// Validate command-specific parameters
	if err := cmd.Validate(); err != nil {
		return common.WrapError(err, "command validation failed")
	}

	// Execute the command (command manages its own connections)
	if err := cmd.Execute(); err != nil {
		return common.WrapError(err, "command execution failed")
	}

	return nil
}
