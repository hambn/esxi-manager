package commands

import (
	"golang.org/x/crypto/ssh"
)

// Command defines the interface all commands must implement
// Specific command types are implemented in their respective packages:
// - internal/esxi/vm for VM operations
// - internal/esxi/network for network operations
// - internal/esxi/storage for storage operations
type Command interface {
	Validate() error
	Execute(client *ssh.Client) error
}
