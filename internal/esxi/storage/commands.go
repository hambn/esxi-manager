package storage

import (
	"golang.org/x/crypto/ssh"
)

// Command defines the interface storage commands must implement
type Command interface {
	Validate() error
	Execute(client *ssh.Client) error
}

// ListDatastoresCommand lists all datastores
type ListDatastoresCommand struct{}

func (c *ListDatastoresCommand) Validate() error {
	return nil
}

func (c *ListDatastoresCommand) Execute(client *ssh.Client) error {
	// TODO: Implement datastore listing via SSH
	return nil
}
