package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type TestConnectionCommand struct {
	params *config.Params
}

func (c *TestConnectionCommand) Validate() error {
	return nil
}

func (c *TestConnectionCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.params)
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
	config.Register("test-connection", func(params *config.Params) config.CommandInterface {
		return &TestConnectionCommand{params: params}
	})
}
