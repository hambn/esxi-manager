package vmknics

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// InspectVMKnic represents the inspect VM kernel NIC command
type InspectVMKnic struct {
	params *config.Params
}

// NewInspectVMKnic creates a new inspect VM kernel NIC command instance
func NewInspectVMKnic(params *config.Params) *InspectVMKnic {
	return &InspectVMKnic{params: params}
}

// Validate checks that required parameters are present
func (i *InspectVMKnic) Validate() error {
	if i.params.VMName == "" {
		return fmt.Errorf("--vm-name parameter is required for inspect-vmknic (use interface name like vmk0)")
	}
	return nil
}

// Execute gathers and outputs detailed VM kernel NIC information as JSON
func (i *InspectVMKnic) Execute() (string, error) {
	mgr, err := utils.NewSSHManager(i.params)
	if err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to connect", err)
	}

	// Get KNIC details
	knic := &config.VMKnicInfo{Name: i.params.VMName}

	// Query detailed KNIC info
	detailCmd := fmt.Sprintf("esxcli network ip interface get -i '%s' 2>/dev/null", knic.Name)
	detailOutput, err := mgr.RunCommand(detailCmd)

	if err != nil || strings.TrimSpace(detailOutput) == "" {
		return "", fmt.Errorf("VM kernel NIC '%s' not found", i.params.VMName)
	}

	i.parseVMKnicFullDetail(knic, detailOutput)

	// Get IPv4 info
	ipv4Cmd := fmt.Sprintf("esxcli network ip interface ipv4 get -i '%s' 2>/dev/null", knic.Name)
	ipv4Output, _ := mgr.RunCommand(ipv4Cmd)
	if ipv4Output != "" {
		i.parseIPv4FullInfo(knic, ipv4Output)
	}

	// Format as JSON
	formatted, err := utils.FormatAsJSON(knic)
	if err != nil {
		return "", common.WrapError(err, "failed to format VM kernel NIC info")
	}

	return formatted, nil
}

// parseVMKnicFullDetail extracts comprehensive KNIC details
func (i *InspectVMKnic) parseVMKnicFullDetail(knic *config.VMKnicInfo, output string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "MTU:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var mtu int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &mtu)
				knic.MTU = mtu
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
}

// parseIPv4FullInfo extracts full IPv4 address information
func (i *InspectVMKnic) parseIPv4FullInfo(knic *config.VMKnicInfo, output string) {
	lines := strings.Split(output, "\n")

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

// Register registers the inspect-vmknic command
func init() {
	config.Register("inspect-vmknic", func(params *config.Params) config.CommandInterface {
		return NewInspectVMKnic(params)
	})
}
