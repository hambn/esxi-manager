package commands

import (
	"testing"
)

func TestCloneVMCommand_Validate_Success(t *testing.T) {
	cmd := &CloneVMCommand{
		SourceVMName:  "template-vm",
		DestVMName:    "new-vm",
		DestDiskStore: "datastore1",
		DestRAM:       2048,
		DestCPU:       2,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestCloneVMCommand_Validate_MissingSource(t *testing.T) {
	cmd := &CloneVMCommand{
		DestVMName:    "new-vm",
		DestDiskStore: "datastore1",
		DestRAM:       2048,
		DestCPU:       2,
	}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for missing source VM")
	}
}

func TestCloneVMCommand_Validate_MissingDest(t *testing.T) {
	cmd := &CloneVMCommand{
		SourceVMName:  "template-vm",
		DestDiskStore: "datastore1",
		DestRAM:       2048,
		DestCPU:       2,
	}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for missing dest VM name")
	}
}

func TestCloneVMCommand_Validate_InvalidRAM(t *testing.T) {
	cmd := &CloneVMCommand{
		SourceVMName:  "template-vm",
		DestVMName:    "new-vm",
		DestDiskStore: "datastore1",
		DestRAM:       0,
		DestCPU:       2,
	}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for invalid RAM")
	}
}

func TestCreateVMCommand_Validate_Success(t *testing.T) {
	cmd := &CreateVMCommand{
		VMName:    "new-vm",
		DiskStore: "datastore1",
		RAM:       2048,
		CPU:       2,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestCreateVMCommand_Validate_MissingName(t *testing.T) {
	cmd := &CreateVMCommand{
		DiskStore: "datastore1",
		RAM:       2048,
		CPU:       2,
	}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for missing VM name")
	}
}

func TestDeleteVMCommand_Validate_Success(t *testing.T) {
	cmd := &DeleteVMCommand{
		VMName: "vm-to-delete",
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestDeleteVMCommand_Validate_MissingName(t *testing.T) {
	cmd := &DeleteVMCommand{}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for missing VM name")
	}
}

func TestGetVMInfoCommand_Validate_Success(t *testing.T) {
	cmd := &GetVMInfoCommand{
		VMName: "my-vm",
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestListVMsCommand_Validate(t *testing.T) {
	cmd := &ListVMsCommand{}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestCreateVSwitchCommand_Validate_Success(t *testing.T) {
	cmd := &CreateVSwitchCommand{
		Name: "vswitch0",
		MTU:  1500,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestCreateVSwitchCommand_DefaultMTU(t *testing.T) {
	cmd := &CreateVSwitchCommand{
		Name: "vswitch0",
		MTU:  0,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if cmd.MTU != 1500 {
		t.Errorf("expected default MTU 1500, got %d", cmd.MTU)
	}
}

func TestCreatePortgroupCommand_Validate_Success(t *testing.T) {
	cmd := &CreatePortgroupCommand{
		Name:        "vm-network",
		VSwitchName: "vswitch0",
		VLAN:        100,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestCreatePortgroupCommand_Validate_InvalidVLAN(t *testing.T) {
	cmd := &CreatePortgroupCommand{
		Name:        "vm-network",
		VSwitchName: "vswitch0",
		VLAN:        5000,
	}

	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error for invalid VLAN")
	}
}

func TestListDatastoresCommand_Validate(t *testing.T) {
	cmd := &ListDatastoresCommand{}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}
