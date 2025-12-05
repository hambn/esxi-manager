package netstacks

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// Stack mapping of CLI names to display names and their order
var netstackNames = map[string]string{
	"defaultTcpipStack": "Default TCP/IP stack",
	"vmotion":           "vMotion stack",
	"provisioning":      "Provisioning stack",
	"ops":               "ops",
	"mirror":            "mirror",
}

// netstackOrder defines the order stacks should be displayed
var netstackOrder = []string{
	"defaultTcpipStack",
	"vmotion",
	"provisioning",
	"ops",
	"mirror",
}

// listNetworkingNetstacks lists all network stacks on the ESXi host with gateway and DNS info
func listNetworkingNetstacks(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query network stacks using esxcli to get list of existing stacks
	output, err := mgr.RunCommand("esxcli network ip netstack list 2>/dev/null")
	existingStacks := make(map[string]bool)
	if err == nil && strings.TrimSpace(output) != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "Key") && !strings.HasPrefix(trimmed, "Name") && !strings.HasPrefix(trimmed, "State") {
				existingStacks[trimmed] = true
			}
		}
	}

	var stacks []config.NetworkStackInfo

	// Check all known stacks in order
	for _, stackName := range netstackOrder {
		stack := &config.NetworkStackInfo{
			Name: netstackNames[stackName],
		}

		// Only query if stack exists
		if existingStacks[stackName] {
			// Get IPv4 gateway address
			ipv4GW := getIPv4Gateway(mgr, stackName)
			// ipv6GW := getIPv6Gateway(mgr, stackName) // TODO: Store when struct supports ipv6_gateway
			// preferredDNS, alternateDNS := getDNSServers(mgr, stackName) // TODO: Store when struct supports preferred/alternate dns

			// Store in appropriate fields
			stack.Instance = 1 // Indicates stack exists
			stack.DNSResolver = ipv4GW // Using DNSResolver field for IPv4 gateway
			if ipv4GW != "" && ipv4GW != "--" {
				stack.Enabled = true
			}
		}

		stacks = append(stacks, *stack)
	}

	return utils.FormatAsJSON(stacks)
}

// getIPv4Gateway retrieves the IPv4 default gateway for a netstack
func getIPv4Gateway(mgr *utils.SSHManager, stackName string) string {
	routeOutput, _ := mgr.RunCommand("esxcli network ip route ipv4 list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(routeOutput) != "" {
		lines := strings.Split(routeOutput, "\n")
		for idx, line := range lines {
			if idx < 2 {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 3 && fields[0] == "default" {
				return fields[2]
			}
		}
	}
	return "--"
}

// getIPv6Gateway retrieves the IPv6 default gateway for a netstack
func getIPv6Gateway(mgr *utils.SSHManager, stackName string) string {
	routeOutput, _ := mgr.RunCommand("esxcli network ip route ipv6 list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(routeOutput) != "" {
		lines := strings.Split(routeOutput, "\n")
		for idx, line := range lines {
			if idx < 2 {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 3 && fields[0] == "default" {
				return fields[2]
			}
		}
	}
	return "--"
}

// getDNSServers retrieves the DNS servers for a netstack
func getDNSServers(mgr *utils.SSHManager, stackName string) (string, string) {
	dnsOutput, _ := mgr.RunCommand("esxcli network ip dns server list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(dnsOutput) == "" {
		return "--", "--"
	}

	// Format: DNSServers: 192.168.0.1, 114.114.114.114
	if strings.Contains(dnsOutput, "DNSServers:") {
		parts := strings.SplitN(dnsOutput, ":", 2)
		if len(parts) == 2 {
			dnsStr := strings.TrimSpace(parts[1])
			dnsServers := strings.Split(dnsStr, ",")
			preferred := "--"
			alternate := "--"
			if len(dnsServers) >= 1 {
				preferred = strings.TrimSpace(dnsServers[0])
			}
			if len(dnsServers) >= 2 {
				alternate = strings.TrimSpace(dnsServers[1])
			}
			return preferred, alternate
		}
	}
	return "--", "--"
}

func init() {
	config.RegisterFunc("list-networking-netstacks", listNetworkingNetstacks)
}
