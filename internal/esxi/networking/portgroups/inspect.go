package portgroups

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectNetworkingPortgroups inspects a specific portgroup with detailed information
func inspectNetworkingPortgroups(params *config.Params) (string, error) {
	if params.PortgroupName == "" {
		return "", fmt.Errorf("--portgroup-name parameter is required for inspect-networking-portgroups")
	}

	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Get all port groups
	output, err := mgr.RunCommand("esxcli network vswitch standard portgroup list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("failed to query port groups: %w", err)
	}

	// Parse port groups listing
	var portgroups []config.PortGroupInfo
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		vlanID := 0
		activeClients := 0
		fmt.Sscanf(fields[len(fields)-1], "%d", &vlanID)
		fmt.Sscanf(fields[len(fields)-2], "%d", &activeClients)

		vswitchName := fields[len(fields)-3]
		pgName := strings.Join(fields[:len(fields)-3], " ")

		pg := config.PortGroupInfo{
			Name:          pgName,
			VSwitch:       vswitchName,
			VLANID:        vlanID,
			ActiveClients: activeClients,
		}
		portgroups = append(portgroups, pg)
	}

	// Find the requested port group
	var targetPortGroup config.PortGroupInfo
	found := false
	for _, pg := range portgroups {
		if pg.Name == params.PortgroupName {
			targetPortGroup = pg
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("portgroup '%s' not found", params.PortgroupName)
	}

	// Enrich with detailed information
	if targetPortGroup.Name != "" {
		// Query detailed port group info
		pgDetailCmd := fmt.Sprintf("esxcli network vswitch standard portgroup get -p '%s' 2>/dev/null", targetPortGroup.Name)
		pgDetailOutput, _ := mgr.RunCommand(pgDetailCmd)

		if pgDetailOutput != "" {
			lines := strings.Split(pgDetailOutput, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				if strings.Contains(line, "Accessible") && strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						val := strings.TrimSpace(parts[1])
						targetPortGroup.Accessible = val == "Yes" || val == "true" || val == "1"
					}
				} else if strings.Contains(line, "Virtual machines") && strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetPortGroup.VMCount)
					}
				} else if strings.Contains(line, "Active ports") && strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetPortGroup.ActivePorts)
					}
				}
			}
		}

		// Query security policy
		securityCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy security get -p '%s' 2>/dev/null", targetPortGroup.Name)
		secOutput, _ := mgr.RunCommand(securityCmd)
		if secOutput != "" {
			policy := &config.SecurityPolicy{}
			lines := strings.Split(secOutput, "\n")
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
			targetPortGroup.Security = policy
		}

		// Query NIC teaming policy
		teamingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy failover get -p '%s' 2>/dev/null", targetPortGroup.Name)
		teamOutput, _ := mgr.RunCommand(teamingCmd)
		if teamOutput != "" {
			policy := &config.NICTeamingPolicy{}
			lines := strings.Split(teamOutput, "\n")
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
			targetPortGroup.NICTeaming = policy
		}

		// Query shaping policy
		shapingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy shaping get -p '%s' 2>/dev/null", targetPortGroup.Name)
		shapOutput, _ := mgr.RunCommand(shapingCmd)
		if shapOutput != "" {
			policy := &config.ShapingPolicy{}
			lines := strings.Split(shapOutput, "\n")
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
			targetPortGroup.Shaping = policy
		}
	}

	return utils.FormatAsJSON(targetPortGroup)
}

func init() {
	config.RegisterFunc("inspect-networking-portgroups", inspectNetworkingPortgroups)
}
