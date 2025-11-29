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
	BiosUuid       string
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
	CreateDate     string
	PowerState     string
	UpTime         string

	// Networks and Storage
	Networks       []NetworkInfo
	StorageDevices []DiskInfo
	Datastores     []DatastoreInfo
	Snapshots      []SnapshotInfo
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
	Name       string
	Path       string
	Size       int64 // in bytes
	Datastore  string
	Controller string
	DeviceType string
	Filename   string
}

// DatastoreInfo holds datastore information
type DatastoreInfo struct {
	Name       string
	Path       string
	Capacity   int64 // in bytes
	FreeSpace  int64 // in bytes
	UsedSpace  int64 // in bytes
	Type       string
	URL        string
}

// SnapshotInfo holds snapshot information
type SnapshotInfo struct {
	Key        string
	Name       string
	Description string
	CreateTime string
	State      string
	ParentKey  string
	ChildKeys  []string
	Quiesced   bool
	BackupMode bool
	Size       int64
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

	// Get snapshots
	if err := i.getSnapshots(mgr, vmID, info); err != nil {
		common.Debug("get snapshots", "error", err.Error())
	}

	// Get datastores
	if err := i.getDatastores(mgr, vmID, info); err != nil {
		common.Debug("get datastores", "error", err.Error())
	}

	return info, nil
}

// getSnapshots retrieves snapshot information
func (i *InspectVM) getSnapshots(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/snapshot.get %s 2>/dev/null", vmID))
	if err != nil || output == "" {
		return fmt.Errorf("no snapshots or error getting snapshots")
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "--") {
			continue
		}

		// Parse snapshot lines
		parts := strings.Split(line, "--")
		if len(parts) >= 2 {
			snapshot := SnapshotInfo{
				Name: strings.TrimSpace(parts[0]),
			}

			// Extract details from the line
			if strings.Contains(line, "state:") {
				fields := strings.Fields(line)
				for idx, field := range fields {
					if field == "state:" && idx+1 < len(fields) {
						snapshot.State = fields[idx+1]
					}
					if field == "size:" && idx+1 < len(fields) {
						fmt.Sscanf(fields[idx+1], "%d", &snapshot.Size)
					}
				}
			}

			info.Snapshots = append(info.Snapshots, snapshot)
		}
	}

	return nil
}

// getDatastores retrieves datastore information
func (i *InspectVM) getDatastores(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	// Get VM's config path to determine its datastore
	if info.ConfigPath == "" {
		return fmt.Errorf("no config path available")
	}

	// Extract datastore from config path (format: /vmfs/volumes/datastore-uuid/...)
	parts := strings.Split(info.ConfigPath, "/")
	if len(parts) > 3 {
		datastorePath := parts[3]

		output, err := mgr.RunCommand(fmt.Sprintf("ls -lh /vmfs/volumes/ 2>/dev/null | grep '%s'", datastorePath))
		if err == nil && output != "" {
			datastore := DatastoreInfo{
				Path: "/vmfs/volumes/" + datastorePath,
			}

			// Try to get datastore info
			infoOutput, err := mgr.RunCommand(fmt.Sprintf("df -h /vmfs/volumes/%s 2>/dev/null | tail -1", datastorePath))
			if err == nil && infoOutput != "" {
				fields := strings.Fields(infoOutput)
				if len(fields) >= 5 {
					datastore.Name = datastorePath
					// Parse capacity and free space
					parseSizeValue(fields[1], &datastore.Capacity)
					parseSizeValue(fields[3], &datastore.FreeSpace)
					if datastore.Capacity > 0 && datastore.FreeSpace > 0 {
						datastore.UsedSpace = datastore.Capacity - datastore.FreeSpace
					}
					datastore.Type = "VMFS"
					info.Datastores = append(info.Datastores, datastore)
				}
			}
		}
	}

	return nil
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
	// Get full summary output - no grep, we'll parse it all
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get basic VM info")
	}

	// Parse all values from summary with multiple fallback patterns
	parseConfigValueFlexible(output, []string{"name =", "name="}, &info.Name)
	parseConfigValueFlexible(output, []string{"state =", "state=", "config.name.state =", "config.name.state=", "cpuHotAddEnabled =", "memoryHotAddEnabled ="}, &info.State)
	parseConfigValueFlexible(output, []string{"config.annotation =", "config.annotation=", "annotation =", "annotation="}, &info.Annotation)
	parseConfigValueFlexible(output, []string{"config.uuid =", "config.uuid=", "uuid =", "uuid="}, &info.Uuid)
	parseConfigValueFlexible(output, []string{"uuid.bios =", "uuid.bios=", "bios.uuid =", "config.uuid.bios =", "uuid ="}, &info.BiosUuid)
	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname =", "guestOS =", "config.guestFullName ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"toolsRunningStatus =", "toolsRunningStatus=", "tools.runningStatus =", "guest.toolsRunningStatus ="}, &info.ToolsRunning)
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "tools.version =", "guest.toolsVersion =", "guestToolsVersion ="}, &info.ToolsVersion)
	parseConfigValueFlexible(output, []string{"powerState =", "powerState=", "runtime.powerState =", "runtime.powerState ="}, &info.PowerState)

	// Get config file path - try multiple approaches
	parseConfigValueFlexible(output, []string{"config.files.vmPathName =", "config.files.vmPathName=", "vmPathName =", "vmPathName="}, &info.ConfigPath)

	if info.ConfigPath == "" {
		configOutput, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s 2>/dev/null", vmID))
		if err == nil && configOutput != "" {
			parseConfigValueFlexible(configOutput, []string{"configFile =", "configFile=", "config.files.vmPathName =", "vmPathName =", "files.vmPathName =", "path ="}, &info.ConfigPath)
		}
	}

	// Last resort: try to find .vmx file in /vmfs/volumes
	if info.ConfigPath == "" && info.Name != "" {
		searchOutput, err := mgr.RunCommand(fmt.Sprintf("find /vmfs/volumes -name '%s.vmx' 2>/dev/null | head -1", info.Name))
		if err == nil && strings.TrimSpace(searchOutput) != "" {
			info.ConfigPath = strings.TrimSpace(searchOutput)
		}
	}

	// If we still don't have UUID, try alternate patterns
	if info.Uuid == "" {
		parseConfigValueFlexible(output, []string{"uuid =", "uuid=", "config.uuid ="}, &info.Uuid)
	}

	return nil
}

// getHardwareInfo retrieves hardware configuration
func (i *InspectVM) getHardwareInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s 2>/dev/null", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get hardware info")
	}

	parseIntValueFlexible(output, []string{"memoryMB =", "memoryMB=", "config.hardware.memoryMB ="}, &info.Memory)
	parseIntValueFlexible(output, []string{"numCPU =", "numCPU=", "config.hardware.numCPU ="}, &info.CPUs)
	parseConfigValueFlexible(output, []string{"version =", "version=", "config.version ="}, &info.Version)
	parseConfigValueFlexible(output, []string{"firmware =", "firmware=", "config.hardware.firmware ="}, &info.Firmware)
	parseIntValueFlexible(output, []string{"bootDelay =", "bootDelay=", "config.hardware.bootDelay ="}, &info.BootDelay)

	// Get max resources
	parseIntValueFlexible(output, []string{"maxCpus =", "maxCpus=", "config.hardware.maxCpus ="}, &info.MaxCPUs)
	parseIntValueFlexible(output, []string{"maxMemory =", "maxMemory=", "config.hardware.maxMemory ="}, &info.MaxMemory)

	return nil
}

// getGuestInfo retrieves guest OS information
func (i *InspectVM) getGuestInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get guest info")
	}

	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname =", "config.guestFullName ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"toolsRunningStatus =", "toolsRunningStatus=", "tools.runningStatus =", "guest.toolsRunningStatus ="}, &info.ToolsRunning)
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "tools.version =", "guest.toolsVersion ="}, &info.ToolsVersion)

	return nil
}

// getRuntimeInfo retrieves runtime information
func (i *InspectVM) getRuntimeInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get runtime info")
	}

	parseIntValueFlexible(output, []string{"numEthernetCards =", "numEthernetCards=", "config.hardware.numEthernetCards ="}, &info.NICs)
	parseIntValueFlexible(output, []string{"numVirtualDisks =", "numVirtualDisks=", "config.hardware.numVirtualDisks ="}, &info.DiskCount)

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
	fmt.Printf("  Name:                    %s\n", formatFieldValue(info.Name))
	fmt.Printf("  VM ID:                   %s\n", info.ID)
	if info.State != "" {
		fmt.Printf("  State:                   %s\n", formatFieldValue(info.State))
	}
	fmt.Printf("  Power State:             %s\n", formatFieldValue(info.PowerState))
	fmt.Printf("  UUID:                    %s\n", formatFieldValue(info.Uuid))
	if info.BiosUuid != "" {
		fmt.Printf("  BIOS UUID:               %s\n", formatFieldValue(info.BiosUuid))
	}
	if info.ConfigPath != "" {
		fmt.Printf("  Config Path:             %s\n", formatFieldValue(info.ConfigPath))
	}
	if info.Annotation != "" {
		fmt.Printf("  Annotation:              %s\n", formatFieldValue(info.Annotation))
	}
	if info.CreateDate != "" {
		fmt.Printf("  Create Date:             %s\n", info.CreateDate)
	}
	if info.UpTime != "" {
		fmt.Printf("  Up Time:                 %s\n", info.UpTime)
	}

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
	if info.GuestOS != "" && !isPlaceholderValue(info.GuestOS) {
		fmt.Printf("  Guest OS:                %s\n", formatFieldValue(info.GuestOS))
	} else {
		fmt.Printf("  Guest OS:                N/A\n")
	}
	fmt.Printf("  VMware Tools Status:     %s\n", formatFieldValue(info.ToolsRunning))
	if info.ToolsVersion != "" {
		fmt.Printf("  VMware Tools Version:    %s\n", formatFieldValue(info.ToolsVersion))
	}

	fmt.Println("\n[NETWORK & STORAGE]")
	fmt.Printf("  Network Adapters:        %d\n", info.NICs)
	fmt.Printf("  Virtual Disks:           %d\n", info.DiskCount)

	if len(info.Networks) > 0 {
		fmt.Println("\n[NETWORK ADAPTERS]")
		for _, net := range info.Networks {
			fmt.Printf("  %s:\n", net.Name)
			if net.MacAddress != "" {
				fmt.Printf("    MAC Address:           %s\n", formatFieldValue(net.MacAddress))
			}
			if net.Network != "" {
				fmt.Printf("    Network:               %s\n", formatFieldValue(net.Network))
			}
			if net.Connected {
				fmt.Printf("    Connected:             %v\n", net.Connected)
			}
		}
	}

	if len(info.StorageDevices) > 0 {
		fmt.Println("\n[STORAGE DEVICES]")
		for _, disk := range info.StorageDevices {
			fmt.Printf("  %s:\n", disk.Path)
			if disk.Size > 0 {
				fmt.Printf("    Size:                  %d bytes (%.2f GB)\n", disk.Size, float64(disk.Size)/(1024*1024*1024))
			}
			if disk.Datastore != "" {
				fmt.Printf("    Datastore:             %s\n", formatFieldValue(disk.Datastore))
			}
			if disk.Controller != "" {
				fmt.Printf("    Controller:            %s\n", formatFieldValue(disk.Controller))
			}
		}
	}

	// Display Datastores
	if len(info.Datastores) > 0 {
		fmt.Println("\n[DATASTORES]")
		for idx, ds := range info.Datastores {
			fmt.Printf("  Datastore %d: %s\n", idx+1, ds.Name)
			fmt.Printf("    Path:                  %s\n", ds.Path)
			fmt.Printf("    Type:                  %s\n", ds.Type)
			fmt.Printf("    Capacity:              %d bytes (%.2f GB)\n", ds.Capacity, float64(ds.Capacity)/(1024*1024*1024))
			fmt.Printf("    Used Space:            %d bytes (%.2f GB)\n", ds.UsedSpace, float64(ds.UsedSpace)/(1024*1024*1024))
			fmt.Printf("    Free Space:            %d bytes (%.2f GB)\n", ds.FreeSpace, float64(ds.FreeSpace)/(1024*1024*1024))
			if ds.Capacity > 0 {
				usage := float64(ds.UsedSpace) / float64(ds.Capacity) * 100
				fmt.Printf("    Usage:                 %.2f%%\n", usage)
			}
		}
	}

	// Display Snapshots
	if len(info.Snapshots) > 0 {
		fmt.Println("\n[SNAPSHOTS]")
		for idx, snap := range info.Snapshots {
			fmt.Printf("  Snapshot %d: %s\n", idx+1, snap.Name)
			if snap.State != "" {
				fmt.Printf("    State:                 %s\n", snap.State)
			}
			if snap.Description != "" {
				fmt.Printf("    Description:           %s\n", snap.Description)
			}
			if snap.CreateTime != "" {
				fmt.Printf("    Create Time:           %s\n", snap.CreateTime)
			}
			if snap.Size > 0 {
				fmt.Printf("    Size:                  %d bytes (%.2f GB)\n", snap.Size, float64(snap.Size)/(1024*1024*1024))
			}
			fmt.Printf("    Quiesced:              %v\n", snap.Quiesced)
			fmt.Printf("    Backup Mode:           %v\n", snap.BackupMode)
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

// parseConfigValueFlexible tries multiple key patterns to find a value
func parseConfigValueFlexible(output string, keys []string, target *string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, key := range keys {
			if strings.Contains(line, key) {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					value := strings.TrimSpace(strings.Trim(parts[1], "\",'"))
					// Skip <unset>, (unset) and other placeholder values - check with and without quotes
					value = strings.Trim(value, "\"'")
					if !isPlaceholderValue(value) && value != "" {
						*target = value
						return // Found, stop searching
					}
				}
				return // Found but empty/unset, stop searching
			}
		}
	}
}

// isPlaceholderValue checks if a value is a placeholder like <unset>
func isPlaceholderValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "<unset>" || value == "(unset)" || value == "unset" ||
	       value == "<unknown>" || value == "(unknown)" || value == "unknown"
}

// formatFieldValue formats a field value for display, handling quotes and placeholders
func formatFieldValue(value string) string {
	value = strings.Trim(value, "\"'")
	if isPlaceholderValue(value) || value == "" {
		return "N/A"
	}
	return value
}

func parseConfigValue(output, key string, target *string) {
	parseConfigValueFlexible(output, []string{key}, target)
}

// parseIntValueFlexible tries multiple key patterns to find an integer value
func parseIntValueFlexible(output string, keys []string, target *int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, key := range keys {
			if strings.Contains(line, key) {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					value := strings.TrimSpace(parts[1])
					fmt.Sscanf(value, "%d", target)
				}
				return // Found, stop searching
			}
		}
	}
}

func parseIntValue(output, key string, target *int) {
	parseIntValueFlexible(output, []string{key}, target)
}

// parseSizeValue parses human-readable size (1K, 1M, 1G) to bytes
func parseSizeValue(sizeStr string, target *int64) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var size float64
	var multiplier int64 = 1

	// Extract numeric part and unit
	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		fmt.Sscanf(sizeStr, "%f", &size)
	} else {
		fmt.Sscanf(sizeStr, "%f", &size)
	}

	*target = int64(size * float64(multiplier))
}

// Register registers the inspect command
func init() {
	config.Register("vm-inspect", func(params *config.Params) config.CommandInterface {
		return NewInspectVM(params)
	})
}
