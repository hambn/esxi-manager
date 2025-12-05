package netstacks

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingNetstacks lists all network stacks on the ESXi host
func listNetworkingNetstacks(params *config.Params) (string, error) {
	// TODO: Implement network stack listing
	// esxcli network ip config get (for default stack)
	// esxcli network ip netstack list (for custom stacks)
	return utils.FormatAsJSON([]config.NetworkStackInfo{})
}

func init() {
	config.RegisterFunc("list-networking-netstacks", listNetworkingNetstacks)
}
