package netstacks

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingNetstacks lists all network stacks on the ESXi host
func listNetworkingNetstacks(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query network stacks using esxcli
	output, err := mgr.RunCommand("esxcli network ip netstack list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.NetworkStackInfo{})
	}

	var stacks []config.NetworkStackInfo
	lines := strings.Split(output, "\n")

	// Parse netstack listing - netstack names are non-indented lines
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isIndented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')

		if trimmed == "" {
			continue
		}

		// Check if this is a netstack name line (not indented)
		if !isIndented {
			// This is a netstack name, get its full configuration
			stackName := trimmed
			stack, err := getNetstackDetails(mgr, stackName)
			if err == nil && stack != nil {
				stacks = append(stacks, *stack)
			}
		}
	}

	return utils.FormatAsJSON(stacks)
}

// getNetstackDetails retrieves detailed configuration for a specific netstack
func getNetstackDetails(mgr *utils.SSHManager, stackName string) (*config.NetworkStackInfo, error) {
	output, err := mgr.RunCommand("esxcli network ip netstack get -N " + stackName + " 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return nil, err
	}

	stack := &config.NetworkStackInfo{
		Name: stackName,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.Contains(trimmed, ":") {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			stack.Name = val
		case "Enabled":
			stack.Enabled = val == "true" || val == "True"
		case "Max Connections":
			// Not directly mapped in NetworkStackInfo, but could be stored if needed
		case "IPv6 Enabled":
			// Could map to a new field if added to struct
		}
	}

	return stack, nil
}

func init() {
	config.RegisterFunc("list-networking-netstacks", listNetworkingNetstacks)
}
