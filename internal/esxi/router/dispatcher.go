package router

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/network"
	"github.com/esxi-manager/esxi-manager/internal/esxi/storage"
	"github.com/esxi-manager/esxi-manager/internal/esxi/vm"
)

// Dispatcher routes commands by name and creates the appropriate command type
type Dispatcher struct{}

// NewDispatcher creates a new command dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch creates a command from a command name and parameters
func (d *Dispatcher) Dispatch(commandName string, params command.Params) (command.Interface, error) {
	switch commandName {
	// VM operations
	case "clone-vm":
		cmd := &vm.CloneCommand{
			SourceVMName:  params.SourceVMName,
			SourceVMID:    params.SourceVMID,
			DestVMName:    params.DestVMName,
			DestDiskStore: params.DestDiskStore,
			DestRAM:       params.DestRAM,
			DestCPU:       params.DestCPU,
			DestNetwork:   params.DestNetwork,
		}
		return cmd, nil

	case "create-vm":
		cmd := &vm.CreateCommand{
			VMName:    params.DestVMName,
			DiskStore: params.DestDiskStore,
			RAM:       params.DestRAM,
			CPU:       params.DestCPU,
			Network:   params.DestNetwork,
			Datastore: params.DatastoreName,
		}
		return cmd, nil

	case "delete-vm":
		cmd := &vm.DeleteCommand{
			VMName: params.VMName,
		}
		return cmd, nil

	case "list-vms":
		return &vm.ListCommand{}, nil

	case "get-vm-info":
		cmd := &vm.GetInfoCommand{
			VMName: params.VMName,
		}
		return cmd, nil

	case "power-on-vm":
		cmd := &vm.PowerOnCommand{
			VMName: params.VMName,
		}
		return cmd, nil

	case "power-off-vm":
		cmd := &vm.PowerOffCommand{
			VMName: params.VMName,
		}
		return cmd, nil

	// Network operations
	case "create-vswitch":
		cmd := &network.CreateVSwitchCommand{
			Name:    params.VSwitchName,
			MTU:     params.MTU,
			Uplinks: params.Uplinks,
		}
		return cmd, nil

	case "delete-vswitch":
		cmd := &network.DeleteVSwitchCommand{
			Name: params.VSwitchName,
		}
		return cmd, nil

	case "create-portgroup":
		cmd := &network.CreatePortgroupCommand{
			Name:        params.PortgroupName,
			VSwitchName: params.VSwitchName,
			VLAN:        params.VLAN,
		}
		return cmd, nil

	case "delete-portgroup":
		cmd := &network.DeletePortgroupCommand{
			Name: params.PortgroupName,
		}
		return cmd, nil

	// Storage operations
	case "list-datastores":
		return &storage.ListDatastoresCommand{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
}
