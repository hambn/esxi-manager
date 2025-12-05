package vswitch

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type VSwitchInfo struct {
	Name          string
	NumPorts      string
	UsedPorts     string
	MTU           string
	Uplinks       string
	PortGroups    string
}

type ListVSwitchesCommand struct {
	params *config.Params
}

func (c *ListVSwitchesCommand) Validate() error {
	return nil
}

func (c *ListVSwitchesCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network vswitch standard list")
	if err != nil {
		return fmt.Errorf("failed to list vswitches: %w", err)
	}

	vswitches := c.parseVSwitches(output)
	c.displayVSwitches(vswitches)
	return nil
}

func (c *ListVSwitchesCommand) parseVSwitches(output string) []VSwitchInfo {
	var vswitches []VSwitchInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var currentSwitch *VSwitchInfo

	// Use shared parser for multi-line key-value format
	// recordFunc: called when a new unindented line is found (new record/switch)
	recordFunc := func() {
		if currentSwitch != nil {
			vswitches = append(vswitches, *currentSwitch)
		}
	}

	// parseFunc: called for each key-value pair in indented lines
	parseFunc := func(key, value string) {
		if currentSwitch != nil {
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

	// Process each line manually to handle switch name initialization
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if this is an unindented line (new switch)
		if !utils.IsLineIndented(line) {
			// Save previous switch and start new one
			recordFunc()
			currentSwitch = &VSwitchInfo{Name: trimmed}
		} else if currentSwitch != nil && strings.Contains(trimmed, ":") {
			// Parse key-value pair
			key, value, ok := utils.ParseKeyValueLine(trimmed)
			if ok {
				parseFunc(key, value)
			}
		}
	}

	// Don't forget the last switch
	recordFunc()

	return vswitches
}

func (c *ListVSwitchesCommand) displayVSwitches(vswitches []VSwitchInfo) {
	if len(vswitches) == 0 {
		fmt.Println("No vswitches found")
		return
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tTotal Ports\tUsed Ports\tMTU\tUplinks\tPort Groups\n")

	for _, vs := range vswitches {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			vs.Name, vs.NumPorts, vs.UsedPorts, vs.MTU, vs.Uplinks, vs.PortGroups,
		)
	}

	w.Flush()
	fmt.Print(buf.String())
}

func init() {
	config.Register("list-vswitches", func(params *config.Params) config.CommandInterface {
		return &ListVSwitchesCommand{params: params}
	})
}
