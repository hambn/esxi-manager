package adapters

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListStorageAdapters represents the list storage adapters command
type ListStorageAdapters struct {
	params *config.Params
}

// NewListStorageAdapters creates a new list storage adapters command instance
func NewListStorageAdapters(params *config.Params) *ListStorageAdapters {
	return &ListStorageAdapters{params: params}
}

// Validate checks that required parameters are present
func (l *ListStorageAdapters) Validate() error {
	return nil
}

// Execute gathers and outputs all storage adapters as JSON
func (l *ListStorageAdapters) Execute() (string, error) {
	mgr, err := utils.NewSSHManager(l.params)
	if err != nil {
		return "", common.NewConnectionError(l.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(l.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all storage adapters
	adapters, err := l.gatherStorageAdapters(mgr)
	if err != nil {
		return "", err
	}

	// Format as JSON
	formatted, err := utils.FormatAsJSON(adapters)
	if err != nil {
		return "", common.WrapError(err, "failed to format storage adapters")
	}

	return formatted, nil
}

// gatherStorageAdapters gathers all storage adapters
func (l *ListStorageAdapters) gatherStorageAdapters(mgr *utils.SSHManager) ([]config.StorageAdapterInfo, error) {
	// Query storage adapters using esxcli
	output, err := mgr.RunCommand("esxcli storage adapter list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return []config.StorageAdapterInfo{}, nil
	}

	adapters := l.parseStorageAdapterListing(output)

	// Enrich each adapter with detailed information
	for idx := range adapters {
		l.enrichAdapterDetails(mgr, &adapters[idx])
	}

	return adapters, nil
}

// parseStorageAdapterListing parses esxcli storage adapter list output
func (l *ListStorageAdapters) parseStorageAdapterListing(output string) []config.StorageAdapterInfo {
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

		adapters = append(adapters, adapter)
	}

	return adapters
}

// enrichAdapterDetails gathers detailed information for each adapter
func (l *ListStorageAdapters) enrichAdapterDetails(mgr *utils.SSHManager, adapter *config.StorageAdapterInfo) {
	// Query detailed adapter info
	detailCmd := fmt.Sprintf("esxcli storage adapter info -a '%s' 2>/dev/null", adapter.Name)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		l.parseAdapterDetail(adapter, detailOutput)
	}
}

// parseAdapterDetail extracts adapter details from esxcli output
func (l *ListStorageAdapters) parseAdapterDetail(adapter *config.StorageAdapterInfo, output string) {
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
		}
	}
}

// listStorageAdapters lists all storage adapters
func listStorageAdapters(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	cmd := NewListStorageAdapters(params)
	return cmd.Execute()
}

func init() {
	config.RegisterFunc("list-storage-adapters", listStorageAdapters)
}
