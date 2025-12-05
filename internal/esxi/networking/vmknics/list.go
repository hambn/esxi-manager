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

	// Parse KNIC listing with details (all details are in the list output)
	var knics []config.VMKnicInfo
	lines := strings.Split(output, "\n")
	var currentKnic *config.VMKnicInfo

	for _, line := range lines {
		isIndented := strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if !isIndented {
			// This is a new interface name line
			if currentKnic != nil {
				knics = append(knics, *currentKnic)
			}
			currentKnic = &config.VMKnicInfo{
				Name: trimmed,
			}
		} else if currentKnic != nil {
			// Parse the detail line (indented)
			if strings.Contains(trimmed, "MAC Address:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentKnic.MAC = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(trimmed, "Enabled:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					val := strings.TrimSpace(parts[1])
					currentKnic.Enabled = val == "true" || val == "True" || val == "yes"
				}
			} else if strings.Contains(trimmed, "Portgroup:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentKnic.Portgroup = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(trimmed, "MTU:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &currentKnic.MTU)
				}
			}
		}
	}

	// Don't forget the last knic
	if currentKnic != nil {
		knics = append(knics, *currentKnic)
	}

	return utils.FormatAsJSON(knics)
}

func init() {
	config.RegisterFunc("list-networking-vmknics", listNetworkingVmknics)
}
