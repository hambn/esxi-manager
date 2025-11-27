package commands

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/esxi/network"
	"github.com/esxi-manager/esxi-manager/internal/esxi/storage"
	"github.com/esxi-manager/esxi-manager/internal/esxi/vm"
)

func TestDispatcher_CloneVM(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		SourceVMName:  "template",
		DestVMName:    "new-vm",
		DestDiskStore: "datastore1",
		DestRAM:       2048,
		DestCPU:       2,
	}

	cmd, err := d.Dispatch("clone-vm", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.CloneCommand); !ok {
		t.Fatalf("expected vm.CloneCommand, got %T", cmd)
	}
}

func TestDispatcher_CreateVM(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		DestVMName:    "new-vm",
		DestDiskStore: "datastore1",
		DestRAM:       2048,
		DestCPU:       2,
	}

	cmd, err := d.Dispatch("create-vm", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.CreateCommand); !ok {
		t.Fatalf("expected vm.CreateCommand, got %T", cmd)
	}
}

func TestDispatcher_DeleteVM(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VMName: "vm-to-delete",
	}

	cmd, err := d.Dispatch("delete-vm", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.DeleteCommand); !ok {
		t.Fatalf("expected vm.DeleteCommand, got %T", cmd)
	}
}

func TestDispatcher_ListVMs(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{}

	cmd, err := d.Dispatch("list-vms", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.ListCommand); !ok {
		t.Fatalf("expected vm.ListCommand, got %T", cmd)
	}
}

func TestDispatcher_GetVMInfo(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VMName: "my-vm",
	}

	cmd, err := d.Dispatch("get-vm-info", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.GetInfoCommand); !ok {
		t.Fatalf("expected vm.GetInfoCommand, got %T", cmd)
	}
}

func TestDispatcher_PowerOnVM(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VMName: "my-vm",
	}

	cmd, err := d.Dispatch("power-on-vm", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.PowerOnCommand); !ok {
		t.Fatalf("expected vm.PowerOnCommand, got %T", cmd)
	}
}

func TestDispatcher_PowerOffVM(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VMName: "my-vm",
	}

	cmd, err := d.Dispatch("power-off-vm", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*vm.PowerOffCommand); !ok {
		t.Fatalf("expected vm.PowerOffCommand, got %T", cmd)
	}
}

func TestDispatcher_CreateVSwitch(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VSwitchName: "vswitch0",
		MTU:         1500,
	}

	cmd, err := d.Dispatch("create-vswitch", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*network.CreateVSwitchCommand); !ok {
		t.Fatalf("expected network.CreateVSwitchCommand, got %T", cmd)
	}
}

func TestDispatcher_DeleteVSwitch(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		VSwitchName: "vswitch0",
	}

	cmd, err := d.Dispatch("delete-vswitch", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*network.DeleteVSwitchCommand); !ok {
		t.Fatalf("expected network.DeleteVSwitchCommand, got %T", cmd)
	}
}

func TestDispatcher_CreatePortgroup(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		PortgroupName: "vm-network",
		VSwitchName:   "vswitch0",
		VLAN:          100,
	}

	cmd, err := d.Dispatch("create-portgroup", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*network.CreatePortgroupCommand); !ok {
		t.Fatalf("expected network.CreatePortgroupCommand, got %T", cmd)
	}
}

func TestDispatcher_DeletePortgroup(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{
		PortgroupName: "vm-network",
	}

	cmd, err := d.Dispatch("delete-portgroup", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*network.DeletePortgroupCommand); !ok {
		t.Fatalf("expected network.DeletePortgroupCommand, got %T", cmd)
	}
}

func TestDispatcher_ListDatastores(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{}

	cmd, err := d.Dispatch("list-datastores", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*storage.ListDatastoresCommand); !ok {
		t.Fatalf("expected storage.ListDatastoresCommand, got %T", cmd)
	}
}

func TestDispatcher_UnknownCommand(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{}

	_, err := d.Dispatch("unknown-command", params)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
