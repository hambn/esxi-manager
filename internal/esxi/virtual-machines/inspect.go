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

	// VMX File Configuration
	VMXConfig      map[string]string
	VMXPath        string

	// VMDK File Configuration
	VMDKConfigs    []VMDKInfo
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

// VMDKInfo holds VMDK descriptor file information
type VMDKInfo struct {
	Filename       string            // e.g., VM-DS-Oracle.vmdk
	Version        string            // Descriptor file version
	Encoding       string            // File encoding (UTF-8, etc)
	CID            string            // Content ID
	ParentCID      string            // Parent CID
	CreateType     string            // vmfs, vmfsSparse, etc
	Extents        []ExtentInfo      // Disk extents
	DDBParameters  map[string]string // Database parameters
	Capacity       int64             // Total capacity in sectors
	AdapterType    string            // lsilogic, ide, buslogic, etc
	Geometry       GeometryInfo      // Disk geometry
	ThinProvisioned bool
	UUID           string
	VirtualHWVer   string
	LongContentID  string
}

// ExtentInfo holds extent description information
type ExtentInfo struct {
	Access    string // RW, RDONLY
	Sectors   int64
	Type      string // FLAT, VMFS, SPARSE, ZERO
	Filename  string
}

// GeometryInfo holds disk geometry information
type GeometryInfo struct {
	Cylinders int
	Heads     int
	Sectors   int
}

// gatherVMInfo collects comprehensive VM information
func (i *InspectVM) gatherVMInfo(mgr *utils.SSHManager, vmID string) (*InspectVMInfo, error) {
	info := &InspectVMInfo{ID: vmID, VMXConfig: make(map[string]string)}

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

	// Parse VMX file for detailed configuration
	if err := i.parseVMXFile(mgr, info); err != nil {
		common.Debug("parse VMX file", "error", err.Error())
	}

	// Parse VMDK files for disk information
	if err := i.parseVMDKFiles(mgr, info); err != nil {
		common.Debug("parse VMDK files", "error", err.Error())
	}

	return info, nil
}

// parseVMXFile reads and parses the .vmx configuration file
func (i *InspectVM) parseVMXFile(mgr *utils.SSHManager, info *InspectVMInfo) error {
	if info.ConfigPath == "" {
		return fmt.Errorf("no config path available")
	}

	// Extract directory from config path
	parts := strings.Split(info.ConfigPath, "/")
	vmxPath := strings.Join(parts[:len(parts)-1], "/") + "/" + info.Name + ".vmx"

	output, err := mgr.RunCommand(fmt.Sprintf("cat '%s' 2>/dev/null", vmxPath))
	if err != nil {
		return common.WrapError(err, "failed to read VMX file")
	}

	info.VMXPath = vmxPath
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(strings.Trim(parts[1], "\"' "))
			info.VMXConfig[key] = val
		}
	}

	return nil
}

// parseVMDKFiles reads and parses .vmdk descriptor files
func (i *InspectVM) parseVMDKFiles(mgr *utils.SSHManager, info *InspectVMInfo) error {
	if info.ConfigPath == "" {
		return fmt.Errorf("no config path available")
	}

	// Extract directory from config path
	parts := strings.Split(info.ConfigPath, "/")
	vmxDir := strings.Join(parts[:len(parts)-1], "/")

	// List VMDK files in VM directory
	output, err := mgr.RunCommand(fmt.Sprintf("ls '%s'/*.vmdk 2>/dev/null | head -20", vmxDir))
	if err != nil || output == "" {
		return fmt.Errorf("no VMDK files found")
	}

	vmxFiles := strings.Split(strings.TrimSpace(output), "\n")
	for _, vmxFile := range vmxFiles {
		vmxFile = strings.TrimSpace(vmxFile)
		if vmxFile == "" {
			continue
		}

		vmdk, err := i.parseVMDKFile(mgr, vmxFile)
		if err != nil {
			common.Debug("parse VMDK file", "file", vmxFile, "error", err.Error())
			continue
		}
		info.VMDKConfigs = append(info.VMDKConfigs, vmdk)
	}

	return nil
}

// parseVMDKFile parses a single VMDK descriptor file
func (i *InspectVM) parseVMDKFile(mgr *utils.SSHManager, vmxPath string) (VMDKInfo, error) {
	vmdk := VMDKInfo{
		Filename:      vmxPath,
		DDBParameters: make(map[string]string),
	}

	output, err := mgr.RunCommand(fmt.Sprintf("cat '%s' 2>/dev/null", vmxPath))
	if err != nil {
		return vmdk, err
	}

	section := ""
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "Extent") {
				section = "extent"
			} else if strings.Contains(line, "DDB") {
				section = "ddb"
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(strings.Trim(parts[1], "\"' "))

		// Parse extent descriptions
		if section == "extent" && strings.Contains(key, "RW") || strings.Contains(key, "RDONLY") {
			extent := i.parseExtentLine(line)
			vmdk.Extents = append(vmdk.Extents, extent)
			continue
		}

		// Parse basic properties
		switch key {
		case "version":
			vmdk.Version = val
		case "encoding":
			vmdk.Encoding = val
		case "CID":
			vmdk.CID = val
		case "parentCID":
			vmdk.ParentCID = val
		case "createType":
			vmdk.CreateType = val
		}

		// Parse DDB parameters
		if strings.HasPrefix(key, "ddb.") {
			ddbKey := strings.TrimPrefix(key, "ddb.")
			vmdk.DDBParameters[ddbKey] = val

			// Extract specific DDB values
			switch ddbKey {
			case "adapterType":
				vmdk.AdapterType = val
			case "uuid":
				vmdk.UUID = val
			case "virtualHWVersion":
				vmdk.VirtualHWVer = val
			case "longContentID":
				vmdk.LongContentID = val
			case "thinProvisioned":
				vmdk.ThinProvisioned = val == "1"
			case "geometry.cylinders":
				fmt.Sscanf(val, "%d", &vmdk.Geometry.Cylinders)
			case "geometry.heads":
				fmt.Sscanf(val, "%d", &vmdk.Geometry.Heads)
			case "geometry.sectors":
				fmt.Sscanf(val, "%d", &vmdk.Geometry.Sectors)
			}
		}
	}

	return vmdk, nil
}

// parseExtentLine parses an extent description line
func (i *InspectVM) parseExtentLine(line string) ExtentInfo {
	extent := ExtentInfo{}
	fields := strings.Fields(line)

	if len(fields) >= 4 {
		extent.Access = fields[0]
		fmt.Sscanf(fields[1], "%d", &extent.Sectors)
		extent.Type = fields[2]
		extent.Filename = strings.Trim(fields[3], "\"'")
	}

	return extent
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

	// Display VMX Configuration
	if len(info.VMXConfig) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("[VMX CONFIGURATION FILE]")
		fmt.Printf("  File Path: %s\n\n", info.VMXPath)
		fmt.Println("  Configuration Parameters:")
		fmt.Println("  " + strings.Repeat("-", 76))

		// Group and display VMX settings
		categories := map[string][]string{
			"Hardware": {"numvcpus", "memSize", "virtualHW.version", "cpuid.coresPerSocket"},
			"Firmware": {"firmware", "uefi.secureBoot.enabled", "bios.bootOrder", "boot.bootDelay"},
			"Displays": {"svga.present", "svga.autodetect", "svga.vramSize", "RemoteDisplay.maxConnections"},
			"Disks": {"scsi0.present", "sata0.present", "scsi0:0.fileName", "sata0:0.fileName"},
			"Networks": {"ethernet0.present", "ethernet0.virtualDev", "ethernet0.networkName", "ethernet0.generatedAddress"},
			"USB": {"usb.present", "ehci.present"},
			"VMTools": {"tools.upgrade.policy", "tools.syncTime", "toolScripts.afterPowerOn"},
			"Power": {"powerType.powerOff", "powerType.suspend", "powerType.reset"},
			"Scheduling": {"sched.cpu.units", "sched.cpu.affinity", "sched.cpu.latencySensitivity"},
		}

		displayedKeys := make(map[string]bool)

		for category, keys := range categories {
			hasKeys := false
			for _, key := range keys {
				if val, exists := info.VMXConfig[key]; exists {
					if !hasKeys {
						fmt.Printf("\n  %s:\n", category)
						hasKeys = true
					}
					fmt.Printf("    %-45s = %s\n", key, val)
					displayedKeys[key] = true
				}
			}
		}

		// Display remaining uncategorized settings
		remaining := []string{}
		for key := range info.VMXConfig {
			if !displayedKeys[key] {
				remaining = append(remaining, key)
			}
		}
		if len(remaining) > 0 {
			fmt.Printf("\n  Other Settings:\n")
			for _, key := range remaining {
				fmt.Printf("    %-45s = %s\n", key, info.VMXConfig[key])
			}
		}
	}

	// Display VMDK Information
	if len(info.VMDKConfigs) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("[VIRTUAL DISK INFORMATION]")

		for idx, vmdk := range info.VMDKConfigs {
			fmt.Printf("\n  Disk %d: %s\n", idx+1, vmdk.Filename)
			fmt.Println("  " + strings.Repeat("-", 76))

			fmt.Printf("    Descriptor Information:\n")
			fmt.Printf("      Version:               %s\n", vmdk.Version)
			fmt.Printf("      Encoding:              %s\n", vmdk.Encoding)
			fmt.Printf("      CID:                   %s\n", vmdk.CID)
			fmt.Printf("      Parent CID:            %s\n", vmdk.ParentCID)
			fmt.Printf("      Create Type:           %s\n", vmdk.CreateType)

			if vmdk.UUID != "" {
				fmt.Printf("      UUID:                  %s\n", vmdk.UUID)
			}
			if vmdk.LongContentID != "" {
				fmt.Printf("      Long Content ID:       %s\n", vmdk.LongContentID)
			}

			fmt.Printf("\n    Storage Configuration:\n")
			fmt.Printf("      Adapter Type:          %s\n", vmdk.AdapterType)
			fmt.Printf("      Thin Provisioned:      %v\n", vmdk.ThinProvisioned)
			fmt.Printf("      Virtual HW Version:    %s\n", vmdk.VirtualHWVer)

			if vmdk.Geometry.Cylinders > 0 {
				fmt.Printf("\n    Disk Geometry:\n")
				fmt.Printf("      Cylinders:             %d\n", vmdk.Geometry.Cylinders)
				fmt.Printf("      Heads:                 %d\n", vmdk.Geometry.Heads)
				fmt.Printf("      Sectors:               %d\n", vmdk.Geometry.Sectors)
			}

			if len(vmdk.Extents) > 0 {
				fmt.Printf("\n    Extents:\n")
				for extIdx, extent := range vmdk.Extents {
					fmt.Printf("      Extent %d:\n", extIdx+1)
					fmt.Printf("        Access:            %s\n", extent.Access)
					fmt.Printf("        Sectors:           %d\n", extent.Sectors)
					fmt.Printf("        Type:              %s\n", extent.Type)
					fmt.Printf("        Filename:          %s\n", extent.Filename)
				}
			}

			if len(vmdk.DDBParameters) > 0 {
				fmt.Printf("\n    Database Parameters (DDB):\n")
				for key, val := range vmdk.DDBParameters {
					fmt.Printf("      %-40s = %s\n", key, val)
				}
			}
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
