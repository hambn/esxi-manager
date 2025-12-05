package adapters

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listStorageAdapters lists all storage adapters with detailed information
func listStorageAdapters(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query storage adapters using esxcli
	output, err := mgr.RunCommand("esxcli storage adapter list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.StorageAdapterInfo{})
	}

	var adapters []config.StorageAdapterInfo
	lines := strings.Split(output, "\n")

	// Skip header and parse each line
	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		adapter := config.StorageAdapterInfo{
			Name:   fields[0],
			Type:   fields[1],
			Status: "",
		}

		// Query detailed adapter info
		detailCmd := fmt.Sprintf("esxcli storage adapter info -a '%s' 2>/dev/null", adapter.Name)
		detailOutput, _ := mgr.RunCommand(detailCmd)

		if detailOutput != "" {
			detailLines := strings.Split(detailOutput, "\n")
			for _, detailLine := range detailLines {
				detailLine = strings.TrimSpace(detailLine)
				if detailLine == "" {
					continue
				}

				if strings.Contains(detailLine, "Driver:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						adapter.Driver = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Model:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						adapter.Model = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Vendor:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						adapter.Vendor = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Status:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						adapter.Status = strings.TrimSpace(parts[1])
					}
				}
			}
		}

		adapters = append(adapters, adapter)
	}

	return utils.FormatAsJSON(adapters)
}

func init() {
	config.RegisterFunc("list-storage-adapters", listStorageAdapters)
}
