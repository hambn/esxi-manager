package vmknics

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectNetworkingVmknics inspects a specific VM kernel NIC with detailed information
func inspectNetworkingVmknics(params *config.Params) (string, error) {
	if params.VMName == "" {
		return "", fmt.Errorf("--vm-name parameter is required for inspect-networking-vmknics (use interface name like vmk0)")
	}

	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Get KNIC details
	knic := &config.VMKnicInfo{Name: params.VMName}

	// Query detailed KNIC info
	detailCmd := fmt.Sprintf("esxcli network ip interface get -i '%s' 2>/dev/null", knic.Name)
	detailOutput, err := mgr.RunCommand(detailCmd)

	if err != nil || strings.TrimSpace(detailOutput) == "" {
		return "", fmt.Errorf("VM kernel NIC '%s' not found", params.VMName)
	}

	// Parse KNIC details
	lines := strings.Split(detailOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "MTU:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &knic.MTU)
			}
		} else if strings.Contains(line, "MAC Address:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				knic.MAC = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Enabled:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				knic.Enabled = val == "true" || val == "True" || val == "yes"
			}
		} else if strings.Contains(line, "Portgroup:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				knic.Portgroup = strings.TrimSpace(parts[1])
			}
		}
	}

	// Get IPv4 info
	ipv4Cmd := fmt.Sprintf("esxcli network ip interface ipv4 get -i '%s' 2>/dev/null", knic.Name)
	ipv4Output, _ := mgr.RunCommand(ipv4Cmd)
	if ipv4Output != "" {
		lines := strings.Split(ipv4Output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.Contains(line, "IPv4 Address:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					knic.IPv4Address = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "IPv4 Netmask:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					knic.IPv4Netmask = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "IPv4 Gateway:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					knic.IPv4Gateway = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "DHCP:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					val := strings.TrimSpace(parts[1])
					knic.IPv4DHCP = val == "true" || val == "True" || val == "yes"
				}
			}
		}
	}

	return utils.FormatAsJSON(knic)
}

func init() {
	config.RegisterFunc("inspect-networking-vmknics", inspectNetworkingVmknics)
}
