package devices

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// ListStorageDevices represents the list storage devices command
type ListStorageDevices struct {
	params *config.Params
}

// NewListStorageDevices creates a new list storage devices command instance
func NewListStorageDevices(params *config.Params) *ListStorageDevices {
	return &ListStorageDevices{params: params}
}

// Validate checks that required parameters are present
func (l *ListStorageDevices) Validate() error {
	return nil
}

// Execute gathers and outputs all storage devices as JSON
func (l *ListStorageDevices) Execute() (string, error) {
	mgr, err := utils.NewSSHManager(l.params)
	if err != nil {
		return "", common.NewConnectionError(l.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return "", common.NewConnectionError(l.params.ESXiHostURI, "failed to connect", err)
	}

	// Get all storage devices
	devices, err := l.gatherStorageDevices(mgr)
	if err != nil {
		return "", err
	}

	// Format as JSON
	formatted, err := utils.FormatAsJSON(devices)
	if err != nil {
		return "", common.WrapError(err, "failed to format storage devices")
	}

	return formatted, nil
}

// gatherStorageDevices gathers all storage devices
func (l *ListStorageDevices) gatherStorageDevices(mgr *utils.SSHManager) ([]config.StorageDeviceInfo, error) {
	// Query storage devices using esxcli
	output, err := mgr.RunCommand("esxcli storage core device list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return []config.StorageDeviceInfo{}, nil
	}

	devices := l.parseStorageDeviceListing(output)

	// Enrich each device with detailed information
	for idx := range devices {
		l.enrichDeviceDetails(mgr, &devices[idx])
	}

	return devices, nil
}

// parseStorageDeviceListing parses esxcli storage core device list output
func (l *ListStorageDevices) parseStorageDeviceListing(output string) []config.StorageDeviceInfo {
	var devices []config.StorageDeviceInfo
	lines := strings.Split(output, "\n")

	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		device := config.StorageDeviceInfo{
			Name: fields[0],
		}

		devices = append(devices, device)
	}

	return devices
}

// enrichDeviceDetails gathers detailed information for each device
func (l *ListStorageDevices) enrichDeviceDetails(mgr *utils.SSHManager, device *config.StorageDeviceInfo) {
	// Query detailed device info
	detailCmd := fmt.Sprintf("esxcli storage core device get -d '%s' 2>/dev/null", device.Name)
	detailOutput, _ := mgr.RunCommand(detailCmd)

	if detailOutput != "" {
		l.parseDeviceDetail(device, detailOutput)
	}
}

// parseDeviceDetail extracts device details from esxcli output
func (l *ListStorageDevices) parseDeviceDetail(device *config.StorageDeviceInfo, output string) {
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
		} else if strings.Contains(line, "Revision:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				device.Type = strings.TrimSpace(parts[1])
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
		}
	}
}

// Register registers the list-storage-devices command
func init() {
	config.Register("list-storage-devices", func(params *config.Params) config.CommandInterface {
		return NewListStorageDevices(params)
	})
}
