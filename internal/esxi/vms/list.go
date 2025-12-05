package vms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type VMListInfo struct {
	ID        string
	Name      string
	GuestOS   string
	FilePath  string
	Status    string
	CPU       string
	Memory    string
	UsedSpace string
}

// listVMs lists all VMs on the ESXi host with detailed information
func listVMs(params *config.Params) (string, error) {
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	output, err := manager.RunCommand("vim-cmd vmsvc/getallvms")
	if err != nil {
		return "", fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms []VMListInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Vmid") {
			continue
		}

		// Parse VM line
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		vm := &VMListInfo{
			ID:        fields[0],
			Name:      fields[1],
			GuestOS:   extractVMField(fields, 4, "-"),
			Status:    "unknown",
			CPU:       "-",
			Memory:    "-",
			UsedSpace: "-",
		}

		// Extract file path: [datastore] path
		if match := regexp.MustCompile(`\[([^\]]+)\]\s+(.+?)(?:\s+\w+Guest|\s*$)`).FindStringSubmatch(line); len(match) > 2 {
			vm.FilePath = fmt.Sprintf("[%s] %s", match[1], match[2])
		}

		// Enrich VM with details
		summaryCmd := fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vm.ID)
		summaryOutput, err := manager.RunCommand(summaryCmd)
		if err == nil {
			// Extract VM status
			switch {
			case strings.Contains(summaryOutput, `powerState = "poweredOn"`):
				vm.Status = "on"
			case strings.Contains(summaryOutput, `powerState = "poweredOff"`):
				vm.Status = "off"
			case strings.Contains(summaryOutput, `powerState = "suspended"`):
				vm.Status = "suspended"
			}

			// Extract CPU and Memory
			if match := regexp.MustCompile(`numCpu\s*=\s*(\d+)`).FindStringSubmatch(summaryOutput); len(match) > 1 {
				vm.CPU = match[1]
			}

			if match := regexp.MustCompile(`memorySizeMB\s*=\s*(\d+)`).FindStringSubmatch(summaryOutput); len(match) > 1 {
				memMB := vmToInt64(match[1])
				vm.Memory = utils.FormatBytes(memMB * 1024 * 1024)
			}
		}

		// Extract disk usage
		if vm.FilePath != "" {
			match := regexp.MustCompile(`\[([^\]]+)\]\s+(.+)`).FindStringSubmatch(vm.FilePath)
			if len(match) >= 3 {
				datastore := match[1]
				vmPath := match[2]

				if idx := strings.LastIndex(vmPath, "/"); idx >= 0 {
					vmPath = vmPath[:idx]
				}

				fullPath := fmt.Sprintf("/vmfs/volumes/%s/%s", datastore, vmPath)
				duCmd := fmt.Sprintf("du -hs %s 2>/dev/null | awk '{print $1}'", fullPath)

				duOutput, err := manager.RunCommand(duCmd)
				if err == nil && strings.TrimSpace(duOutput) != "" {
					vm.UsedSpace = strings.TrimSpace(duOutput)
				}
			}
		}

		vms = append(vms, *vm)
	}

	return utils.FormatAsJSON(vms)
}

func extractVMField(fields []string, index int, defaultValue string) string {
	if index < len(fields) {
		return fields[index]
	}
	return defaultValue
}

func vmToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func init() {
	config.RegisterFunc("list-vms", listVMs)
}
