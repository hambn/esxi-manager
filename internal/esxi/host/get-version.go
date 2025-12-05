package host

import (
	"fmt"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// getHostVersion gets the ESXi host version
func getHostVersion(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("vmware -v")
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return output, nil
}

func init() {
	config.RegisterFunc("get-host-version", getHostVersion)
}
