package netstacks

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingNetstacks lists all network stacks on the ESXi host with gateway and DNS info
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

// getNetstackDetails retrieves detailed configuration for a specific netstack including gateway and DNS
func getNetstackDetails(mgr *utils.SSHManager, stackName string) (*config.NetworkStackInfo, error) {
	stack := &config.NetworkStackInfo{
		Name: stackName,
	}

	// Get basic netstack info
	output, err := mgr.RunCommand("esxcli network ip netstack get -N " + stackName + " 2>/dev/null")
	if err == nil && strings.TrimSpace(output) != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || !strings.Contains(trimmed, ":") {
				continue
			}

			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "Enabled":
				stack.Enabled = val == "true" || val == "True"
			}
		}
	}

	// Get IPv4 gateway from routes
	routeOutput, err := mgr.RunCommand("esxcli network ip route ipv4 list -N " + stackName + " 2>/dev/null")
	if err == nil && strings.TrimSpace(routeOutput) != "" {
		lines := strings.Split(routeOutput, "\n")
		// Skip first 2 lines (header and separator)
		for idx, line := range lines {
			if idx < 2 {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			fields := strings.Fields(line)
			// Format: Network Netmask Gateway Interface Source
			// Look for "default" network to get the gateway
			if len(fields) >= 3 && fields[0] == "default" {
				stack.DNSResolver = fields[2] // Gateway is in 3rd field
				break
			}
		}
	}

	// Get DNS servers
	dnsOutput, err := mgr.RunCommand("esxcli network ip dns server list -N " + stackName + " 2>/dev/null")
	if err == nil && strings.TrimSpace(dnsOutput) != "" {
		// Format: DNSServers: 192.168.0.1, 114.114.114.114
		if strings.Contains(dnsOutput, "DNSServers:") {
			parts := strings.SplitN(dnsOutput, ":", 2)
			if len(parts) == 2 {
				stack.DNSResolver = strings.TrimSpace(parts[1])
			}
		}
	}

	return stack, nil
}

func init() {
	config.RegisterFunc("list-networking-netstacks", listNetworkingNetstacks)
}
