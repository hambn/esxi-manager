package virtualmachines

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

type VMInfo struct {
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

func (c *ListVMsCommand) Execute() error {
	manager, err := utils.NewSSHManager(c.params)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	output, err := manager.RunCommand("vim-cmd vmsvc/getallvms")
	if err != nil {
		return fmt.Errorf("failed to list VMs: %w", err)
	}

	vms := c.parseVMs(output, manager)
	c.displayVMs(vms)
	return nil
}

func (c *ListVMsCommand) parseVMs(output string, manager *utils.SSHManager) []VMInfo {
	var vms []VMInfo
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

func (c *ListVMsCommand) parseVMLine(line string) *VMInfo {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	vm := &VMInfo{
		ID:        fields[0],
		Name:      fields[1],
		GuestOS:   extractField(fields, 4, "-"),
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

func (c *ListVMsCommand) enrichVM(vm *VMInfo, manager *utils.SSHManager) {
	summaryCmd := fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vm.ID)
	summaryOutput, err := manager.RunCommand(summaryCmd)
	if err != nil {
		return
	}

	c.extractVMStatus(summaryOutput, vm)
	c.extractCPUAndMemory(summaryOutput, vm)
	c.extractDiskUsage(vm, manager)
}

func (c *ListVMsCommand) extractVMStatus(summary string, vm *VMInfo) {
	switch {
	case strings.Contains(summary, `powerState = "poweredOn"`):
		vm.Status = "on"
	case strings.Contains(summary, `powerState = "poweredOff"`):
		vm.Status = "off"
	case strings.Contains(summary, `powerState = "suspended"`):
		vm.Status = "suspended"
	}
}

func (c *ListVMsCommand) extractCPUAndMemory(summary string, vm *VMInfo) {
	if match := regexp.MustCompile(`numCpu\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		vm.CPU = match[1]
	}

	if match := regexp.MustCompile(`memorySizeMB\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		memMB := toInt64(match[1])
		vm.Memory = utils.FormatBytes(memMB * 1024 * 1024)
	}
}

func (c *ListVMsCommand) extractDiskUsage(vm *VMInfo, manager *utils.SSHManager) {
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

func (c *ListVMsCommand) displayVMs(vms []VMInfo) {
	if len(vms) == 0 {
		fmt.Println("No virtual machines found")
		return
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tName\tGuest OS\tStatus\tCPU\tMemory\tUsed Space\n")

	for _, vm := range vms {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			vm.ID, vm.Name, vm.GuestOS, vm.Status, vm.CPU, vm.Memory, vm.UsedSpace,
		)
	}

	w.Flush()
	fmt.Print(buf.String())
}

func extractField(fields []string, index int, defaultValue string) string {
	if index < len(fields) {
		return fields[index]
	}
	return defaultValue
}

func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func init() {
	config.Register("list-vms", func(params *config.Params) config.CommandInterface {
		return &ListVMsCommand{params: params}
	})
}
