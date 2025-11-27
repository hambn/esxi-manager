package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type TestConnectionCommand struct {
	host *config.ESXiHost
}

func (c *TestConnectionCommand) Validate() error {
	return nil
}

func (c *TestConnectionCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	if _, err := manager.RunCommand("echo 'ESXi connection test'"); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	fmt.Println("Connection successful")
	return nil
}

func init() {
	command.Register("test-connection", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &TestConnectionCommand{host: host}
	})
}
