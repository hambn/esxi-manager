package networking

import (
	"bytes"
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListPortGroupsCommand lists all port groups on the ESXi host
type ListPortGroupsCommand struct {
	host *config.ESXiHost
}

func (c *ListPortGroupsCommand) Validate() error {
	return nil
}

func (c *ListPortGroupsCommand) Execute() error {
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

	// List port groups using esxcli command
	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run("esxcli network vswitch standard portgroup list"); err != nil {
		return fmt.Errorf("failed to execute list port groups command: %w", err)
	}

	portGroups := stdout.String()
	fmt.Print(portGroups)
	return nil
}

// init registers the list-port-groups command
func init() {
	command.Register("list-port-groups", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListPortGroupsCommand{
			host: host,
		}
	})
}
