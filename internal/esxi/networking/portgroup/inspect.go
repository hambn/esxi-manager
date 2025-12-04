package portgroup

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
	"github.com/esxi-manager/esxi-manager/internal/presenter"
)

// PortGroupInspect represents the portgroup inspect command
type PortGroupInspect struct {
	params *config.Params
}

// NewPortGroupInspect creates a new portgroup inspect command instance
func NewPortGroupInspect(params *config.Params) *PortGroupInspect {
	return &PortGroupInspect{params: params}
}

// Validate checks that required parameters are present
func (p *PortGroupInspect) Validate() error {
	return nil
}

// Execute gathers and outputs comprehensive port group information as JSON
func (p *PortGroupInspect) Execute() error {
	if p.params.PortgroupName == "" {
		return fmt.Errorf("--portgroup-name parameter is required for portgroup-inspect")
	}

	mgr, err := utils.NewSSHManager(p.params)
	if err != nil {
		return common.NewConnectionError(p.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return common.NewConnectionError(p.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all port groups
	portgroups, err := p.gatherPortGroups(mgr)
	if err != nil {
		return err
	}

	// Find the requested port group
	var targetPortGroup config.PortGroupInfo
	found := false
	for _, pg := range portgroups {
		if pg.Name == p.params.PortgroupName {
			targetPortGroup = pg
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("portgroup '%s' not found", p.params.PortgroupName)
	}

	// Enrich single port group with policies
	enrichedPortGroup, err := p.enrichPortGroupWithPolicies(mgr, targetPortGroup)
	if err != nil {
		common.Debug("enrich portgroup", "error", err.Error())
		// Don't fail if enrichment fails
	}

	// Format as JSON
	formatted, err := presenter.FormatAsJSON(enrichedPortGroup)
	if err != nil {
		return common.WrapError(err, "failed to format portgroup info")
	}

	fmt.Println(formatted)
	return nil
}

// gatherPortGroups gathers port group information
func (p *PortGroupInspect) gatherPortGroups(mgr *utils.SSHManager) ([]config.PortGroupInfo, error) {
	output, err := mgr.RunCommand("esxcli network vswitch standard portgroup list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return nil, fmt.Errorf("failed to query port groups: %w", err)
	}

	return p.parsePortGroupListing(output), nil
}

// parsePortGroupListing parses esxcli port group output
func (p *PortGroupInspect) parsePortGroupListing(output string) []config.PortGroupInfo {
	var portgroups []config.PortGroupInfo

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
		activeClients := 0
		vlanID := 0

		fmt.Sscanf(fields[2], "%d", &activeClients)
		fmt.Sscanf(fields[3], "%d", &vlanID)

		pg := config.PortGroupInfo{
			Name:          pgName,
			VSwitch:       vswitchName,
			VLANID:        vlanID,
			ActiveClients: activeClients,
		}
		portgroups = append(portgroups, pg)
	}

	return portgroups
}

// enrichPortGroupWithPolicies queries detailed information and policies for a single port group
func (p *PortGroupInspect) enrichPortGroupWithPolicies(mgr *utils.SSHManager, pg config.PortGroupInfo) (config.PortGroupInfo, error) {
	if pg.Name == "" {
		return pg, nil
	}

	// Query detailed port group info
	pgDetailCmd := fmt.Sprintf("esxcli network vswitch standard portgroup get -p '%s' 2>/dev/null", pg.Name)
	pgDetailOutput, _ := mgr.RunCommand(pgDetailCmd)

	if pgDetailOutput != "" {
		pg = p.parsePortGroupDetail(pg, pgDetailOutput)
	}

	// Query security policy
	securityCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy security get -p '%s' 2>/dev/null", pg.Name)
	secOutput, _ := mgr.RunCommand(securityCmd)
	if secOutput != "" {
		pg.Security = p.parseSecurityPolicy(secOutput)
	}

	// Query NIC teaming policy
	teamingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy failover get -p '%s' 2>/dev/null", pg.Name)
	teamOutput, _ := mgr.RunCommand(teamingCmd)
	if teamOutput != "" {
		pg.NICTeaming = p.parseTeamingPolicy(teamOutput)
	}

	// Query shaping policy
	shapingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy shaping get -p '%s' 2>/dev/null", pg.Name)
	shapOutput, _ := mgr.RunCommand(shapingCmd)
	if shapOutput != "" {
		pg.Shaping = p.parseShapingPolicy(shapOutput)
	}

	return pg, nil
}

// parsePortGroupDetail parses detailed port group information
func (p *PortGroupInspect) parsePortGroupDetail(pg config.PortGroupInfo, output string) config.PortGroupInfo {
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
				pg.Accessible = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Virtual machines") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &pg.VMCount)
			}
		} else if strings.Contains(line, "Active ports") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &pg.ActivePorts)
			}
		}
	}

	return pg
}

// parseSecurityPolicy extracts security policy
func (p *PortGroupInspect) parseSecurityPolicy(output string) *config.SecurityPolicy {
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
func (p *PortGroupInspect) parseTeamingPolicy(output string) *config.NICTeamingPolicy {
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
func (p *PortGroupInspect) parseShapingPolicy(output string) *config.ShapingPolicy {
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

// Register registers the portgroup-inspect command
func init() {
	config.Register("portgroup-inspect", func(params *config.Params) config.CommandInterface {
		return NewPortGroupInspect(params)
	})
}
