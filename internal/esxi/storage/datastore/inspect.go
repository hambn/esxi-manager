package datastore

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
	"github.com/esxi-manager/esxi-manager/internal/presenter"
)

// DatastoreInspect represents the datastore inspect command
type DatastoreInspect struct {
	params *config.Params
}

// NewDatastoreInspect creates a new datastore inspect command instance
func NewDatastoreInspect(params *config.Params) *DatastoreInspect {
	return &DatastoreInspect{params: params}
}

// Validate checks that required parameters are present
func (d *DatastoreInspect) Validate() error {
	// At least one datastore identifier should be provided
	// This can be extended to support multiple ways to identify datastores
	return nil
}

// Execute gathers and outputs comprehensive datastore information as JSON
func (d *DatastoreInspect) Execute() error {
	if d.params.DatastoreName == "" {
		return fmt.Errorf("--datastore-name parameter is required for datastore-inspect")
	}

	mgr, err := utils.NewSSHManager(d.params)
	if err != nil {
		return common.NewConnectionError(d.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return common.NewConnectionError(d.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all datastores
	datastores, err := d.gatherDatastores(mgr)
	if err != nil {
		return err
	}

	// Find the requested datastore
	var targetDatastore config.DatastoreInfo
	found := false
	for _, ds := range datastores {
		// Match by name or UUID
		if ds.Name == d.params.DatastoreName || ds.UUID == d.params.DatastoreName {
			targetDatastore = ds
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("datastore '%s' not found", d.params.DatastoreName)
	}

	// Enrich single datastore with metadata
	enrichedDatastore, err := d.enrichDatastoreWithMetadata(mgr, targetDatastore)
	if err != nil {
		common.Debug("enrich datastore", "error", err.Error())
		// Don't fail if enrichment fails - return basic info anyway
	}

	// Format as JSON
	formatted, err := presenter.FormatAsJSON(enrichedDatastore)
	if err != nil {
		return common.WrapError(err, "failed to format datastore info")
	}

	fmt.Println(formatted)
	return nil
}

// gatherDatastores gathers basic datastore information
func (d *DatastoreInspect) gatherDatastores(mgr *utils.SSHManager) ([]config.DatastoreInfo, error) {
	// Query datastores using esxcli
	output, err := mgr.RunCommand("esxcli storage filesystem list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return nil, fmt.Errorf("failed to query datastores: %w", err)
	}

	return d.parseDatastoreListing(output), nil
}

// parseDatastoreListing parses esxcli datastore output
func (d *DatastoreInspect) parseDatastoreListing(output string) []config.DatastoreInfo {
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
					Name:      currentUUID,
					Path:      currentPath,
					Type:      currentType,
					UUID:      currentUUID,
					MountPoint: currentMount,
					Mounted:    true,
					Accessible: true,
				}
				datastores = append(datastores, ds)

				// Reset for next datastore
				currentMount = ""
				currentUUID = ""
				currentType = ""
				currentPath = ""
			}
		}
	}

	return datastores
}

// enrichDatastoreWithMetadata queries additional datastore metadata for a single datastore
func (d *DatastoreInspect) enrichDatastoreWithMetadata(mgr *utils.SSHManager, ds config.DatastoreInfo) (config.DatastoreInfo, error) {
	if ds.UUID == "" {
		return ds, nil
	}

	// Query detailed info using esxcli
	detailCmd := fmt.Sprintf("esxcli storage filesystem info -l '%s' 2>/dev/null", ds.UUID)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		// Parse detailed information
		ds = d.parseDatastoreDetail(ds, detailOutput)
	}

	// Get capacity info using df
	if ds.Path != "" {
		dfCmd := fmt.Sprintf("df -h %s 2>/dev/null | tail -1", ds.Path)
		dfOutput, _ := mgr.RunCommand(dfCmd)
		if dfOutput != "" {
			ds = d.parseDatastoreCapacity(ds, dfOutput)
		}
	}

	return ds, nil
}

// parseDatastoreDetail parses detailed datastore information
func (d *DatastoreInspect) parseDatastoreDetail(ds config.DatastoreInfo, output string) config.DatastoreInfo {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Type:") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ds.Type = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ds.Version = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Local:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				ds.Local = val == "Yes" || val == "true" || val == "1"
			}
		} else if strings.Contains(line, "Block size:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ds.BlockSize = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Hosts:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var hostCount int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &hostCount)
				if hostCount > 0 {
					ds.HostCount = hostCount
				}
			}
		} else if strings.Contains(line, "Virtual Machines:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var vmCount int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &vmCount)
				if vmCount > 0 {
					ds.VMCount = vmCount
				}
			}
		} else if strings.Contains(line, "Extent") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				extentStr := strings.TrimSpace(parts[1])
				if extentStr != "" && extentStr != "Unknown" {
					ds.Extents = []string{extentStr}
				}
			}
		} else if strings.Contains(line, "Accessible:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				ds.Accessible = val == "Yes" || val == "true" || val == "1"
			}
		}
	}

	return ds
}

// parseDatastoreCapacity parses df output for capacity information
func (d *DatastoreInspect) parseDatastoreCapacity(ds config.DatastoreInfo, dfOutput string) config.DatastoreInfo {
	fields := strings.Fields(dfOutput)
	if len(fields) < 4 {
		return ds
	}

	// df -h format: Filesystem Size Used Avail Use% Mounted on
	// fields[0] = filesystem, fields[1] = size, fields[2] = used, fields[3] = available
	parseSizeValue(fields[1], &ds.Capacity)
	parseSizeValue(fields[2], &ds.UsedSpace)
	parseSizeValue(fields[3], &ds.FreeSpace)

	// Calculate usage percentage
	if ds.Capacity > 0 {
		ds.UsagePercent = float64(ds.UsedSpace) / float64(ds.Capacity) * 100
	}

	return ds
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

// Register registers the datastore-inspect command
func init() {
	config.Register("datastore-inspect", func(params *config.Params) config.CommandInterface {
		return NewDatastoreInspect(params)
	})
}
