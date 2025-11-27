package network

import (
	"golang.org/x/crypto/ssh"
)

// Command defines the interface network commands must implement
type Command interface {
	Validate() error
	Execute(client *ssh.Client) error
}

// CreateVSwitchCommand parameters for creating a virtual switch
type CreateVSwitchCommand struct {
	Name    string
	MTU     int
	Uplinks []string
}

func (c *CreateVSwitchCommand) Validate() error {
	if c.Name == "" {
		return NewValidationError("vswitch-name is required")
	}
	if c.MTU <= 0 {
		c.MTU = 1500 // default MTU
	}
	return nil
}

func (c *CreateVSwitchCommand) Execute(client *ssh.Client) error {
	// TODO: Implement vswitch creation via SSH
	return nil
}

// DeleteVSwitchCommand parameters for deleting a virtual switch
type DeleteVSwitchCommand struct {
	Name string
}

func (c *DeleteVSwitchCommand) Validate() error {
	if c.Name == "" {
		return NewValidationError("vswitch-name is required")
	}
	return nil
}

func (c *DeleteVSwitchCommand) Execute(client *ssh.Client) error {
	// TODO: Implement vswitch deletion via SSH
	return nil
}

// CreatePortgroupCommand parameters for creating a port group
type CreatePortgroupCommand struct {
	Name        string
	VSwitchName string
	VLAN        int
}

func (c *CreatePortgroupCommand) Validate() error {
	if c.Name == "" {
		return NewValidationError("portgroup-name is required")
	}
	if c.VSwitchName == "" {
		return NewValidationError("vswitch-name is required")
	}
	if c.VLAN < 0 || c.VLAN > 4094 {
		return NewValidationError("vlan must be between 0 and 4094")
	}
	return nil
}

func (c *CreatePortgroupCommand) Execute(client *ssh.Client) error {
	// TODO: Implement portgroup creation via SSH
	return nil
}

// DeletePortgroupCommand parameters for deleting a port group
type DeletePortgroupCommand struct {
	Name string
}

func (c *DeletePortgroupCommand) Validate() error {
	if c.Name == "" {
		return NewValidationError("portgroup-name is required")
	}
	return nil
}

func (c *DeletePortgroupCommand) Execute(client *ssh.Client) error {
	// TODO: Implement portgroup deletion via SSH
	return nil
}
