package firewall

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingFirewall lists all firewall rules on the ESXi host
func listNetworkingFirewall(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query firewall rulesets using esxcli
	output, err := mgr.RunCommand("esxcli network firewall ruleset list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.FirewallInfo{})
	}

	var firewallRules []config.FirewallInfo
	lines := strings.Split(output, "\n")

	// Skip header (first two lines: header and separator line "---  ---  ...")
	for idx, line := range lines {
		if idx < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Extract name and enabled status
		rule := config.FirewallInfo{
			Name: fields[0],
		}

		// Second field is "true" or "false" for enabled status
		if len(fields) > 1 {
			rule.Enabled = fields[1] == "true"
		}

		firewallRules = append(firewallRules, rule)
	}

	return utils.FormatAsJSON(firewallRules)
}

func init() {
	config.RegisterFunc("list-networking-firewall", listNetworkingFirewall)
}
