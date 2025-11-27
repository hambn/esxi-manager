package virtualmachines

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// VMInfo represents a virtual machine with detailed info
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

// ListVMsCommand lists all virtual machines on the ESXi host
type ListVMsCommand struct {
	host *config.ESXiHost
}

func (c *ListVMsCommand) Validate() error {
	return nil
}

func (c *ListVMsCommand) Execute() error {
	output, err := c.runCommand("vim-cmd vmsvc/getallvms")
	if err != nil {
		return fmt.Errorf("failed to list VMs: %w", err)
	}

	vms := c.parseVMs(output)
	c.displayVMs(vms)
	return nil
}

// runCommand executes a command on the ESXi host and returns output
func (c *ListVMsCommand) runCommand(cmd string) (string, error) {
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return "", err
	}
	defer manager.Close()

	if err := manager.Connect(); err != nil {
		return "", err
	}

	client, err := manager.GetClient()
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run(cmd); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// parseVMs extracts VM information from vim-cmd output and fetches detailed info
func (c *ListVMsCommand) parseVMs(output string) []VMInfo {
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

		// Fetch additional details
		c.enrichVM(vm)
		vms = append(vms, *vm)
	}

	return vms
}

// parseVMLine parses a single VM line from vim-cmd output
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

	// Extract file path from [datastore] path pattern
	if match := regexp.MustCompile(`\[([^\]]+)\]\s+(.+?)(?:\s+\w+Guest|\s*$)`).FindStringSubmatch(line); len(match) > 2 {
		vm.FilePath = fmt.Sprintf("[%s] %s", match[1], match[2])
	}

	return vm
}

// enrichVM fetches additional VM details (status, CPU, memory, disk usage)
func (c *ListVMsCommand) enrichVM(vm *VMInfo) {
	summaryOutput, err := c.runCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vm.ID))
	if err != nil {
		return
	}

	c.extractVMStatus(summaryOutput, vm)
	c.extractCPUAndMemory(summaryOutput, vm)
	c.extractDiskUsage(vm)
}

// extractVMStatus parses power state from vm summary
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

// extractCPUAndMemory parses CPU and memory info from vm summary
func (c *ListVMsCommand) extractCPUAndMemory(summary string, vm *VMInfo) {
	// Extract CPU count
	if match := regexp.MustCompile(`numCpu\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		vm.CPU = match[1]
	}

	// Extract memory in MB
	if match := regexp.MustCompile(`memorySizeMB\s*=\s*(\d+)`).FindStringSubmatch(summary); len(match) > 1 {
		memMB := toInt64(match[1])
		vm.Memory = utils.FormatBytes(memMB * 1024 * 1024)
	}
}

// extractDiskUsage calculates VM disk usage from file path
func (c *ListVMsCommand) extractDiskUsage(vm *VMInfo) {
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
	output, err := c.runCommand(fmt.Sprintf("du -hs %s 2>/dev/null | awk '{print $1}'", fullPath))
	if err == nil && strings.TrimSpace(output) != "" {
		vm.UsedSpace = strings.TrimSpace(output)
	}
}

// displayVMs prints VM information in table format
func (c *ListVMsCommand) displayVMs(vms []VMInfo) {
	if len(vms) == 0 {
		fmt.Println("No virtual machines found")
		return
	}

	var buf bytes.Buffer
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

// Helper functions

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

// init registers the list-vms command
func init() {
	command.Register("list-vms", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListVMsCommand{host: host}
	})
}
