package virtualmachines

import (
	"fmt"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/common"
	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// InspectVM represents the inspect command for detailed VM information
type InspectVM struct {
	params *config.Params
}

// NewInspectVM creates a new inspect command instance
func NewInspectVM(params *config.Params) *InspectVM {
	return &InspectVM{params: params}
}

// Validate checks that required parameters are present
func (i *InspectVM) Validate() error {
	// Either ID or Name must be provided
	if i.params.VMInspectID == "" && i.params.VMInspectName == "" {
		return common.NewValidationError("vm-inspect-id or vm-inspect-name", "at least one must be provided")
	}
	return nil
}

// Execute gathers and displays comprehensive VM information
func (i *InspectVM) Execute() error {
	mgr, err := utils.NewSSHManager(i.params)
	if err != nil {
		return common.NewConnectionError(i.params.ESXiHostURI, "failed to create SSH manager", err)
	}
	defer mgr.Close()

	if err := mgr.Connect(); err != nil {
		return common.NewConnectionError(i.params.ESXiHostURI, "failed to connect", err)
	}

	// Resolve VM ID if only name is provided
	vmID := i.params.VMInspectID
	if vmID == "" {
		id, err := i.resolveVMIDFromName(mgr, i.params.VMInspectName)
		if err != nil {
			return err
		}
		vmID = id
	}

	// Gather all VM information
	info, err := i.gatherVMInfo(mgr, vmID)
	if err != nil {
		return err
	}

	// Display the information
	i.displayVMInfo(info)
	return nil
}

// resolveVMIDFromName finds VM ID by name
func (i *InspectVM) resolveVMIDFromName(mgr *utils.SSHManager, vmName string) (string, error) {
	output, err := mgr.RunCommand("vim-cmd vmsvc/getallvms | grep '" + vmName + "'")
	if err != nil {
		return "", common.WrapError(err, "failed to query VMs")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("VM '%s' not found", vmName)
	}

	parts := strings.Fields(lines[0])
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid VM list format")
	}

	return parts[0], nil
}

// InspectVMInfo holds all gathered VM inspection information
type InspectVMInfo struct {
	ID             string
	Name           string
	State          string
	Annotation     string
	ConfigPath     string
	Uuid           string
	Version        string
	CPUs           int
	Memory         int // in MB
	NICs           int
	DiskCount      int
	MaxCPUs        int
	MaxMemory      int
	BootDelay      int
	Firmware       string
	GuestOS        string
	ToolsRunning   string
	ToolsVersion   string
	Networks       []NetworkInfo
	StorageDevices []DiskInfo
	CPUInfo        CPUInfo
}

// NetworkInfo holds NIC information
type NetworkInfo struct {
	Index      int
	Name       string
	MacAddress string
	Network    string
	Connected  bool
}

// DiskInfo holds disk information
type DiskInfo struct {
	Index      int
	Path       string
	Size       int64 // in bytes
	Datastore  string
	Controller string
}

// CPUInfo holds CPU configuration
type CPUInfo struct {
	Cores   int
	Threads int
	HZ      string
}

// gatherVMInfo collects comprehensive VM information
func (i *InspectVM) gatherVMInfo(mgr *utils.SSHManager, vmID string) (*InspectVMInfo, error) {
	info := &InspectVMInfo{ID: vmID}

	// Get basic info
	if err := i.getBasicInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Get hardware info
	if err := i.getHardwareInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Get guest info
	if err := i.getGuestInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Get runtime info
	if err := i.getRuntimeInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Get network info
	if err := i.getNetworkInfo(mgr, vmID, info); err != nil {
		// Don't fail if network info unavailable
		common.Debug("network info", "error", err.Error())
	}

	// Get storage info
	if err := i.getStorageInfo(mgr, vmID, info); err != nil {
		// Don't fail if storage info unavailable
		common.Debug("storage info", "error", err.Error())
	}

	return info, nil
}

// getBasicInfo retrieves basic VM information
func (i *InspectVM) getBasicInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s | grep -E '(name|state|config.annotation|config.uuid)'", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get basic VM info")
	}

	parseConfigValue(output, "name =", &info.Name)
	parseConfigValue(output, "state =", &info.State)
	parseConfigValue(output, "config.annotation =", &info.Annotation)
	parseConfigValue(output, "config.uuid =", &info.Uuid)

	// Get config path
	pathOutput, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep 'configFile'", vmID))
	if err == nil {
		parseConfigValue(pathOutput, "configFile =", &info.ConfigPath)
	}

	return nil
}

// getHardwareInfo retrieves hardware configuration
func (i *InspectVM) getHardwareInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep -E '(memoryMB|numCPU|version|firmware|bootDelay)'", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get hardware info")
	}

	parseIntValue(output, "memoryMB =", &info.Memory)
	parseIntValue(output, "numCPU =", &info.CPUs)
	parseConfigValue(output, "version =", &info.Version)
	parseConfigValue(output, "firmware =", &info.Firmware)
	parseIntValue(output, "bootDelay =", &info.BootDelay)

	// Get max resources
	maxOutput, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep -E '(maxCpus|maxMemory)'", vmID))
	if err == nil {
		parseIntValue(maxOutput, "maxCpus =", &info.MaxCPUs)
		parseIntValue(maxOutput, "maxMemory =", &info.MaxMemory)
	}

	return nil
}

// getGuestInfo retrieves guest OS information
func (i *InspectVM) getGuestInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s | grep -E '(guestFullName|toolsRunningStatus|toolsVersion)'", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get guest info")
	}

	parseConfigValue(output, "guestFullName =", &info.GuestOS)
	parseConfigValue(output, "toolsRunningStatus =", &info.ToolsRunning)
	parseConfigValue(output, "toolsVersion =", &info.ToolsVersion)

	return nil
}

// getRuntimeInfo retrieves runtime information
func (i *InspectVM) getRuntimeInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s | grep -E '(runtime|numEthernetCards|numVirtualDisks)'", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get runtime info")
	}

	parseIntValue(output, "numEthernetCards =", &info.NICs)
	parseIntValue(output, "numVirtualDisks =", &info.DiskCount)

	return nil
}

// getNetworkInfo retrieves network configuration
func (i *InspectVM) getNetworkInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	// This is a simplified version - real implementation would parse detailed NIC info
	if info.NICs > 0 {
		for idx := 0; idx < info.NICs; idx++ {
			network := NetworkInfo{
				Index: idx,
				Name:  fmt.Sprintf("Network adapter %d", idx+1),
			}
			info.Networks = append(info.Networks, network)
		}
	}
	return nil
}

// getStorageInfo retrieves storage information
func (i *InspectVM) getStorageInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	// This is a simplified version - real implementation would parse detailed disk info
	if info.DiskCount > 0 {
		for idx := 0; idx < info.DiskCount; idx++ {
			disk := DiskInfo{
				Index: idx,
				Path:  fmt.Sprintf("Hard disk %d", idx+1),
			}
			info.StorageDevices = append(info.StorageDevices, disk)
		}
	}
	return nil
}

// displayVMInfo prints formatted VM information
func (i *InspectVM) displayVMInfo(info *InspectVMInfo) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("VM INSPECTION REPORT - %s (ID: %s)\n", info.Name, info.ID)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[BASIC INFORMATION]")
	fmt.Printf("  Name:                    %s\n", info.Name)
	fmt.Printf("  VM ID:                   %s\n", info.ID)
	fmt.Printf("  State:                   %s\n", info.State)
	fmt.Printf("  UUID:                    %s\n", info.Uuid)
	fmt.Printf("  Config Path:             %s\n", info.ConfigPath)
	fmt.Printf("  Annotation:              %s\n", info.Annotation)

	fmt.Println("\n[HARDWARE CONFIGURATION]")
	fmt.Printf("  CPUs:                    %d cores\n", info.CPUs)
	fmt.Printf("  Memory:                  %d MB (%.2f GB)\n", info.Memory, float64(info.Memory)/1024)
	fmt.Printf("  Version:                 %s\n", info.Version)
	fmt.Printf("  Firmware:                %s\n", info.Firmware)
	fmt.Printf("  Boot Delay:              %d ms\n", info.BootDelay)
	if info.MaxCPUs > 0 {
		fmt.Printf("  Max CPUs:                %d\n", info.MaxCPUs)
	}
	if info.MaxMemory > 0 {
		fmt.Printf("  Max Memory:              %d MB (%.2f GB)\n", info.MaxMemory, float64(info.MaxMemory)/1024)
	}

	fmt.Println("\n[GUEST INFORMATION]")
	fmt.Printf("  Guest OS:                %s\n", info.GuestOS)
	fmt.Printf("  VMware Tools Status:     %s\n", info.ToolsRunning)
	fmt.Printf("  VMware Tools Version:    %s\n", info.ToolsVersion)

	fmt.Println("\n[NETWORK & STORAGE]")
	fmt.Printf("  Network Adapters:        %d\n", info.NICs)
	fmt.Printf("  Virtual Disks:           %d\n", info.DiskCount)

	if len(info.Networks) > 0 {
		fmt.Println("\n[NETWORK ADAPTERS]")
		for _, net := range info.Networks {
			fmt.Printf("  %s:\n", net.Name)
			fmt.Printf("    MAC Address:           %s\n", net.MacAddress)
			fmt.Printf("    Network:               %s\n", net.Network)
			fmt.Printf("    Connected:             %v\n", net.Connected)
		}
	}

	if len(info.StorageDevices) > 0 {
		fmt.Println("\n[STORAGE DEVICES]")
		for _, disk := range info.StorageDevices {
			fmt.Printf("  %s:\n", disk.Path)
			fmt.Printf("    Size:                  %d bytes\n", disk.Size)
			fmt.Printf("    Datastore:             %s\n", disk.Datastore)
			fmt.Printf("    Controller:            %s\n", disk.Controller)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// Helper functions to parse vim-cmd output

func parseConfigValue(output, key string, target *string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, key) {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				*target = strings.TrimSpace(strings.Trim(parts[1], "\",'"))
			}
			break
		}
	}
}

func parseIntValue(output, key string, target *int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, key) {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				fmt.Sscanf(value, "%d", target)
			}
			break
		}
	}
}

// Register registers the inspect command
func init() {
	config.Register("vm-inspect", func(params *config.Params) config.CommandInterface {
		return NewInspectVM(params)
	})
}
