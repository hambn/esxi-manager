package netstacks

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListNetworkStacksCommand lists all network stacks on the ESXi host
type ListNetworkStacksCommand struct {
	params *config.Params
}

func (c *ListNetworkStacksCommand) Validate() error {
	return nil
}

func (c *ListNetworkStacksCommand) Execute() (string, error) {
	// TODO: Implement network stack listing
	// esxcli network ip config get (for default stack)
	// esxcli network ip netstack list (for custom stacks)
	return utils.FormatAsJSON([]config.NetworkStackInfo{})
}

func init() {
	config.Register("list-netstacks", func(params *config.Params) config.CommandInterface {
		return &ListNetworkStacksCommand{params: params}
	})
}
