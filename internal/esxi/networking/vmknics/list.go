package vmknics

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listNetworkingVmknics lists all VM kernel NICs with detailed information
func listNetworkingVmknics(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query VM kernel NICs using esxcli
	output, err := mgr.RunCommand("esxcli network ip interface list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.VMKnicInfo{})
	}

	// Parse KNIC listing
	var knics []config.VMKnicInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		isIndented := strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if !isIndented {
			knic := config.VMKnicInfo{
				Name: trimmed,
			}
			knics = append(knics, knic)
		}
	}

	// Enrich each KNIC with detailed information
	for idx := range knics {
		// Query detailed KNIC info
		detailCmd := fmt.Sprintf("esxcli network ip interface get -i '%s' 2>/dev/null", knics[idx].Name)
		detailOutput, _ := mgr.RunCommand(detailCmd)

		if detailOutput != "" {
			lines := strings.Split(detailOutput, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				if strings.Contains(line, "MTU:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &knics[idx].MTU)
					}
				} else if strings.Contains(line, "MAC Address:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						knics[idx].MAC = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(line, "Enabled:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						val := strings.TrimSpace(parts[1])
						knics[idx].Enabled = val == "true" || val == "True" || val == "yes"
					}
				} else if strings.Contains(line, "Portgroup:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						knics[idx].Portgroup = strings.TrimSpace(parts[1])
					}
				}
			}
		}

		// Query IPv4 address info
		ipv4Cmd := fmt.Sprintf("esxcli network ip interface ipv4 get -i '%s' 2>/dev/null", knics[idx].Name)
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
						knics[idx].IPv4Address = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(line, "IPv4 Netmask:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						knics[idx].IPv4Netmask = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(line, "IPv4 Gateway:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						knics[idx].IPv4Gateway = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(line, "DHCP:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						val := strings.TrimSpace(parts[1])
						knics[idx].IPv4DHCP = val == "true" || val == "True" || val == "yes"
					}
				}
			}
		}
	}

	return utils.FormatAsJSON(knics)
}

func init() {
	config.RegisterFunc("list-networking-vmknics", listNetworkingVmknics)
}
