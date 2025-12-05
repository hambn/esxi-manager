package firewall

import (
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingFirewall lists all firewall rules on the ESXi host
func listNetworkingFirewall(params *config.Params) (string, error) {
	// TODO: Implement firewall rules listing
	// esxcli network firewall ruleset list (for rulesets)
	// esxcli network firewall get (for firewall status)
	return utils.FormatAsJSON([]config.FirewallInfo{})
}

func init() {
	config.RegisterFunc("list-networking-firewall", listNetworkingFirewall)
}
