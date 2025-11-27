package storage

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type Datastore struct {
	VolumeName string
	Type       string
	Size       int64
	Free       int64
	Mounted    string
}

type ListDatastoresCommand struct {
	host *config.ESXiHost
}

func (c *ListDatastoresCommand) Validate() error {
	return nil
}

func (c *ListDatastoresCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli storage filesystem list")
	if err != nil {
		return fmt.Errorf("failed to list datastores: %w", err)
	}

	datastores := c.parseDatastores(output)
	c.displayDatastores(datastores)
	return nil
}

func (c *ListDatastoresCommand) parseDatastores(output string) []Datastore {
	var datastores []Datastore
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Skip header (first line) and separator (second line)
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		ds := Datastore{
			VolumeName: fields[1],
			Type:       fields[4],
			Size:       toInt64(fields[5]),
			Free:       toInt64(fields[6]),
			Mounted:    fields[3],
		}
		datastores = append(datastores, ds)
	}

	return datastores
}

func (c *ListDatastoresCommand) displayDatastores(datastores []Datastore) {
	if len(datastores) == 0 {
		fmt.Println("No datastores found")
		return
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tType\tSize\tFree\tMounted\n")

	for _, ds := range datastores {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ds.VolumeName,
			ds.Type,
			utils.FormatBytes(ds.Size),
			utils.FormatBytes(ds.Free),
			ds.Mounted,
		)
	}

	w.Flush()
	fmt.Print(buf.String())
}

func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func init() {
	command.Register("list-datastores", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListDatastoresCommand{host: host}
	})
}
