package storage

import (
	"bytes"
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListDatastoresCommand lists all datastores on the ESXi host
type ListDatastoresCommand struct {
	host *config.ESXiHost
}

func (c *ListDatastoresCommand) Validate() error {
	return nil
}

func (c *ListDatastoresCommand) Execute() error {
	// Manage connection internally
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	// Connect to host
	if err := manager.Connect(); err != nil {
		return fmt.Errorf("failed to connect to ESXi host: %w", err)
	}

	// Get client
	client, err := manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %w", err)
	}

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// List datastores using esxcli command
	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run("esxcli storage filesystem list"); err != nil {
		return fmt.Errorf("failed to execute list datastores command: %w", err)
	}

	datastores := stdout.String()
	fmt.Print(datastores)
	return nil
}

// init registers the list-datastores command
func init() {
	command.Register("list-datastores", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListDatastoresCommand{
			host: host,
		}
	})
}
