package portgroup

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type PortGroupInfo struct {
	Name          string
	VSwitch       string
	ActiveClients string
	VLANID        string
}

type ListPortGroupsCommand struct {
	params *config.Params
}

func (c *ListPortGroupsCommand) Validate() error {
	return nil
}

func (c *ListPortGroupsCommand) Execute() (string, error) {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard portgroup list")
	if err != nil {
		return "", fmt.Errorf("failed to list port groups: %w", err)
	}

	portgroups := c.parsePortGroups(output)
	return c.formatPortGroups(portgroups)
}

func (c *ListPortGroupsCommand) parsePortGroups(output string) []PortGroupInfo {
	var portgroups []PortGroupInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Skip header line and separator line
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Last 3 fields are: VLANID, ActiveClients, VSwitch (in reverse order from the end)
		// Everything before that is the port group name (may have spaces)
		// Use shared parser to extract multi-word name
		vlanID := fields[len(fields)-1]
		activeClients := fields[len(fields)-2]
		vswitch := fields[len(fields)-3]
		name := utils.ParseMultiWordName(fields, 3)

		pg := PortGroupInfo{
			Name:          name,
			VSwitch:       vswitch,
			ActiveClients: activeClients,
			VLANID:        vlanID,
		}
		portgroups = append(portgroups, pg)
	}

	return portgroups
}

func (c *ListPortGroupsCommand) formatPortGroups(portgroups []PortGroupInfo) (string, error) {
	if len(portgroups) == 0 {
		// Return empty JSON array
		return utils.FormatAsJSON([]PortGroupInfo{})
	}
	return utils.FormatAsJSON(portgroups)
}

func init() {
	config.Register("list-portgroups", func(params *config.Params) config.CommandInterface {
		return &ListPortGroupsCommand{params: params}
	})
}
