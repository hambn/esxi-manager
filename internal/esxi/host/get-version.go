package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/framework/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type GetVersionCommand struct {
	host *config.ESXiHost
}

func (c *GetVersionCommand) Validate() error {
	return nil
}

func (c *GetVersionCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("vmware -v")
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	fmt.Print(output)
	return nil
}

func init() {
	command.Register("get-version", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &GetVersionCommand{host: host}
	})
}
