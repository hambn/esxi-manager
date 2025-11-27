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
	ID          string
	Name        string
	GuestOS     string
	Status      string
	UsedSpace   string
	CPU         string
	Memory      string
}

// ListVMsCommand lists all virtual machines on the ESXi host
type ListVMsCommand struct {
	host *config.ESXiHost
}

func (c *ListVMsCommand) Validate() error {
	return nil
}

func (c *ListVMsCommand) Execute() error {
	// Manage connection internally
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	// Connect to host
	if err := manager.Connect(); err != nil {
		return fmt.Errorf("failed to connect to ESXi host: %w", err)
	}

	// Get client
	client, err := manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %w", err)
	}

	// List VMs using vim-cmd command
	var stdout bytes.Buffer
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	session.Stdout = &stdout

	if err := session.Run("vim-cmd vmsvc/getallvms"); err != nil {
		session.Close()
		return fmt.Errorf("failed to execute list VMs command: %w", err)
	}
	session.Close()

	// Parse VMs and get detailed info for each
	vms, err := c.parseVMs(stdout.String(), client)
	if err != nil {
		return err
	}

	c.displayVMs(vms)
	return nil
}

func (c *ListVMsCommand) parseVMs(output string, client interface{}) ([]VMInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var vms []VMInfo

	// Skip header and empty lines
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Vmid") {
			continue
		}

		// Parse: Vmid   Name               File                  Guest OS     Version
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			vmid := fields[0]
			name := fields[1]

			// Extract Guest OS (typically at index 4)
			guestOS := "-"
			if len(fields) > 4 {
				guestOS = fields[4]
			}

			// Get detailed info for this VM
			vmInfo := VMInfo{
				ID:      vmid,
				Name:    name,
				GuestOS: guestOS,
				Status:  "unknown",
				CPU:     "-",
				Memory:  "-",
			}

			// Try to get summary info and merge with existing info
			if summaryInfo, err := c.getVMSummary(vmid, client); err == nil {
				vmInfo.Status = summaryInfo.Status
				vmInfo.CPU = summaryInfo.CPU
				vmInfo.Memory = summaryInfo.Memory
				vmInfo.UsedSpace = summaryInfo.UsedSpace
			}

			vms = append(vms, vmInfo)
		}
	}

	return vms, nil
}

func (c *ListVMsCommand) getVMSummary(vmid string, clientInterface interface{}) (VMInfo, error) {
	vmInfo := VMInfo{
		ID:     vmid,
		Status: "unknown",
		CPU:    "-",
		Memory: "-",
	}

	// Get a new SSH client from the connection manager
	// We need to recreate it since we don't have direct access to the manager
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return vmInfo, err
	}
	defer manager.Close()

	if err := manager.Connect(); err != nil {
		return vmInfo, err
	}

	client, err := manager.GetClient()
	if err != nil {
		return vmInfo, err
	}

	session, err := client.NewSession()
	if err != nil {
		return vmInfo, err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	// Get VM summary
	cmd := fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vmid)
	if err := session.Run(cmd); err != nil {
		return vmInfo, err
	}

	// Parse summary output
	summaryText := stdout.String()
	vmInfo = c.parseSummary(summaryText, vmInfo)

	return vmInfo, nil
}

func (c *ListVMsCommand) parseSummary(summary string, vmInfo VMInfo) VMInfo {
	// Extract status
	if strings.Contains(summary, "powerState = \"poweredOn\"") {
		vmInfo.Status = "on"
	} else if strings.Contains(summary, "powerState = \"poweredOff\"") {
		vmInfo.Status = "off"
	} else if strings.Contains(summary, "powerState = \"suspended\"") {
		vmInfo.Status = "suspended"
	}

	// Extract memory in MB
	memRegex := regexp.MustCompile(`memorySizeMB\s*=\s*(\d+)`)
	if matches := memRegex.FindStringSubmatch(summary); len(matches) > 1 {
		memMB := matches[1]
		memBytes := parseInt64(memMB) * 1024 * 1024
		vmInfo.Memory = utils.FormatBytes(memBytes)
	}

	// Extract CPU count
	cpuRegex := regexp.MustCompile(`numCpu\s*=\s*(\d+)`)
	if matches := cpuRegex.FindStringSubmatch(summary); len(matches) > 1 {
		vmInfo.CPU = matches[1]
	}

	// For used space, we would need to run du command - skip for now and mark as "-"
	vmInfo.UsedSpace = "-"

	return vmInfo
}

func (c *ListVMsCommand) displayVMs(vms []VMInfo) {
	if len(vms) == 0 {
		fmt.Println("No virtual machines found")
		return
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tName\tGuest OS\tStatus\tCPU\tMemory\n")

	for _, vm := range vms {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			vm.ID,
			vm.Name,
			vm.GuestOS,
			vm.Status,
			vm.CPU,
			vm.Memory,
		)
	}
	w.Flush()
	fmt.Print(buf.String())
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// init registers the list-vms command
func init() {
	command.Register("list-vms", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListVMsCommand{
			host: host,
		}
	})
}
