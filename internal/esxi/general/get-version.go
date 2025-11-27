package general

import (
	"bytes"

	"github.com/esxi-manager/esxi-manager/internal/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// GetVersionCommand retrieves the ESXi host version information
type GetVersionCommand struct {
	host *config.ESXiHost
}

func (c *GetVersionCommand) Validate() error {
	return nil
}

func (c *GetVersionCommand) Execute() error {
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

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return NewExecutionError("failed to create SSH session: " + err.Error())
	}
	defer session.Close()

	// Get version using vmware -v command
	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run("vmware -v"); err != nil {
		return NewExecutionError("failed to execute version command: " + err.Error())
	}

	version := stdout.String()
	common.Info("ESXi version retrieved", "version", version)
	return nil
}

// init registers the get-version command
func init() {
	command.Register("get-version", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &GetVersionCommand{
			host: host,
		}
	})
}
