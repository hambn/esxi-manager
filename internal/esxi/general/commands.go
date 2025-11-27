package general

import (
	"golang.org/x/crypto/ssh"

	"github.com/esxi-manager/esxi-manager/internal/common"
)

// TestConnectionCommand tests the SSH connection to ESXi host
type TestConnectionCommand struct{}

func (c *TestConnectionCommand) Validate() error {
	return nil
}

func (c *TestConnectionCommand) Execute(client *ssh.Client) error {
	// Create a session to test the connection
	session, err := client.NewSession()
	if err != nil {
		return NewExecutionError("failed to create SSH session: " + err.Error())
	}
	defer session.Close()

	// Run a simple command to verify connection works
	if err := session.Run("echo 'ESXi connection test successful'"); err != nil {
		return NewExecutionError("failed to execute test command: " + err.Error())
	}

	common.Info("SSH connection test successful")
	return nil
}
