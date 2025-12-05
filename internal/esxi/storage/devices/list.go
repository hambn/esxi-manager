package devices

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listStorageDevices lists all storage devices with detailed information
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

		// Query detailed device info
		detailCmd := fmt.Sprintf("esxcli storage core device get -d '%s' 2>/dev/null", device.Name)
		detailOutput, _ := mgr.RunCommand(detailCmd)

		if detailOutput != "" {
			detailLines := strings.Split(detailOutput, "\n")
			for _, detailLine := range detailLines {
				detailLine = strings.TrimSpace(detailLine)
				if detailLine == "" {
					continue
				}

				if strings.Contains(detailLine, "Display Name:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						device.DisplayName = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Model:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						device.Model = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Vendor:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						device.Vendor = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Serial:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						device.SerialNumber = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Revision:") {
					parts := strings.SplitN(detailLine, ":", 2)
					if len(parts) == 2 {
						device.Type = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(detailLine, "Size:") {
					parts := strings.SplitN(detailLine, ":", 2)
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

		devices = append(devices, device)
	}

	return utils.FormatAsJSON(devices)
}

func init() {
	config.RegisterFunc("list-storage-devices", listStorageDevices)
}
