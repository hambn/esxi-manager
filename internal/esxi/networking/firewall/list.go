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

	// Query firewall rules using esxcli
	output, err := mgr.RunCommand("esxcli network firewall ruleset rule list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.FirewallRuleInfo{})
	}

	var rules []config.FirewallRuleInfo
	ruleMap := make(map[string]*config.FirewallRuleInfo)
	lines := strings.Split(output, "\n")

	// Skip header (first two lines: header and separator line "---  ---  ...")
	for idx, line := range lines {
		if idx < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Parse rule: Ruleset Direction Protocol PortType PortBegin PortEnd
		rulesetName := fields[0]
		direction := fields[1]
		protocol := fields[2]
		// portType := fields[3]
		// portBegin := fields[4]
		// portEnd := fields[5] if exists

		// Create or update rule entry for this ruleset
		if _, exists := ruleMap[rulesetName]; !exists {
			ruleMap[rulesetName] = &config.FirewallRuleInfo{
				Name: rulesetName,
			}
		}

		// Set direction and protocol flags
		rule := ruleMap[rulesetName]
		if direction == "Inbound" {
			rule.Inbound = true
		} else if direction == "Outbound" {
			rule.Outbound = true
		}

		if rule.Protocol == "" {
			rule.Protocol = protocol
		} else if !strings.Contains(rule.Protocol, protocol) {
			rule.Protocol += ", " + protocol
		}

		rule.Direction = direction
	}

	// Convert map to slice
	for _, rule := range ruleMap {
		rules = append(rules, *rule)
	}

	return utils.FormatAsJSON(rules)
}

func init() {
	config.RegisterFunc("list-networking-firewall", listNetworkingFirewall)
}
