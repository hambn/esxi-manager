package adapters

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// InspectStorageAdapter represents the inspect storage adapter command
type InspectStorageAdapter struct {
	params *config.Params
}

// NewInspectStorageAdapter creates a new inspect storage adapter command instance
func NewInspectStorageAdapter(params *config.Params) *InspectStorageAdapter {
	return &InspectStorageAdapter{params: params}
}

// Validate checks that required parameters are present
func (i *InspectStorageAdapter) Validate() error {
	if i.params.DatastoreName == "" {
		return fmt.Errorf("--datastore-name parameter is required for inspect-storage-adapter (use adapter name)")
	}
	return nil
}

// Execute gathers and outputs detailed storage adapter information as JSON
func (i *InspectStorageAdapter) Execute() (string, error) {
	mgr, err := utils.NewSSHManager(i.params)
	if err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all storage adapters
	allAdapters, err := i.gatherAllAdapters(mgr)
	if err != nil {
		return "", err
	}

	// Find the requested adapter
	var targetAdapter *config.StorageAdapterInfo
	for idx := range allAdapters {
		if allAdapters[idx].Name == i.params.DatastoreName {
			targetAdapter = &allAdapters[idx]
			break
		}
	}

	if targetAdapter == nil {
		return "", fmt.Errorf("storage adapter '%s' not found", i.params.DatastoreName)
	}

	// Enrich with detailed information
	i.enrichAdapterWithFullDetails(mgr, targetAdapter)

	// Format as JSON
	formatted, err := utils.FormatAsJSON(targetAdapter)
	if err != nil {
		return "", common.WrapError(err, "failed to format storage adapter info")
	}

	return formatted, nil
}

// gatherAllAdapters gathers all storage adapters
func (i *InspectStorageAdapter) gatherAllAdapters(mgr *utils.SSHManager) ([]config.StorageAdapterInfo, error) {
	output, err := mgr.RunCommand("esxcli storage adapter list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return []config.StorageAdapterInfo{}, nil
	}

	return i.parseStorageAdapterListing(output), nil
}

// parseStorageAdapterListing parses esxcli storage adapter list output
func (i *InspectStorageAdapter) parseStorageAdapterListing(output string) []config.StorageAdapterInfo {
	var adapters []config.StorageAdapterInfo
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

		adapters = append(adapters, adapter)
	}

	return adapters
}

// enrichAdapterWithFullDetails gathers comprehensive adapter information
func (i *InspectStorageAdapter) enrichAdapterWithFullDetails(mgr *utils.SSHManager, adapter *config.StorageAdapterInfo) {
	// Query detailed adapter info
	detailCmd := fmt.Sprintf("esxcli storage adapter info -a '%s' 2>/dev/null", adapter.Name)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		i.parseAdapterFullDetail(adapter, detailOutput)
	}

	// Query target information (LUN mappings)
	targetCmd := fmt.Sprintf("esxcli storage core adapter target list -a '%s' 2>/dev/null", adapter.Name)
	targetOutput, _ := mgr.RunCommand(targetCmd)

	if targetOutput != "" {
		i.parseAdapterTargets(adapter, targetOutput)
	}
}

// parseAdapterFullDetail extracts comprehensive adapter details
func (i *InspectStorageAdapter) parseAdapterFullDetail(adapter *config.StorageAdapterInfo, output string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Driver:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				adapter.Driver = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Model:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				adapter.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Vendor:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				adapter.Vendor = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Status:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				adapter.Status = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Path Count:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var pathCount int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &pathCount)
				adapter.PathCount = pathCount
			}
		} else if strings.Contains(line, "Queue Length:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var queue int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &queue)
				adapter.Queue = queue
			}
		}
	}
}

// parseAdapterTargets parses target information from adapter
func (i *InspectStorageAdapter) parseAdapterTargets(adapter *config.StorageAdapterInfo, output string) {
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		// Parse first target for Bus, Slot, Target info
		fields := strings.Fields(lines[1])
		if len(fields) > 0 {
			adapter.Target = fields[0]
		}
	}
}

// Register registers the inspect-storage-adapter command
func init() {
	config.Register("inspect-storage-adapter", func(params *config.Params) config.CommandInterface {
		return NewInspectStorageAdapter(params)
	})
}
