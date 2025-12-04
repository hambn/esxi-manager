package portgroup

import (
	"fmt"
	"strings"
	"text/tabwriter"

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

func (c *ListPortGroupsCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard portgroup list")
	if err != nil {
		return fmt.Errorf("failed to list port groups: %w", err)
	}

	portgroups := c.parsePortGroups(output)
	c.displayPortGroups(portgroups)
	return nil
}

func (c *ListPortGroupsCommand) parsePortGroups(output string) []PortGroupInfo {
	var portgroups []PortGroupInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Skip header line and separator line
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		// The output format has fixed column positions:
		// Name (up to ~19 chars), VSwitch (~15 chars), ActiveClients (~14 chars), VLANID
		// We need to handle names with spaces, so we look for the pattern

		// Try to find vswitch and active clients using regex or column positions
		// For now, use a simpler approach: split on multiple spaces to get meaningful fields
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// The last two fields are always ActiveClients (number) and VLANID (number)
		// VSwitch is typically a single word (vSwitch0, vSwitch-test-01, etc.)
		vlanID := fields[len(fields)-1]
		activeClients := fields[len(fields)-2]
		vswitch := fields[len(fields)-3]

		// Everything before the vswitch is the name
		name := strings.Join(fields[:len(fields)-3], " ")

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

func (c *ListPortGroupsCommand) displayPortGroups(portgroups []PortGroupInfo) {
	if len(portgroups) == 0 {
		fmt.Println("No port groups found")
		return
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tVSwitch\tActive Clients\tVLAN ID\n")

	for _, pg := range portgroups {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			pg.Name, pg.VSwitch, pg.ActiveClients, pg.VLANID,
		)
	}

	w.Flush()
	fmt.Print(buf.String())
}

func init() {
	config.Register("list-portgroups", func(params *config.Params) config.CommandInterface {
		return &ListPortGroupsCommand{params: params}
	})
}
