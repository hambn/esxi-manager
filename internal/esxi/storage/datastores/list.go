package datastores

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

// listStorageDatastores lists all datastores on the ESXi host
func listStorageDatastores(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("esxcli storage filesystem list")
	if err != nil {
		return "", fmt.Errorf("failed to list datastores: %w", err)
	}

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

	return utils.FormatAsJSON(datastores)
}

func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func init() {
	config.RegisterFunc("list-storage-datastores", listStorageDatastores)
}
