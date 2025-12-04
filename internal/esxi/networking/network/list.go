package network

import (
	"fmt"
	"strings"
	"text/tabwriter"

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

type ListNetworkAdaptersCommand struct {
	params *config.Params
}

func (c *ListNetworkAdaptersCommand) Validate() error {
	return nil
}

func (c *ListNetworkAdaptersCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli network nic list")
	if err != nil {
		return fmt.Errorf("failed to list network adapters: %w", err)
	}

	adapters := c.parseNetworkAdapters(output)
	c.displayNetworkAdapters(adapters)
	return nil
}

func (c *ListNetworkAdaptersCommand) parseNetworkAdapters(output string) []NetworkAdapterListInfo {
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

		adapter := NetworkAdapterListInfo{
			Name:        fields[0],
			PCI:         fields[1],
			Driver:      fields[2],
			AdminStatus: fields[3],
			LinkStatus:  fields[4],
			Speed:       fields[5],
			Duplex:      fields[6],
			MACAddress:  fields[7],
		}

		adapters = append(adapters, adapter)
	}

	return adapters
}

func (c *ListNetworkAdaptersCommand) displayNetworkAdapters(adapters []NetworkAdapterListInfo) {
	if len(adapters) == 0 {
		fmt.Println("No network adapters found")
		return
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tPCI Device\tDriver\tAdmin Status\tLink Status\tSpeed\tDuplex\tMAC Address\n")

	for _, adapter := range adapters {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			adapter.Name, adapter.PCI, adapter.Driver, adapter.AdminStatus,
			adapter.LinkStatus, adapter.Speed, adapter.Duplex, adapter.MACAddress,
		)
	}

	w.Flush()
	fmt.Print(buf.String())
}

func init() {
	config.Register("list-network-adapters", func(params *config.Params) config.CommandInterface {
		return &ListNetworkAdaptersCommand{params: params}
	})
}
