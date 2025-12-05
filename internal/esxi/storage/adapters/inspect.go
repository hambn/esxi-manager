package adapters

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectStorageAdapters inspects a specific storage adapter with detailed information
func inspectStorageAdapters(params *config.Params) (string, error) {
	if params.DatastoreName == "" {
		return "", fmt.Errorf("--datastore-name parameter is required (use adapter name)")
	}

	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Get all storage adapters
	output, err := mgr.RunCommand("esxcli storage adapter list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("failed to list storage adapters: %w", err)
	}

	var allAdapters []config.StorageAdapterInfo
	lines := strings.Split(output, "\n")

	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		adapter := config.StorageAdapterInfo{
			Name: fields[0],
			Type: fields[1],
		}

		allAdapters = append(allAdapters, adapter)
	}

	// Find the requested adapter
	var targetAdapter *config.StorageAdapterInfo
	for idx := range allAdapters {
		if allAdapters[idx].Name == params.DatastoreName {
			targetAdapter = &allAdapters[idx]
			break
		}
	}

	if targetAdapter == nil {
		return "", fmt.Errorf("storage adapter '%s' not found", params.DatastoreName)
	}

	// Query detailed adapter info
	detailCmd := fmt.Sprintf("esxcli storage adapter info -a '%s' 2>/dev/null", targetAdapter.Name)
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
					targetAdapter.Driver = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(detailLine, "Model:") {
				parts := strings.SplitN(detailLine, ":", 2)
				if len(parts) == 2 {
					targetAdapter.Model = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(detailLine, "Vendor:") {
				parts := strings.SplitN(detailLine, ":", 2)
				if len(parts) == 2 {
					targetAdapter.Vendor = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(detailLine, "Status:") {
				parts := strings.SplitN(detailLine, ":", 2)
				if len(parts) == 2 {
					targetAdapter.Status = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(detailLine, "Path Count:") {
				parts := strings.SplitN(detailLine, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetAdapter.PathCount)
				}
			} else if strings.Contains(detailLine, "Queue Length:") {
				parts := strings.SplitN(detailLine, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetAdapter.Queue)
				}
			}
		}
	}

	// Query target information (LUN mappings)
	targetCmd := fmt.Sprintf("esxcli storage core adapter target list -a '%s' 2>/dev/null", targetAdapter.Name)
	targetOutput, _ := mgr.RunCommand(targetCmd)

	if targetOutput != "" {
		targetLines := strings.Split(targetOutput, "\n")
		if len(targetLines) > 1 {
			fields := strings.Fields(targetLines[1])
			if len(fields) > 0 {
				targetAdapter.Target = fields[0]
			}
		}
	}

	return utils.FormatAsJSON(targetAdapter)
}

func init() {
	config.RegisterFunc("inspect-storage-adapters", inspectStorageAdapters)
}
