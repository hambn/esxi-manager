package adapters

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listStorageAdapters lists all storage adapters (HBA adapters)
func listStorageAdapters(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query storage adapters using esxcli
	output, err := mgr.RunCommand("esxcli storage core adapter list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.StorageAdapterInfo{})
	}

	var adapters []config.StorageAdapterInfo
	lines := strings.Split(output, "\n")

	// Skip header (first two lines: header and separator line "---  ---  ...")
	for idx, line := range lines {
		if idx < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		adapter := config.StorageAdapterInfo{
			Name:   fields[0],
			Driver: fields[1],
		}

		// Add remaining fields if available (Link State, UID)
		if len(fields) > 2 {
			adapter.Status = fields[2]
		}
		if len(fields) > 3 {
			adapter.Type = fields[3]
		}

		adapters = append(adapters, adapter)
	}

	return utils.FormatAsJSON(adapters)
}

func init() {
	config.RegisterFunc("list-storage-adapters", listStorageAdapters)
}
