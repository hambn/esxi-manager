package devices

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listStorageDevices lists all storage devices
func listStorageDevices(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Query storage devices using esxcli
	output, err := mgr.RunCommand("esxcli storage core device list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]config.StorageDeviceInfo{})
	}

	var devices []config.StorageDeviceInfo
	lines := strings.Split(output, "\n")
	var currentDevice *config.StorageDeviceInfo

	// Parse device listing - device names start with "mpx." and are not indented
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isIndented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')

		if trimmed == "" {
			continue
		}

		// Check if this is a device name line (starts with mpx. and not indented)
		if !isIndented && strings.HasPrefix(trimmed, "mpx.") {
			// Save previous device if any
			if currentDevice != nil {
				devices = append(devices, *currentDevice)
			}
			// Start new device
			currentDevice = &config.StorageDeviceInfo{
				Name: trimmed,
			}
		} else if !isIndented && currentDevice == nil {
			// Skip header lines that aren't indented and aren't device names
			continue
		} else if isIndented && currentDevice != nil {
			// Parse detail line
			if strings.HasPrefix(trimmed, "Display Name:") && !strings.Contains(trimmed, "Settable") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentDevice.DisplayName = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Don't forget the last device
	if currentDevice != nil {
		devices = append(devices, *currentDevice)
	}

	return utils.FormatAsJSON(devices)
}

func init() {
	config.RegisterFunc("list-storage-devices", listStorageDevices)
}
