package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// TestConnectionCommand tests the SSH connection to ESXi host
type TestConnectionCommand struct {
	host *config.ESXiHost
}

func (c *TestConnectionCommand) Validate() error {
	return nil
}

func (c *TestConnectionCommand) Execute() error {
	// Manage connection internally
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return NewExecutionError("failed to create SSH manager: " + err.Error())
	}
	defer manager.Close()

	// Connect to host
	if err := manager.Connect(); err != nil {
		return NewExecutionError("failed to connect to ESXi host: " + err.Error())
	}

	// Get client
	client, err := manager.GetClient()
	if err != nil {
		return NewExecutionError("failed to get SSH client: " + err.Error())
	}

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

	fmt.Println("Connection successful")
	return nil
}

// init registers the test-connection command
func init() {
	command.Register("test-connection", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &TestConnectionCommand{
			host: host,
		}
	})
}
