package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
)

// PrintUsage prints the usage help text
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

// CommandParams is an alias for command.Params
type CommandParams = command.Params

// Params holds all CLI command parameters
type Params struct {
	// Connection parameters
	ESXiHostURI      string
	ESXiHostUsername string
	ESXiHostPassword string
	ESXiHostPort     int

	// Command type
	Command string

	// VM operations
	VMName           string
	SourceVMName     string
	SourceVMID       string
	DestVMName       string
	DestDiskStore    string
	DestRAM          int
	DestCPU          int
	DestNetwork      string

	// Network operations
	VSwitchName      string
	PortgroupName    string
	VLAN             int
	MTU              int

	// Storage operations
	DatastoreName string
}

// ParseFlags parses command-line flags and returns Params
func ParseFlags() (*Params, error) {
	params := &Params{}

	// Connection flags
	flag.StringVar(&params.ESXiHostURI, "esxi-host-uri", "", "ESXi host URI/IP address")
	flag.StringVar(&params.ESXiHostUsername, "esxi-host-username", "", "ESXi host username")
	flag.StringVar(&params.ESXiHostPassword, "esxi-host-password", "", "ESXi host password")
	flag.IntVar(&params.ESXiHostPort, "esxi-host-port", 22, "ESXi host SSH port")

	// Command selection
	flag.StringVar(&params.Command, "command", "", "Command to execute (clone-vm, create-vm, delete-vm, list-vms, etc)")

	// VM operation flags
	flag.StringVar(&params.VMName, "vm-name", "", "Virtual machine name")
	flag.StringVar(&params.SourceVMName, "source-vm-name", "", "Source VM name for cloning")
	flag.StringVar(&params.SourceVMID, "source-vm-id", "", "Source VM ID for cloning")
	flag.StringVar(&params.DestVMName, "dest-vm-name", "", "Destination VM name")
	flag.StringVar(&params.DestDiskStore, "dest-vm-disk-store", "", "Destination datastore for VM")
	flag.IntVar(&params.DestRAM, "dest-vm-ram", 0, "Destination VM RAM in MB")
	flag.IntVar(&params.DestCPU, "dest-vm-cpu", 0, "Destination VM CPU count")
	flag.StringVar(&params.DestNetwork, "dest-vm-network", "", "Destination VM network/portgroup")

	// Network operation flags
	flag.StringVar(&params.VSwitchName, "vswitch-name", "", "Virtual switch name")
	flag.StringVar(&params.PortgroupName, "portgroup-name", "", "Port group name")
	flag.IntVar(&params.VLAN, "vlan", 0, "VLAN ID")
	flag.IntVar(&params.MTU, "mtu", 1500, "MTU size")

	// Storage operation flags
	flag.StringVar(&params.DatastoreName, "datastore-name", "", "Datastore name")

	flag.Parse()

	// Check if there are positional arguments (means they were before flags)
	if flag.NArg() > 0 {
		args := flag.Args()
		return nil, fmt.Errorf("unexpected positional argument(s): %v\n\nHint: All flags (including --command) must come at the start, before any positional arguments.\nDid you put '%s' before your flags?", args, args[0])
	}

	if err := params.Validate(); err != nil {
		return nil, err
	}

	return params, nil
}

// Validate validates the parameters
func (p *Params) Validate() error {
	if p.ESXiHostURI == "" {
		return fmt.Errorf("esxi-host-uri is required")
	}
	if p.ESXiHostUsername == "" {
		return fmt.Errorf("esxi-host-username is required")
	}
	if p.ESXiHostPassword == "" {
		return fmt.Errorf("esxi-host-password is required")
	}
	if p.ESXiHostPort <= 0 || p.ESXiHostPort > 65535 {
		return fmt.Errorf("invalid port: %d", p.ESXiHostPort)
	}
	if p.Command == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}

// ToESXiHost converts params to config.ESXiHost
func (p *Params) ToESXiHost() *config.ESXiHost {
	return &config.ESXiHost{
		URI:      p.ESXiHostURI,
		Username: p.ESXiHostUsername,
		Password: p.ESXiHostPassword,
		Port:     p.ESXiHostPort,
	}
}

// ToCommandParams converts CLI params to command parameters
func (p *Params) ToCommandParams() CommandParams {
	return CommandParams{
		VMName:        p.VMName,
		SourceVMName:  p.SourceVMName,
		SourceVMID:    p.SourceVMID,
		DestVMName:    p.DestVMName,
		DestDiskStore: p.DestDiskStore,
		DestRAM:       p.DestRAM,
		DestCPU:       p.DestCPU,
		DestNetwork:   p.DestNetwork,
		VSwitchName:   p.VSwitchName,
		PortgroupName: p.PortgroupName,
		VLAN:          p.VLAN,
		MTU:           p.MTU,
		DatastoreName: p.DatastoreName,
	}
}
