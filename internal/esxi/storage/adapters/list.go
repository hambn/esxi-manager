package adapters

import (
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// listStorageAdapters lists all storage adapters with detailed information
func listStorageAdapters(params *config.Params) (string, error) {
	// Note: esxcli storage adapter list is not available in this ESXi version
	// Try querying FC adapters and iSCSI adapters instead
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	var adapters []config.StorageAdapterInfo

	// Try FC adapters
	fcOutput, _ := mgr.RunCommand("esxcli storage san fc list 2>/dev/null")
	if fcOutput != "" && strings.TrimSpace(fcOutput) != "" {
		lines := strings.Split(fcOutput, "\n")
		for idx, line := range lines {
			if idx == 0 || strings.TrimSpace(line) == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}
			adapters = append(adapters, config.StorageAdapterInfo{
				Name: fields[0],
				Type: "FC",
			})
		}
	}

	// Try iSCSI adapters
	iscsiOutput, _ := mgr.RunCommand("esxcli storage san iscsi list 2>/dev/null")
	if iscsiOutput != "" && strings.TrimSpace(iscsiOutput) != "" {
		lines := strings.Split(iscsiOutput, "\n")
		for idx, line := range lines {
			if idx == 0 || strings.TrimSpace(line) == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}
			adapters = append(adapters, config.StorageAdapterInfo{
				Name: fields[0],
				Type: "iSCSI",
			})
		}
	}

	return utils.FormatAsJSON(adapters)
}

func init() {
	config.RegisterFunc("list-storage-adapters", listStorageAdapters)
}
