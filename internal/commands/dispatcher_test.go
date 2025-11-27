package commands

import (
	"testing"
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

	if _, ok := cmd.(*CloneVMCommand); !ok {
		t.Fatalf("expected CloneVMCommand, got %T", cmd)
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

	if _, ok := cmd.(*CreateVMCommand); !ok {
		t.Fatalf("expected CreateVMCommand, got %T", cmd)
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

	if _, ok := cmd.(*DeleteVMCommand); !ok {
		t.Fatalf("expected DeleteVMCommand, got %T", cmd)
	}
}

func TestDispatcher_ListVMs(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{}

	cmd, err := d.Dispatch("list-vms", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*ListVMsCommand); !ok {
		t.Fatalf("expected ListVMsCommand, got %T", cmd)
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

	if _, ok := cmd.(*GetVMInfoCommand); !ok {
		t.Fatalf("expected GetVMInfoCommand, got %T", cmd)
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

	if _, ok := cmd.(*PowerOnVMCommand); !ok {
		t.Fatalf("expected PowerOnVMCommand, got %T", cmd)
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

	if _, ok := cmd.(*PowerOffVMCommand); !ok {
		t.Fatalf("expected PowerOffVMCommand, got %T", cmd)
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

	if _, ok := cmd.(*CreateVSwitchCommand); !ok {
		t.Fatalf("expected CreateVSwitchCommand, got %T", cmd)
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

	if _, ok := cmd.(*DeleteVSwitchCommand); !ok {
		t.Fatalf("expected DeleteVSwitchCommand, got %T", cmd)
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

	if _, ok := cmd.(*CreatePortgroupCommand); !ok {
		t.Fatalf("expected CreatePortgroupCommand, got %T", cmd)
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

	if _, ok := cmd.(*DeletePortgroupCommand); !ok {
		t.Fatalf("expected DeletePortgroupCommand, got %T", cmd)
	}
}

func TestDispatcher_ListDatastores(t *testing.T) {
	d := NewDispatcher()
	params := CommandParams{}

	cmd, err := d.Dispatch("list-datastores", params)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if _, ok := cmd.(*ListDatastoresCommand); !ok {
		t.Fatalf("expected ListDatastoresCommand, got %T", cmd)
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
