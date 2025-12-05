package firewall

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListFirewallRulesCommand lists all firewall rules on the ESXi host
type ListFirewallRulesCommand struct {
	params *config.Params
}

func (c *ListFirewallRulesCommand) Validate() error {
	return nil
}

func (c *ListFirewallRulesCommand) Execute() (string, error) {
	// TODO: Implement firewall rules listing
	// esxcli network firewall ruleset list (for rulesets)
	// esxcli network firewall get (for firewall status)
	return utils.FormatAsJSON([]config.FirewallInfo{})
}

func init() {
	config.Register("list-firewall-rules", func(params *config.Params) config.CommandInterface {
		return &ListFirewallRulesCommand{params: params}
	})
}
