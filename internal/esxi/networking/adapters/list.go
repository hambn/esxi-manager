package adapters

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type NetworkAdapterListInfo struct {
	Name        string
	PCI         string
	Driver      string
	AdminStatus string
	LinkStatus  string
	Speed       string
	Duplex      string
	MACAddress  string
}

// listNetworkAdapters is the ONE function that does everything
func listNetworkAdapters(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network nic list")
	if err != nil {
		return "", fmt.Errorf("failed to list network adapters: %w", err)
	}

	var adapters []NetworkAdapterListInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Skip header line and separator
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		adapters = append(adapters, NetworkAdapterListInfo{
			Name:        fields[0],
			PCI:         fields[1],
			Driver:      fields[2],
			AdminStatus: fields[3],
			LinkStatus:  fields[4],
			Speed:       fields[5],
			Duplex:      fields[6],
			MACAddress:  fields[7],
		})
	}

	return utils.FormatAsJSON(adapters)
}

func init() {
	config.RegisterFunc("list-networking-adapters", listNetworkAdapters)
}
