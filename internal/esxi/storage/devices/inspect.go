package devices

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// InspectStorageDevice represents the inspect storage device command
type InspectStorageDevice struct {
	params *config.Params
}

// NewInspectStorageDevice creates a new inspect storage device command instance
func NewInspectStorageDevice(params *config.Params) *InspectStorageDevice {
	return &InspectStorageDevice{params: params}
}

// Validate checks that required parameters are present
func (i *InspectStorageDevice) Validate() error {
	if i.params.DatastoreName == "" {
		return fmt.Errorf("--datastore-name parameter is required for inspect-storage-device (use device name)")
	}
	return nil
}

// Execute gathers and outputs detailed storage device information as JSON
func (i *InspectStorageDevice) Execute() (string, error) {
	mgr, err := utils.NewSSHManager(i.params)
	if err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(i.params.ESXiHostURI, "failed to connect", err)
	}

	// Get device details
	device := &config.StorageDeviceInfo{Name: i.params.DatastoreName}

	// Query detailed device info
	detailCmd := fmt.Sprintf("esxcli storage core device get -d '%s' 2>/dev/null", device.Name)
	detailOutput, err := mgr.RunCommand(detailCmd)

	if err != nil || strings.TrimSpace(detailOutput) == "" {
		return "", fmt.Errorf("storage device '%s' not found", i.params.DatastoreName)
	}

	i.parseDeviceFullDetail(device, detailOutput)

	// Format as JSON
	formatted, err := utils.FormatAsJSON(device)
	if err != nil {
		return "", common.WrapError(err, "failed to format storage device info")
	}

	return formatted, nil
}

// parseDeviceFullDetail extracts comprehensive device details
func (i *InspectStorageDevice) parseDeviceFullDetail(device *config.StorageDeviceInfo, output string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Display Name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.DisplayName = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Model:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Vendor:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.Vendor = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Serial:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.SerialNumber = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Size:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				sizeStr := strings.TrimSpace(parts[1])
				var sizeBytes int64
				fmt.Sscanf(sizeStr, "%d", &sizeBytes)
				device.Size = sizeBytes
				device.SizeGB = fmt.Sprintf("%.2f GB", float64(sizeBytes)/1024/1024/1024)
			}
		} else if strings.Contains(line, "Status:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.Status = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Physical Location:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.PhysicalLocation = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "SCSI Level:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				var level int
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &level)
				device.ScsiLevel = level
			}
		}
	}
}

// inspectStorageDevices inspects storage devices
func inspectStorageDevices(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	cmd := NewInspectStorageDevice(params)
	return cmd.Execute()
}

func init() {
	config.RegisterFunc("inspect-storage-devices", inspectStorageDevices)
}
