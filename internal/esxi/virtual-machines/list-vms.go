package virtualmachines

import (
	"bytes"
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListVMsCommand lists all virtual machines on the ESXi host
type ListVMsCommand struct {
	host *config.ESXiHost
}

func (c *ListVMsCommand) Validate() error {
	return nil
}

func (c *ListVMsCommand) Execute() error {
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

	// List VMs using vim-cmd command
	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run("vim-cmd vmsvc/getallvms"); err != nil {
		return fmt.Errorf("failed to execute list VMs command: %w", err)
	}

	vms := stdout.String()
	fmt.Print(vms)
	return nil
}

// init registers the list-vms command
func init() {
	command.Register("list-vms", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListVMsCommand{
			host: host,
		}
	})
}
