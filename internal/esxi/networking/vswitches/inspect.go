package vswitches

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// VSwitchInspect represents the vswitch inspect command
type VSwitchInspect struct {
	params *config.Params
}

// NewVSwitchInspect creates a new vswitch inspect command instance
func NewVSwitchInspect(params *config.Params) *VSwitchInspect {
	return &VSwitchInspect{params: params}
}

// Validate checks that required parameters are present
func (v *VSwitchInspect) Validate() error {
	return nil
}

// Execute gathers and outputs comprehensive vswitch information as JSON
func (v *VSwitchInspect) Execute() (string, error) {
	if v.params.VSwitchName == "" {
		return "", fmt.Errorf("--vswitch-name parameter is required for vswitch-inspect")
	}

	mgr, err := utils.NewSSHManager(v.params)
	if err != nil {
		return "", common.NewConnectionError(v.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(v.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all vswitches
	vswitches, err := v.gatherVSwitches(mgr)
	if err != nil {
		return "", err
	}

	// Find the requested vswitch
	var targetVSwitch config.VSwitchInfo
	found := false
	for _, vs := range vswitches {
		if vs.Name == v.params.VSwitchName {
			targetVSwitch = vs
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("vswitch '%s' not found", v.params.VSwitchName)
	}

	// Enrich single vswitch with policies
	enrichedVSwitch, err := v.enrichVSwitchWithPolicies(mgr, targetVSwitch)
	if err != nil {
		common.Debug("enrich vswitch", "error", err.Error())
		// Don't fail if enrichment fails
	}

	// Format as JSON
	formatted, err := utils.FormatAsJSON(enrichedVSwitch)
	if err != nil {
		return "", common.WrapError(err, "failed to format vswitch info")
	}

	return formatted, nil
}

// gatherVSwitches gathers vswitch information
func (v *VSwitchInspect) gatherVSwitches(mgr *utils.SSHManager) ([]config.VSwitchInfo, error) {
	output, err := mgr.RunCommand("esxcli network vswitch standard list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return nil, fmt.Errorf("failed to query vswitches: %w", err)
	}

	return v.parseVSwitchListing(output), nil
}

// parseVSwitchListing parses esxcli vswitch output
func (v *VSwitchInspect) parseVSwitchListing(output string) []config.VSwitchInfo {
	var vswitches []config.VSwitchInfo
	var currentVSwitch config.VSwitchInfo
	var inVSwitch bool

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if inVSwitch && currentVSwitch.Name != "" {
				vswitches = append(vswitches, currentVSwitch)
				currentVSwitch = config.VSwitchInfo{}
				inVSwitch = false
			}
			continue
		}

		if !inVSwitch && strings.HasPrefix(line, "Name:") {
			inVSwitch = true
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentVSwitch.Name = strings.TrimSpace(parts[1])
			}
		} else if inVSwitch {
			if strings.HasPrefix(line, "Type:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					currentVSwitch.Type = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "Port groups:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &currentVSwitch.PortGroupCount)
				}
			} else if strings.HasPrefix(line, "Uplinks:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					uplinksStr := strings.TrimSpace(parts[1])
					if uplinksStr != "" && uplinksStr != "Unknown" {
						currentVSwitch.Uplinks = strings.FieldsFunc(uplinksStr, func(r rune) bool {
							return r == ',' || r == ' '
						})
					}
				}
			} else if strings.HasPrefix(line, "MTU:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &currentVSwitch.MTU)
				}
			} else if strings.HasPrefix(line, "Ports:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					portStr := strings.TrimSpace(parts[1])
					parsePortsField(portStr, &currentVSwitch.Ports, &currentVSwitch.AvailablePorts)
				}
			} else if strings.HasPrefix(line, "Link discovery:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					currentVSwitch.LinkDiscovery = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "Attached VMs:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					vmsStr := strings.TrimSpace(parts[1])
					parseVMsField(vmsStr, &currentVSwitch.AttachedVMs, &currentVSwitch.ActiveVMs)
				}
			}
		}
	}

	if inVSwitch && currentVSwitch.Name != "" {
		vswitches = append(vswitches, currentVSwitch)
	}

	return vswitches
}

// enrichVSwitchWithPolicies queries security, teaming, and shaping policies for a single vswitch
func (v *VSwitchInspect) enrichVSwitchWithPolicies(mgr *utils.SSHManager, vs config.VSwitchInfo) (config.VSwitchInfo, error) {
	if vs.Name == "" {
		return vs, nil
	}

	// Query security policy
	securityCmd := fmt.Sprintf("esxcli network vswitch standard policy security get -v '%s' 2>/dev/null", vs.Name)
	secOutput, _ := mgr.RunCommand(securityCmd)
	if secOutput != "" {
		vs.Security = v.parseSecurityPolicy(secOutput)
	}

	// Query NIC teaming policy
	teamingCmd := fmt.Sprintf("esxcli network vswitch standard policy failover get -v '%s' 2>/dev/null", vs.Name)
	teamOutput, _ := mgr.RunCommand(teamingCmd)
	if teamOutput != "" {
		vs.NICTeaming = v.parseTeamingPolicy(teamOutput)
	}

	// Query shaping policy
	shapingCmd := fmt.Sprintf("esxcli network vswitch standard policy shaping get -v '%s' 2>/dev/null", vs.Name)
	shapOutput, _ := mgr.RunCommand(shapingCmd)
	if shapOutput != "" {
		vs.Shaping = v.parseShapingPolicy(shapOutput)
	}

	return vs, nil
}

// parseSecurityPolicy extracts security policy from esxcli output
func (v *VSwitchInspect) parseSecurityPolicy(output string) *config.SecurityPolicy {
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

// parseTeamingPolicy extracts teaming policy from esxcli output
func (v *VSwitchInspect) parseTeamingPolicy(output string) *config.NICTeamingPolicy {
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

// parseShapingPolicy extracts shaping policy from esxcli output
func (v *VSwitchInspect) parseShapingPolicy(output string) *config.ShapingPolicy {
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

// parsePortsField parses "Ports: 1536 (1529 available)" format
func parsePortsField(portsStr string, totalPorts, availablePorts *int) {
	if portsStr == "" {
		return
	}

	// Extract total ports
	var total int
	if _, err := fmt.Sscanf(portsStr, "%d", &total); err == nil && total > 0 {
		*totalPorts = total
	}

	// Extract available ports from (XXXX available)
	if idx := strings.Index(portsStr, "("); idx != -1 {
		endIdx := strings.Index(portsStr[idx:], ")")
		if endIdx != -1 {
			availStr := portsStr[idx+1 : idx+endIdx]
			var avail int
			if _, err := fmt.Sscanf(availStr, "%d", &avail); err == nil && avail > 0 {
				*availablePorts = avail
			}
		}
	}
}

// parseVMsField parses "X (Y active)" format for attached VMs
func parseVMsField(vmsStr string, totalVMs, activeVMs *int) {
	if vmsStr == "" {
		return
	}

	// Extract total VMs
	var total int
	if _, err := fmt.Sscanf(vmsStr, "%d", &total); err == nil && total > 0 {
		*totalVMs = total
	}

	// Extract active VMs from (X active)
	if idx := strings.Index(vmsStr, "("); idx != -1 {
		endIdx := strings.Index(vmsStr[idx:], ")")
		if endIdx != -1 {
			activeStr := vmsStr[idx+1 : idx+endIdx]
			var active int
			if _, err := fmt.Sscanf(activeStr, "%d", &active); err == nil && active >= 0 {
				*activeVMs = active
			}
		}
	}
}

// Register registers the vswitch-inspect command
func init() {
	config.Register("vswitch-inspect", func(params *config.Params) config.CommandInterface {
		return NewVSwitchInspect(params)
	})
}
