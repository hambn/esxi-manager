package vm

import (
	"golang.org/x/crypto/ssh"
)

// Command defines the interface VM commands must implement
type Command interface {
	Validate() error
	Execute(client *ssh.Client) error
}

// CloneCommand parameters for cloning a virtual machine
type CloneCommand struct {
	SourceVMName   string
	SourceVMID     string
	DestVMName     string
	DestDiskStore  string
	DestRAM        int // in MB
	DestCPU        int
	DestNetwork    string
}

func (c *CloneCommand) Validate() error {
	if c.SourceVMName == "" && c.SourceVMID == "" {
		return NewValidationError("source-vm-name or source-vm-id is required")
	}
	if c.DestVMName == "" {
		return NewValidationError("dest-vm-name is required")
	}
	if c.DestDiskStore == "" {
		return NewValidationError("dest-vm-disk-store is required")
	}
	if c.DestRAM <= 0 {
		return NewValidationError("dest-vm-ram must be greater than 0")
	}
	if c.DestCPU <= 0 {
		return NewValidationError("dest-vm-cpu must be greater than 0")
	}
	return nil
}

func (c *CloneCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM clone via SSH
	return nil
}

// CreateCommand parameters for creating a new virtual machine
type CreateCommand struct {
	VMName      string
	DiskStore   string
	RAM         int // in MB
	CPU         int
	Network     string
	Datastore   string
}

func (c *CreateCommand) Validate() error {
	if c.VMName == "" {
		return NewValidationError("vm-name is required")
	}
	if c.DiskStore == "" {
		return NewValidationError("disk-store is required")
	}
	if c.RAM <= 0 {
		return NewValidationError("ram must be greater than 0")
	}
	if c.CPU <= 0 {
		return NewValidationError("cpu must be greater than 0")
	}
	return nil
}

func (c *CreateCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM creation via SSH
	return nil
}

// DeleteCommand parameters for deleting a virtual machine
type DeleteCommand struct {
	VMName string
}

func (c *DeleteCommand) Validate() error {
	if c.VMName == "" {
		return NewValidationError("vm-name is required")
	}
	return nil
}

func (c *DeleteCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM deletion via SSH
	return nil
}

// ListCommand lists all virtual machines
type ListCommand struct{}

func (c *ListCommand) Validate() error {
	return nil
}

func (c *ListCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM listing via SSH
	return nil
}

// GetInfoCommand gets information about a specific VM
type GetInfoCommand struct {
	VMName string
}

func (c *GetInfoCommand) Validate() error {
	if c.VMName == "" {
		return NewValidationError("vm-name is required")
	}
	return nil
}

func (c *GetInfoCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM info retrieval via SSH
	return nil
}

// PowerOnCommand powers on a virtual machine
type PowerOnCommand struct {
	VMName string
}

func (c *PowerOnCommand) Validate() error {
	if c.VMName == "" {
		return NewValidationError("vm-name is required")
	}
	return nil
}

func (c *PowerOnCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM power on via SSH
	return nil
}

// PowerOffCommand powers off a virtual machine
type PowerOffCommand struct {
	VMName string
}

func (c *PowerOffCommand) Validate() error {
	if c.VMName == "" {
		return NewValidationError("vm-name is required")
	}
	return nil
}

func (c *PowerOffCommand) Execute(client *ssh.Client) error {
	// TODO: Implement VM power off via SSH
	return nil
}
