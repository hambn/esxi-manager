package netstacks

import (
	"encoding/json"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// NetworkStack represents a network stack with its configuration
type NetworkStack struct {
	Name         string `json:"name"`
	IPv4Gateway  string `json:"ipv4_gateway"`
	IPv6Gateway  string `json:"ipv6_gateway"`
	PreferredDNS string `json:"preferred_dns"`
	AlternateDNS string `json:"alternate_dns"`
}

// listNetworkingNetstacks lists all network stacks on the ESXi host
func listNetworkingNetstacks(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Get list of network stacks
	output, err := mgr.RunCommand("esxcli network ip netstack list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]NetworkStack{})
	}

	var stacks []NetworkStack
	lines := strings.Split(output, "\n")

	// Parse netstack names - they are non-indented lines that don't contain ":"
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.Contains(trimmed, ":") {
			continue
		}

		stack := NetworkStack{
			Name:         trimmed,
			IPv4Gateway:  getIPv4Gateway(mgr, trimmed),
			IPv6Gateway:  getIPv6Gateway(mgr, trimmed),
			PreferredDNS: "--",
			AlternateDNS: "--",
		}

		// Get DNS servers
		stack.PreferredDNS, stack.AlternateDNS = getDNSServers(mgr, trimmed)

		stacks = append(stacks, stack)
	}

	// Return as JSON
	jsonData, _ := json.MarshalIndent(stacks, "", "  ")
	return string(jsonData), nil
}

func getIPv4Gateway(mgr *utils.SSHManager, stackName string) string {
	output, _ := mgr.RunCommand("esxcli network ip route ipv4 list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(output) == "" {
		return "--"
	}

	lines := strings.Split(output, "\n")
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
	return "--"
}

func getIPv6Gateway(mgr *utils.SSHManager, stackName string) string {
	output, _ := mgr.RunCommand("esxcli network ip route ipv6 list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(output) == "" {
		return "--"
	}

	lines := strings.Split(output, "\n")
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
	return "--"
}

func getDNSServers(mgr *utils.SSHManager, stackName string) (string, string) {
	output, _ := mgr.RunCommand("esxcli network ip dns server list --netstack=" + stackName + " 2>/dev/null")
	if strings.TrimSpace(output) == "" {
		return "--", "--"
	}

	// Format: DNSServers: 192.168.0.1, 114.114.114.114
	if strings.Contains(output, "DNSServers:") {
		parts := strings.SplitN(output, ":", 2)
		if len(parts) == 2 {
			dnsStr := strings.TrimSpace(parts[1])
			servers := strings.Split(dnsStr, ",")

			preferred := "--"
			alternate := "--"

			if len(servers) >= 1 {
				preferred = strings.TrimSpace(servers[0])
			}
			if len(servers) >= 2 {
				alternate = strings.TrimSpace(servers[1])
			}

			return preferred, alternate
		}
	}
	return "--", "--"
}

func init() {
	config.RegisterFunc("list-networking-netstacks", listNetworkingNetstacks)
}
