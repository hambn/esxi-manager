package datastore

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
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
	params *config.Params
}

func (c *ListDatastoresCommand) Validate() error {
	return nil
}

func (c *ListDatastoresCommand) Execute() (string, error) {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli storage filesystem list")
	if err != nil {
		return "", fmt.Errorf("failed to list datastores: %w", err)
	}

	datastores := c.parseDatastores(output)
	return c.formatDatastores(datastores)
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

func (c *ListDatastoresCommand) formatDatastores(datastores []Datastore) (string, error) {
	if len(datastores) == 0 {
		// Return empty JSON array
		return utils.FormatAsJSON([]Datastore{})
	}
	return utils.FormatAsJSON(datastores)
}

func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func init() {
	config.Register("list-datastores", func(params *config.Params) config.CommandInterface {
		return &ListDatastoresCommand{params: params}
	})
}
