package networking

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type ListPortGroupsCommand struct {
	host *config.ESXiHost
}

func (c *ListPortGroupsCommand) Validate() error {
	return nil
}

func (c *ListPortGroupsCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard portgroup list")
	if err != nil {
		return fmt.Errorf("failed to list port groups: %w", err)
	}

	fmt.Print(output)
	return nil
}

func init() {
	config.Register("list-port-groups", func(params config.CommandParams, host *config.ESXiHost) config.CommandInterface {
		return &ListPortGroupsCommand{host: host}
	})
}
