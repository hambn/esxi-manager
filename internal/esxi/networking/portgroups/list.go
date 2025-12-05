package portgroups

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

// listNetworkingPortgroups is the ONE function that does everything
func listNetworkingPortgroups(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard portgroup list")
	if err != nil {
		return "", fmt.Errorf("failed to list port groups: %w", err)
	}

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
		vlanID := fields[len(fields)-1]
		activeClients := fields[len(fields)-2]
		vswitch := fields[len(fields)-3]
		name := utils.ParseMultiWordName(fields, 3)

		portgroups = append(portgroups, PortGroupInfo{
			Name:          name,
			VSwitch:       vswitch,
			ActiveClients: activeClients,
			VLANID:        vlanID,
		})
	}

	return utils.FormatAsJSON(portgroups)
}

func init() {
	config.RegisterFunc("list-networking-portgroups", listNetworkingPortgroups)
}
