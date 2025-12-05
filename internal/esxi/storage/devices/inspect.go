package devices

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// inspectStorageDevices inspects a specific storage device with detailed information
func inspectStorageDevices(params *config.Params) (string, error) {
	if params.DatastoreName == "" {
		return "", fmt.Errorf("--datastore-name parameter is required (use device name)")
	}

	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query detailed device info
	detailCmd := fmt.Sprintf("esxcli storage core device get -d '%s' 2>/dev/null", params.DatastoreName)
	detailOutput, err := mgr.RunCommand(detailCmd)

	if err != nil || strings.TrimSpace(detailOutput) == "" {
		return "", fmt.Errorf("storage device '%s' not found", params.DatastoreName)
	}

	device := &config.StorageDeviceInfo{Name: params.DatastoreName}

	// Parse device details from output
	lines := strings.Split(detailOutput, "\n")
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

	return utils.FormatAsJSON(device)
}

func init() {
	config.RegisterFunc("inspect-storage-devices", inspectStorageDevices)
}
