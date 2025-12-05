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

type ListVMsCommand struct {
	params *config.Params
}

func (c *ListVMsCommand) Validate() error {
	return nil
}

func (c *ListVMsCommand) Execute() (string, error) {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("vim-cmd vmsvc/getallvms")
	if err != nil {
		return "", fmt.Errorf("failed to list VMs: %w", err)
	}

	vms := c.parseVMs(output, manager)
	return c.formatVMs(vms)
}

func (c *ListVMsCommand) parseVMs(output string, manager *utils.SSHManager) []VMListInfo {
	var vms []VMListInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Vmid") {
			continue
		}

		vm := c.parseVMLine(line)
		if vm == nil {
			continue
		}

		c.enrichVM(vm, manager)
		vms = append(vms, *vm)
	}

	return vms
}

func (c *ListVMsCommand) parseVMLine(line string) *VMListInfo {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
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

	return vm
}

func (c *ListVMsCommand) enrichVM(vm *VMListInfo, manager *utils.SSHManager) {
	summaryCmd := fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vm.ID)
	summaryOutput, err := manager.RunCommand(summaryCmd)
	if err != nil {
		return
	}

	c.extractVMStatus(summaryOutput, vm)
	c.extractCPUAndMemory(summaryOutput, vm)
	c.extractDiskUsage(vm, manager)
}

func (c *ListVMsCommand) extractVMStatus(summary string, vm *VMListInfo) {
	switch {
	case strings.Contains(summary, `powerState = "poweredOn"`):
		vm.Status = "on"
	case strings.Contains(summary, `powerState = "poweredOff"`):
		vm.Status = "off"
	case strings.Contains(summary, `powerState = "suspended"`):
		vm.Status = "suspended"
	}
}

func (c *ListVMsCommand) extractCPUAndMemory(summary string, vm *VMListInfo) {
	if match := regexp.MustCompile(`numCpu\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		vm.CPU = match[1]
	}

	if match := regexp.MustCompile(`memorySizeMB\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		memMB := vmToInt64(match[1])
		vm.Memory = utils.FormatBytes(memMB * 1024 * 1024)
	}
}

func (c *ListVMsCommand) extractDiskUsage(vm *VMListInfo, manager *utils.SSHManager) {
	if vm.FilePath == "" {
		return
	}

	match := regexp.MustCompile(`\[([^\]]+)\]\s+(.+)`).FindStringSubmatch(vm.FilePath)
	if len(match) < 3 {
		return
	}

	datastore := match[1]
	vmPath := match[2]

	// Get directory path (without .vmx file)
	if idx := strings.LastIndex(vmPath, "/"); idx >= 0 {
		vmPath = vmPath[:idx]
	}

	fullPath := fmt.Sprintf("/vmfs/volumes/%s/%s", datastore, vmPath)
	duCmd := fmt.Sprintf("du -hs %s 2>/dev/null | awk '{print $1}'", fullPath)

	output, err := manager.RunCommand(duCmd)
	if err == nil && strings.TrimSpace(output) != "" {
		vm.UsedSpace = strings.TrimSpace(output)
	}
}

func (c *ListVMsCommand) formatVMs(vms []VMListInfo) (string, error) {
	if len(vms) == 0 {
		// Return empty JSON array
		return utils.FormatAsJSON([]VMListInfo{})
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
	config.Register("list-vms", func(params *config.Params) config.CommandInterface {
		return &ListVMsCommand{params: params}
	})
}
