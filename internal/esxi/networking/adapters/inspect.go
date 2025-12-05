package adapters

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectNetworkingAdapters inspects network adapters (stub - would need specific VM/host implementation)
func inspectNetworkingAdapters(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Get all network adapters (stub - would need specific VM/host implementation)
	adapters := []config.NetworkAdapterInfo{}

	return utils.FormatAsJSON(adapters)
}

func init() {
	config.RegisterFunc("inspect-networking-adapters", inspectNetworkingAdapters)
}
