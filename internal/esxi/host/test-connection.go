package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// testHostConnection tests the connection to the ESXi host
func testHostConnection(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	if _, err := manager.RunCommand("echo 'ESXi connection test'"); err != nil {
		return "", fmt.Errorf("connection test failed: %w", err)
	}

	return "Connection successful", nil
}

func init() {
	config.RegisterFunc("test-host-connection", testHostConnection)
}
