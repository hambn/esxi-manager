package vmknics

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
	"github.com/esxi-manager/esxi-manager/internal/presenter"
)

// ListVMKnics represents the list VM kernel NICs command
type ListVMKnics struct {
	params *config.Params
}

// NewListVMKnics creates a new list VM kernel NICs command instance
func NewListVMKnics(params *config.Params) *ListVMKnics {
	return &ListVMKnics{params: params}
}

// Validate checks that required parameters are present
func (l *ListVMKnics) Validate() error {
	return nil
}

// Execute gathers and outputs all VM kernel NICs as JSON
func (l *ListVMKnics) Execute() error {
	mgr, err := utils.NewSSHManager(l.params)
	if err != nil {
		return common.NewConnectionError(l.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return common.NewConnectionError(l.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all VM kernel NICs
	knics, err := l.gatherVMKnics(mgr)
	if err != nil {
		return err
	}

	// Format as JSON
	formatted, err := presenter.FormatAsJSON(knics)
	if err != nil {
		return common.WrapError(err, "failed to format VM kernel NICs")
	}

	fmt.Println(formatted)
	return nil
}

// gatherVMKnics gathers all VM kernel NICs
func (l *ListVMKnics) gatherVMKnics(mgr *utils.SSHManager) ([]config.VMKnicInfo, error) {
	// Query VM kernel NICs using esxcli
	output, err := mgr.RunCommand("esxcli network ip interface list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return []config.VMKnicInfo{}, nil
	}

	knics := l.parseVMKnicListing(output)

	// Enrich each KNIC with detailed information
	for idx := range knics {
		l.enrichVMKnicDetails(mgr, &knics[idx])
	}

	return knics, nil
}

// parseVMKnicListing parses esxcli network ip interface list output
func (l *ListVMKnics) parseVMKnicListing(output string) []config.VMKnicInfo {
	var knics []config.VMKnicInfo
	lines := strings.Split(output, "\n")

	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		knic := config.VMKnicInfo{
			Name: fields[0],
		}

		knics = append(knics, knic)
	}

	return knics
}

// enrichVMKnicDetails gathers detailed information for each KNIC
func (l *ListVMKnics) enrichVMKnicDetails(mgr *utils.SSHManager, knic *config.VMKnicInfo) {
	// Query detailed KNIC info
	detailCmd := fmt.Sprintf("esxcli network ip interface get -i '%s' 2>/dev/null", knic.Name)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		l.parseVMKnicDetail(knic, detailOutput)
	}

	// Query IPv4 address info
	ipv4Cmd := fmt.Sprintf("esxcli network ip interface ipv4 get -i '%s' 2>/dev/null", knic.Name)
	ipv4Output, _ := mgr.RunCommand(ipv4Cmd)

	if ipv4Output != "" {
		l.parseIPv4Info(knic, ipv4Output)
	}
}

// parseVMKnicDetail extracts KNIC details from esxcli output
func (l *ListVMKnics) parseVMKnicDetail(knic *config.VMKnicInfo, output string) {
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

// parseIPv4Info extracts IPv4 address information
func (l *ListVMKnics) parseIPv4Info(knic *config.VMKnicInfo, output string) {
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

// Register registers the list-vmknics command
func init() {
	config.Register("list-vmknics", func(params *config.Params) config.CommandInterface {
		return NewListVMKnics(params)
	})
}
