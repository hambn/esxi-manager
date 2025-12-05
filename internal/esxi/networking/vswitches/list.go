package vswitches

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type VSwitchInfo struct {
	Name       string
	NumPorts   string
	UsedPorts  string
	MTU        string
	Uplinks    string
	PortGroups string
}

// listNetworkingVswitches lists all vSwitches with their detailed information
func listNetworkingVswitches(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard list")
	if err != nil {
		return "", fmt.Errorf("failed to list vswitches: %w", err)
	}

	var vswitches []VSwitchInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var currentSwitch *VSwitchInfo

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if !utils.IsLineIndented(line) {
			if currentSwitch != nil {
				vswitches = append(vswitches, *currentSwitch)
			}
			currentSwitch = &VSwitchInfo{Name: trimmed}
		} else if currentSwitch != nil && strings.Contains(trimmed, ":") {
			key, value, ok := utils.ParseKeyValueLine(trimmed)
			if ok {
				switch key {
				case "Num Ports":
					currentSwitch.NumPorts = value
				case "Used Ports":
					currentSwitch.UsedPorts = value
				case "MTU":
					currentSwitch.MTU = value
				case "Uplinks":
					currentSwitch.Uplinks = value
				case "Portgroups":
					currentSwitch.PortGroups = value
				}
			}
		}
	}

	if currentSwitch != nil {
		vswitches = append(vswitches, *currentSwitch)
	}

	return utils.FormatAsJSON(vswitches)
}

func init() {
	config.RegisterFunc("list-networking-vswitches", listNetworkingVswitches)
}
