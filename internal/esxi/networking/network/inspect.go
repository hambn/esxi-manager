package network

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
	"github.com/esxi-manager/esxi-manager/internal/presenter"
)

// NetworkInspect represents the network adapter inspect command
type NetworkInspect struct {
	params *config.Params
}

// NewNetworkInspect creates a new network adapter inspect command instance
func NewNetworkInspect(params *config.Params) *NetworkInspect {
	return &NetworkInspect{params: params}
}

// Validate checks that required parameters are present
func (n *NetworkInspect) Validate() error {
	return nil
}

// Execute gathers and outputs comprehensive network adapter information as JSON
func (n *NetworkInspect) Execute() error {
	mgr, err := utils.NewSSHManager(n.params)
	if err != nil {
		return common.NewConnectionError(n.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return common.NewConnectionError(n.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all network adapters (stub - would need specific VM/host implementation)
	adapters := []config.NetworkAdapterInfo{}

	// Format as JSON
	formatted, err := presenter.FormatAsJSON(adapters)
	if err != nil {
		return common.WrapError(err, "failed to format network adapter info")
	}

	fmt.Println(formatted)
	return nil
}

// GatherNetworkAdaptersForVM gathers network adapters for a specific VM
func (n *NetworkInspect) GatherNetworkAdaptersForVM(mgr *utils.SSHManager, vmID string) ([]config.NetworkAdapterInfo, error) {
	// This would be called by vm-inspect to get network adapters
	// For now, returns empty slice - implementation would parse VMX config
	return []config.NetworkAdapterInfo{}, nil
}

// EnrichNetworkAdaptersWithInfrastructure enriches adapters with vswitch and port group details
func (n *NetworkInspect) EnrichNetworkAdaptersWithInfrastructure(mgr *utils.SSHManager, adapters []config.NetworkAdapterInfo) ([]config.NetworkAdapterInfo, error) {
	// Get all port groups with esxcli
	output, err := mgr.RunCommand("esxcli network vswitch standard portgroup list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		common.Debug("enrich networks", "error", "failed to get portgroups from esxcli")
		return adapters, nil
	}

	// Build a map of portgroup name to vswitch info
	portgroupInfo := make(map[string]map[string]string)
	lines := strings.Split(output, "\n")

	// Skip header line
	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pgName := fields[0]
		vswitchName := fields[1]
		activeClients := fields[2]
		vlanID := fields[3]

		portgroupInfo[pgName] = map[string]string{
			"vswitch":       vswitchName,
			"vlan":          vlanID,
			"activeClients": activeClients,
		}
	}

	// Enrich each network adapter with infrastructure details
	for idx, adapter := range adapters {
		if adapter.Network == "" {
			continue
		}

		if pgInfo, ok := portgroupInfo[adapter.Network]; ok {
			// Basic info
			if vswitchName, ok := pgInfo["vswitch"]; ok {
				adapters[idx].VSwitch = vswitchName
			}
			if vlanStr, ok := pgInfo["vlan"]; ok {
				vlanID := 0
				fmt.Sscanf(vlanStr, "%d", &vlanID)
				if vlanID >= 0 {
					adapters[idx].VLANID = vlanID
				}
			}
			if clientsStr, ok := pgInfo["activeClients"]; ok {
				clients := 0
				fmt.Sscanf(clientsStr, "%d", &clients)
				if clients > 0 {
					adapters[idx].ActiveClients = clients
					adapters[idx].VMCount = clients
				}
			}

			// Query detailed portgroup info
			pgDetailCmd := fmt.Sprintf("esxcli network vswitch standard portgroup get -p '%s' 2>/dev/null", adapter.Network)
			pgDetailOutput, _ := mgr.RunCommand(pgDetailCmd)

			// Parse portgroup details
			if pgDetailOutput != "" {
				adapters[idx] = n.parsePortGroupDetail(adapters[idx], pgDetailOutput)
			}

			// Query security policy
			securityCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy security get -p '%s' 2>/dev/null", adapter.Network)
			secOutput, _ := mgr.RunCommand(securityCmd)
			if secOutput != "" {
				adapters[idx].Security = n.parseSecurityPolicy(secOutput)
			}

			// Query NIC teaming policy
			teamingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy failover get -p '%s' 2>/dev/null", adapter.Network)
			teamOutput, _ := mgr.RunCommand(teamingCmd)
			if teamOutput != "" {
				adapters[idx].NICTeaming = n.parseTeamingPolicy(teamOutput)
			}

			// Query shaping policy
			shapingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy shaping get -p '%s' 2>/dev/null", adapter.Network)
			shapOutput, _ := mgr.RunCommand(shapingCmd)
			if shapOutput != "" {
				adapters[idx].Shaping = n.parseShapingPolicy(shapOutput)
			}
		}
	}

	return adapters, nil
}

// parsePortGroupDetail parses portgroup details
func (n *NetworkInspect) parsePortGroupDetail(adapter config.NetworkAdapterInfo, output string) config.NetworkAdapterInfo {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Accessible") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				adapter.Accessible = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Virtual machines") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &adapter.VMCount)
			}
		} else if strings.Contains(line, "Active ports") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &adapter.ActivePorts)
			}
		}
	}

	return adapter
}

// parseSecurityPolicy extracts security policy
func (n *NetworkInspect) parseSecurityPolicy(output string) *config.SecurityPolicy {
	policy := &config.SecurityPolicy{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Promiscuous") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.AllowPromiscuous = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Forged") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.AllowForgedTx = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "MAC") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.AllowMACChanges = val == "Yes" || val == "true" || val == "1"
			}
		}
	}

	return policy
}

// parseTeamingPolicy extracts teaming policy
func (n *NetworkInspect) parseTeamingPolicy(output string) *config.NICTeamingPolicy {
	policy := &config.NICTeamingPolicy{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Notify") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.NotifySwitches = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Policy") && strings.Contains(line, ":") && !strings.Contains(line, "Reverse") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				policy.Policy = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Reverse") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.ReversePolicy = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Failback") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.Failback = val == "Yes" || val == "true" || val == "1"
			}
		}
	}

	return policy
}

// parseShapingPolicy extracts shaping policy
func (n *NetworkInspect) parseShapingPolicy(output string) *config.ShapingPolicy {
	policy := &config.ShapingPolicy{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Enabled") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				policy.Enabled = val == "Yes" || val == "true" || val == "1"
			}
		}
	}

	return policy
}

// Register registers the network-inspect command
func init() {
	config.Register("network-inspect", func(params *config.Params) config.CommandInterface {
		return NewNetworkInspect(params)
	})
}
