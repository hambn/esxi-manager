package datastores

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectStorageDatastores inspects a specific datastore with comprehensive details
func inspectStorageDatastores(params *config.Params) (string, error) {
	if params.DatastoreName == "" {
		return "", fmt.Errorf("--datastore-name parameter is required")
	}

	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query datastores using esxcli
	output, err := mgr.RunCommand("esxcli storage filesystem list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("failed to query datastores: %w", err)
	}

	var datastores []config.DatastoreInfo
	lines := strings.Split(output, "\n")
	var currentMount string
	var currentUUID string
	var currentType string
	var currentPath string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Mount Point:") {
			currentMount = strings.TrimSpace(strings.TrimPrefix(line, "Mount Point:"))
			currentPath = currentMount
		} else if strings.HasPrefix(line, "UUID:") {
			currentUUID = strings.TrimSpace(strings.TrimPrefix(line, "UUID:"))
		} else if strings.HasPrefix(line, "Type:") {
			currentType = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))

			if currentMount != "" {
				ds := config.DatastoreInfo{
					Name:       currentUUID,
					Path:       currentPath,
					Type:       currentType,
					UUID:       currentUUID,
					MountPoint: currentMount,
					Mounted:    true,
					Accessible: true,
				}
				datastores = append(datastores, ds)

				currentMount = ""
				currentUUID = ""
				currentType = ""
				currentPath = ""
			}
		}
	}

	// Find the requested datastore
	var targetDatastore config.DatastoreInfo
	found := false
	for _, ds := range datastores {
		if ds.Name == params.DatastoreName || ds.UUID == params.DatastoreName {
			targetDatastore = ds
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("datastore '%s' not found", params.DatastoreName)
	}

	// Query detailed info using esxcli
	detailCmd := fmt.Sprintf("esxcli storage filesystem info -l '%s' 2>/dev/null", targetDatastore.UUID)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		detailLines := strings.Split(detailOutput, "\n")
		for _, line := range detailLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.Contains(line, "Type:") && strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					targetDatastore.Type = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Version:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					targetDatastore.Version = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Local:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					val := strings.TrimSpace(parts[1])
					targetDatastore.Local = val == "Yes" || val == "true" || val == "1"
				}
			} else if strings.Contains(line, "Block size:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					targetDatastore.BlockSize = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Hosts:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetDatastore.HostCount)
				}
			} else if strings.Contains(line, "Virtual Machines:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &targetDatastore.VMCount)
				}
			} else if strings.Contains(line, "Extent") && strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					extentStr := strings.TrimSpace(parts[1])
					if extentStr != "" && extentStr != "Unknown" {
						targetDatastore.Extents = []string{extentStr}
					}
				}
			} else if strings.Contains(line, "Accessible:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					val := strings.TrimSpace(parts[1])
					targetDatastore.Accessible = val == "Yes" || val == "true" || val == "1"
				}
			}
		}
	}

	// Get capacity info using df
	if targetDatastore.Path != "" {
		dfCmd := fmt.Sprintf("df -h %s 2>/dev/null | tail -1", targetDatastore.Path)
		dfOutput, _ := mgr.RunCommand(dfCmd)
		if dfOutput != "" {
			fields := strings.Fields(dfOutput)
			if len(fields) >= 4 {
				parseSizeValue(fields[1], &targetDatastore.Capacity)
				parseSizeValue(fields[2], &targetDatastore.UsedSpace)
				parseSizeValue(fields[3], &targetDatastore.FreeSpace)

				if targetDatastore.Capacity > 0 {
					targetDatastore.UsagePercent = float64(targetDatastore.UsedSpace) / float64(targetDatastore.Capacity) * 100
				}
			}
		}
	}

	return utils.FormatAsJSON(targetDatastore)
}

// parseSizeValue parses human-readable size (1K, 1M, 1G) to bytes
func parseSizeValue(sizeStr string, target *int64) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var size float64
	var multiplier int64 = 1

	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else {
		fmt.Sscanf(sizeStr, "%f", &size)
	}

	*target = int64(size * float64(multiplier))
}

func init() {
	config.RegisterFunc("inspect-storage-datastores", inspectStorageDatastores)
}
