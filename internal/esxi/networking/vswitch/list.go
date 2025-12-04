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

	// Parse output line by line - each vswitch entry is multi-line
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var currentSwitch *VSwitchInfo

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Lines starting with vSwitch name (not indented) mark a new switch
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if currentSwitch != nil {
				vswitches = append(vswitches, *currentSwitch)
			}
			currentSwitch = &VSwitchInfo{
				Name: trimmed,
			}
		} else if currentSwitch != nil {
			// Parse indented fields with key: value format
			if strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

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
	}

	// Don't forget the last switch
	if currentSwitch != nil {
		vswitches = append(vswitches, *currentSwitch)
	}

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
