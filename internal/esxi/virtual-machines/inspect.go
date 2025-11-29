package virtualmachines

import (
	"encoding/json"
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
	if i.params.VMInspectID == "" && i.params.VMInspectName == "" {
		return common.NewValidationError("vm-inspect-id or vm-inspect-name", "at least one must be provided")
	}
	return nil
}

// Execute gathers and outputs comprehensive VM information as JSON
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

	// Ensure all fields are properly initialized for JSON output
	if info.Networks == nil {
		info.Networks = []NetworkInfo{}
	}
	if info.Disks == nil {
		info.Disks = []DiskInfo{}
	}
	if info.Datastores == nil {
		info.Datastores = []DatastoreInfo{}
	}
	if info.Snapshots == nil {
		info.Snapshots = []SnapshotInfo{}
	}
	if info.VMXConfig == nil {
		info.VMXConfig = make(map[string]string)
	}
	if info.VMDKConfigs == nil {
		info.VMDKConfigs = []VMDKInfo{}
	}

	// Output as JSON with all fields
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return common.WrapError(err, "failed to marshal JSON")
	}

	fmt.Println(string(jsonData))
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

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// InspectVMInfo holds all gathered VM inspection information
type InspectVMInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	State       string          `json:"state"`
	PowerState  string          `json:"power_state"`
	UUID        string          `json:"uuid"`
	BiosUUID    string          `json:"bios_uuid"`
	ConfigPath  string          `json:"config_path"`
	Annotation  string          `json:"annotation"`
	CreateDate  string          `json:"create_date"`
	UpTime      string          `json:"up_time"`
	Version     string          `json:"version"`
	Firmware    string          `json:"firmware"`
	GuestOS     string          `json:"guest_os"`
	ToolsStatus string          `json:"tools_status"`
	ToolsVersion string         `json:"tools_version"`
	Hardware    HardwareInfo    `json:"hardware"`
	Networks    []NetworkInfo   `json:"networks"`
	Disks       []DiskInfo      `json:"disks"`
	Datastores  []DatastoreInfo `json:"datastores"`
	Snapshots   []SnapshotInfo  `json:"snapshots"`
	VMXConfig   map[string]string `json:"vmx_config"`
	VMDKConfigs []VMDKInfo      `json:"vmdk_configs"`

	// Internal fields not exported to JSON
	CPUs           int            `json:"-"`
	Memory         int            `json:"-"`
	NICs           int            `json:"-"`
	DiskCount      int            `json:"-"`
	MaxCPUs        int            `json:"-"`
	MaxMemory      int            `json:"-"`
	BootDelay      int            `json:"-"`
	ToolsRunning   string         `json:"-"`
	VMXPath        string         `json:"-"`
}

// HardwareInfo contains hardware specifications
type HardwareInfo struct {
	CPUs      int `json:"cpus"`
	Memory    int `json:"memory_mb"`
	MaxCPUs   int `json:"max_cpus"`
	MaxMemory int `json:"max_memory_mb"`
	BootDelay int `json:"boot_delay_ms"`
}

// NetworkInfo holds NIC information
type NetworkInfo struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	Network    string `json:"network"`
	Connected  bool   `json:"connected"`
}

// DiskInfo holds disk information
type DiskInfo struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size_bytes"`
	SizeGB     string `json:"size_gb"`
	Datastore  string `json:"datastore"`
	Controller string `json:"controller"`
	DeviceType string `json:"device_type"`
	Filename   string `json:"filename"`
}

// DatastoreInfo holds datastore information
type DatastoreInfo struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Capacity     int64   `json:"capacity_bytes"`
	CapacityGB   string  `json:"capacity_gb"`
	FreeSpace    int64   `json:"free_space_bytes"`
	FreeGB       string  `json:"free_space_gb"`
	UsedSpace    int64   `json:"used_space_bytes"`
	UsedGB       string  `json:"used_space_gb"`
	Type         string  `json:"type"`
	URL          string  `json:"url"`
	UsagePercent float64 `json:"usage_percent"`
}

// SnapshotInfo holds snapshot information
type SnapshotInfo struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CreateTime  string   `json:"create_time"`
	State       string   `json:"state"`
	ParentKey   string   `json:"parent_key"`
	ChildKeys   []string `json:"child_keys"`
	Quiesced    bool     `json:"quiesced"`
	BackupMode  bool     `json:"backup_mode"`
	Size        int64    `json:"size_bytes"`
	SizeGB      string   `json:"size_gb"`
}

// VMDKInfo holds VMDK descriptor file information
type VMDKInfo struct {
	Filename      string            `json:"filename"`
	Version       string            `json:"version"`
	Encoding      string            `json:"encoding"`
	CID           string            `json:"cid"`
	ParentCID     string            `json:"parent_cid"`
	CreateType    string            `json:"create_type"`
	Extents       []ExtentInfo      `json:"extents"`
	DDBParameters map[string]string `json:"ddb_parameters"`
	Capacity      int64             `json:"capacity_sectors"`
	AdapterType   string            `json:"adapter_type"`
	Geometry      GeometryInfo      `json:"geometry"`
	ThinProvisioned bool            `json:"thin_provisioned"`
	UUID          string            `json:"uuid"`
	VirtualHWVer  string            `json:"virtual_hw_version"`
	LongContentID string            `json:"long_content_id"`
}

// ExtentInfo holds extent description information
type ExtentInfo struct {
	Access   string `json:"access"`
	Sectors  int64  `json:"sectors"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

// GeometryInfo holds disk geometry information
type GeometryInfo struct {
	Cylinders int `json:"cylinders"`
	Heads     int `json:"heads"`
	Sectors   int `json:"sectors"`
}

// ============================================================================
// MAIN GATHERING LOGIC
// ============================================================================

// gatherVMInfo collects comprehensive VM information
func (i *InspectVM) gatherVMInfo(mgr *utils.SSHManager, vmID string) (*InspectVMInfo, error) {
	info := &InspectVMInfo{
		ID:        vmID,
		VMXConfig: make(map[string]string),
	}

	// Phase 1: Get basic VM info from vim-cmd
	if err := i.getBasicInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Phase 2: Get hardware and runtime info
	if err := i.getHardwareInfo(mgr, vmID, info); err != nil {
		return nil, err
	}

	// Phase 3: Parse VMX file for detailed configuration
	if err := i.parseVMXFile(mgr, info); err != nil {
		common.Debug("parse VMX file", "error", err.Error())
	}

	// Phase 4: Parse VMDK files for disk information (MUST be before disk extraction!)
	if err := i.parseVMDKFiles(mgr, info); err != nil {
		common.Debug("parse VMDK files", "error", err.Error())
	}

	// Phase 5: Extract organized data from VMX config (now VMDK data is available)
	if len(info.VMXConfig) > 0 {
		i.extractNetworkDetailsFromVMX(info)
		i.extractDiskDetailsFromVMX(info)
		if info.GuestOS == "" {
			if guestOS, ok := info.VMXConfig["guestOS"]; ok && guestOS != "" {
				info.GuestOS = guestOS
			}
		}
	}

	// Phase 6: Get snapshots
	if err := i.getSnapshots(mgr, vmID, info); err != nil {
		common.Debug("get snapshots", "error", err.Error())
	}

	// Phase 7: Get datastores
	if err := i.getDatastores(mgr, vmID, info); err != nil {
		common.Debug("get datastores", "error", err.Error())
	}

	// Phase 8: Populate hardware info
	info.Hardware = HardwareInfo{
		CPUs:      info.CPUs,
		Memory:    info.Memory,
		MaxCPUs:   info.MaxCPUs,
		MaxMemory: info.MaxMemory,
		BootDelay: info.BootDelay,
	}

	return info, nil
}

// ============================================================================
// BASIC INFO GATHERING
// ============================================================================

// getBasicInfo retrieves basic VM information from vim-cmd
func (i *InspectVM) getBasicInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get basic VM info")
	}

	// Basic info
	parseConfigValueFlexible(output, []string{"name =", "name="}, &info.Name)
	parseConfigValueFlexible(output, []string{"state =", "state=", "config.name.state =", "config.name.state="}, &info.State)
	parseConfigValueFlexible(output, []string{"config.annotation =", "config.annotation=", "annotation =", "annotation="}, &info.Annotation)

	// UUID info
	parseConfigValueFlexible(output, []string{"config.uuid =", "config.uuid=", "uuid =", "uuid="}, &info.UUID)
	parseConfigValueFlexible(output, []string{"uuid.bios =", "uuid.bios=", "bios.uuid =", "config.uuid.bios ="}, &info.BiosUUID)

	// Guest and tools info
	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname =", "guestOS =", "config.guestFullName ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"toolsRunningStatus =", "toolsRunningStatus=", "tools.runningStatus =", "guest.toolsRunningStatus ="}, &info.ToolsStatus)
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "tools.version =", "guest.toolsVersion =", "guestToolsVersion ="}, &info.ToolsVersion)

	// Power state
	parseConfigValueFlexible(output, []string{"powerState =", "powerState=", "runtime.powerState ="}, &info.PowerState)

	// Config file path
	parseConfigValueFlexible(output, []string{"config.files.vmPathName =", "config.files.vmPathName=", "vmPathName =", "vmPathName="}, &info.ConfigPath)

	// Fallback to get.config if config path not found
	if info.ConfigPath == "" {
		configOutput, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s 2>/dev/null", vmID))
		if err == nil && configOutput != "" {
			parseConfigValueFlexible(configOutput, []string{"configFile =", "configFile=", "config.files.vmPathName =", "vmPathName =", "files.vmPathName =", "path ="}, &info.ConfigPath)
		}
	}

	// Last resort: find .vmx file
	if info.ConfigPath == "" && info.Name != "" {
		searchOutput, err := mgr.RunCommand(fmt.Sprintf("find /vmfs/volumes -name '%s.vmx' 2>/dev/null | head -1", info.Name))
		if err == nil && strings.TrimSpace(searchOutput) != "" {
			info.ConfigPath = strings.TrimSpace(searchOutput)
		}
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
	parseIntValueFlexible(output, []string{"maxCpus =", "maxCpus=", "config.hardware.maxCpus ="}, &info.MaxCPUs)
	parseIntValueFlexible(output, []string{"maxMemory =", "maxMemory=", "config.hardware.maxMemory ="}, &info.MaxMemory)
	parseIntValueFlexible(output, []string{"numEthernetCards =", "numEthernetCards=", "config.hardware.numEthernetCards ="}, &info.NICs)
	parseIntValueFlexible(output, []string{"numVirtualDisks =", "numVirtualDisks=", "config.hardware.numVirtualDisks ="}, &info.DiskCount)

	return nil
}

// ============================================================================
// SNAPSHOT AND DATASTORE
// ============================================================================

// getSnapshots retrieves snapshot information
func (i *InspectVM) getSnapshots(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/snapshot.get %s 2>/dev/null", vmID))
	if err != nil || output == "" {
		info.Snapshots = nil
		return nil
	}

	info.Snapshots = nil

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "--") {
			continue
		}

		parts := strings.Split(line, "--")
		if len(parts) >= 2 {
			snapshotName := strings.TrimSpace(parts[0])
			if snapshotName == "" {
				continue
			}

			snapshot := SnapshotInfo{Name: snapshotName}

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

// getDatastores retrieves datastore information from disk file paths
func (i *InspectVM) getDatastores(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	seenDatastores := make(map[string]bool)
	var datastorePaths []string

	// Extract datastore paths from disk filenames
	for _, disk := range info.Disks {
		if disk.Path != "" {
			parts := strings.Split(disk.Path, "/")
			if strings.HasPrefix(disk.Path, "/vmfs/volumes/") && len(parts) > 3 {
				datastoreUUID := parts[3]
				if !seenDatastores[datastoreUUID] {
					seenDatastores[datastoreUUID] = true
					datastorePaths = append(datastorePaths, "/vmfs/volumes/"+datastoreUUID)
				}
			}
		}
	}

	// Also try to extract from config path if in VMware format
	if info.ConfigPath != "" && strings.HasPrefix(info.ConfigPath, "[") {
		fsPath := convertVMwarePath(info.ConfigPath, mgr)
		if fsPath != "" {
			parts := strings.Split(fsPath, "/")
			if len(parts) > 3 {
				datastoreUUID := parts[3]
				if !seenDatastores[datastoreUUID] {
					seenDatastores[datastoreUUID] = true
					datastorePaths = append(datastorePaths, "/vmfs/volumes/"+datastoreUUID)
				}
			}
		}
	}

	// Query each unique datastore for capacity information
	seenCapacities := make(map[int64]bool)
	for _, dsPath := range datastorePaths {
		dfOutput, err := mgr.RunCommand(fmt.Sprintf("df -h %s 2>/dev/null | tail -1", dsPath))
		if err != nil || strings.TrimSpace(dfOutput) == "" {
			continue
		}

		fields := strings.Fields(dfOutput)
		if len(fields) < 4 {
			continue
		}

		dsUUID := strings.TrimPrefix(dsPath, "/vmfs/volumes/")

		datastore := DatastoreInfo{
			Path: dsPath,
			Name: dsUUID,
			Type: "VMFS",
		}

		parseSizeValue(fields[1], &datastore.Capacity)
		parseSizeValue(fields[2], &datastore.UsedSpace)
		parseSizeValue(fields[3], &datastore.FreeSpace)

		// Skip if we've already added a datastore with this capacity (likely a symlink/duplicate)
		if seenCapacities[datastore.Capacity] {
			continue
		}
		seenCapacities[datastore.Capacity] = true

		// Calculate usage percentage
		if datastore.Capacity > 0 {
			datastore.UsagePercent = float64(datastore.UsedSpace) / float64(datastore.Capacity) * 100
		}

		// Convert to GB for convenience
		if datastore.Capacity > 0 {
			datastore.CapacityGB = fmt.Sprintf("%.2f GB", float64(datastore.Capacity)/1024/1024/1024)
		}
		if datastore.FreeSpace > 0 {
			datastore.FreeGB = fmt.Sprintf("%.2f GB", float64(datastore.FreeSpace)/1024/1024/1024)
		}
		if datastore.UsedSpace > 0 {
			datastore.UsedGB = fmt.Sprintf("%.2f GB", float64(datastore.UsedSpace)/1024/1024/1024)
		}

		info.Datastores = append(info.Datastores, datastore)
	}

	return nil
}

// ============================================================================
// VMX AND VMDK FILE PARSING
// ============================================================================

// parseVMXFile reads and parses the .vmx configuration file
func (i *InspectVM) parseVMXFile(mgr *utils.SSHManager, info *InspectVMInfo) error {
	if info.ConfigPath == "" {
		return fmt.Errorf("no config path available")
	}

	vmxPath := convertVMwarePath(info.ConfigPath, mgr)
	if vmxPath == "" {
		parts := strings.Split(info.ConfigPath, "/")
		vmxPath = strings.Join(parts[:len(parts)-1], "/") + "/" + info.Name + ".vmx"
	}

	output, err := mgr.RunCommand(fmt.Sprintf("cat '%s' 2>/dev/null", vmxPath))
	if err != nil {
		findOutput, findErr := mgr.RunCommand(fmt.Sprintf("find /vmfs/volumes -name '%s.vmx' 2>/dev/null | head -1", info.Name))
		if findErr == nil && strings.TrimSpace(findOutput) != "" {
			vmxPath = strings.TrimSpace(findOutput)
			output, err = mgr.RunCommand(fmt.Sprintf("cat '%s' 2>/dev/null", vmxPath))
			if err != nil {
				return common.WrapError(err, "failed to read VMX file")
			}
		} else {
			return common.WrapError(err, "failed to read VMX file")
		}
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

	vmxPath := convertVMwarePath(info.ConfigPath, mgr)
	if vmxPath == "" {
		parts := strings.Split(info.ConfigPath, "/")
		vmxPath = strings.Join(parts[:len(parts)-1], "/") + "/" + info.Name + ".vmx"
	}

	// Extract directory from config path
	parts := strings.Split(vmxPath, "/")
	vmxDir := strings.Join(parts[:len(parts)-1], "/")

	// List VMDK files in VM directory
	output, err := mgr.RunCommand(fmt.Sprintf("ls '%s'/*.vmdk 2>/dev/null | head -20", vmxDir))
	if err != nil || output == "" {
		return fmt.Errorf("no VMDK files found in %s", vmxDir)
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

		// Parse extent descriptions (they don't have "=" in them)
		if section == "extent" && (strings.HasPrefix(line, "RW ") || strings.HasPrefix(line, "RDONLY ")) {
			extent := i.parseExtentLine(line)
			vmdk.Extents = append(vmdk.Extents, extent)
			// Set capacity from the first extent's sectors
			if vmdk.Capacity == 0 && extent.Sectors > 0 {
				vmdk.Capacity = extent.Sectors
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(strings.Trim(parts[1], "\"' "))

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

// ============================================================================
// DISK AND NETWORK EXTRACTION FROM VMX
// ============================================================================

// extractNetworkDetailsFromVMX extracts network adapter info from VMX config
func (i *InspectVM) extractNetworkDetailsFromVMX(info *InspectVMInfo) {
	info.Networks = nil

	for idx := 0; idx < 10; idx++ {
		prefix := fmt.Sprintf("ethernet%d", idx)
		presentKey := prefix + ".present"

		if present, exists := info.VMXConfig[presentKey]; !exists || present != "TRUE" {
			continue
		}

		network := NetworkInfo{
			Index: idx,
			Name:  fmt.Sprintf("Network adapter %d", idx+1),
		}

		// Get MAC address
		if mac, ok := info.VMXConfig[prefix+".generatedAddress"]; ok && mac != "" {
			network.MacAddress = mac
		}

		// Get network name
		if netName, ok := info.VMXConfig[prefix+".networkName"]; ok && netName != "" {
			network.Network = netName
		}

		// Connected if has backing
		network.Connected = network.Network != ""

		info.Networks = append(info.Networks, network)
	}
}

// extractDiskDetailsFromVMX extracts disk info from VMX config and enriches with size/datastore
func (i *InspectVM) extractDiskDetailsFromVMX(info *InspectVMInfo) {
	info.Disks = nil

	// Build a map of VMDK filenames (basename) to their sizes from parsed VMDK configs
	vmdkSizeMap := make(map[string]int64)
	vmdkPathMap := make(map[string]string) // Map basename to full path
	for _, vmdk := range info.VMDKConfigs {
		// Capacity is in sectors, convert to bytes (512 bytes per sector)
		sizeBytes := vmdk.Capacity * 512

		// Extract just the filename (basename) for easier matching
		parts := strings.Split(vmdk.Filename, "/")
		basename := parts[len(parts)-1]

		vmdkSizeMap[basename] = sizeBytes
		vmdkPathMap[basename] = vmdk.Filename
	}

	// Check SCSI controllers
	for ctrl := 0; ctrl < 4; ctrl++ {
		ctrlPrefix := fmt.Sprintf("scsi%d", ctrl)
		ctrlPresentKey := ctrlPrefix + ".present"

		if present, exists := info.VMXConfig[ctrlPresentKey]; !exists || present != "TRUE" {
			continue
		}

		// Check for disks on this controller
		for device := 0; device < 10; device++ {
			deviceKey := fmt.Sprintf("%s:%d", ctrlPrefix, device)
			fileKey := deviceKey + ".fileName"
			presentKey := deviceKey + ".present"

			if present, exists := info.VMXConfig[presentKey]; !exists || present != "TRUE" {
				continue
			}

			disk := DiskInfo{
				Index:      len(info.Disks),
				Name:       fmt.Sprintf("Hard disk %d", len(info.Disks)+1),
				Controller: ctrlPrefix,
			}

			// Get filename
			if filename, ok := info.VMXConfig[fileKey]; ok && filename != "" {
				disk.Filename = filename
				disk.Path = filename

				// Try to get size from VMDK map using basename
				basename := filename
				if strings.Contains(filename, "/") {
					parts := strings.Split(filename, "/")
					basename = parts[len(parts)-1]
				}

				// Look up by basename first, then try full path
				var actualPath string
				if size, ok := vmdkSizeMap[basename]; ok && size > 0 {
					disk.Size = size
					disk.SizeGB = fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
					if p, ok := vmdkPathMap[basename]; ok {
						actualPath = p
					}
				}

				// Extract datastore name from actual path (if found) or disk filename
				datastorePath := actualPath
				if datastorePath == "" {
					datastorePath = filename
				}

				if strings.HasPrefix(datastorePath, "/vmfs/volumes/") {
					parts := strings.Split(datastorePath, "/")
					if len(parts) > 3 {
						disk.Datastore = parts[3]
					}
				}
			}

			// Get device type
			if devType, ok := info.VMXConfig[deviceKey+".deviceType"]; ok && devType != "" {
				disk.DeviceType = devType
			}

			info.Disks = append(info.Disks, disk)
		}
	}

	// Check SATA controllers
	for ctrl := 0; ctrl < 4; ctrl++ {
		ctrlPrefix := fmt.Sprintf("sata%d", ctrl)
		ctrlPresentKey := ctrlPrefix + ".present"

		if present, exists := info.VMXConfig[ctrlPresentKey]; !exists || present != "TRUE" {
			continue
		}

		// Check for devices on this controller
		for device := 0; device < 10; device++ {
			deviceKey := fmt.Sprintf("%s:%d", ctrlPrefix, device)
			fileKey := deviceKey + ".fileName"
			presentKey := deviceKey + ".present"

			if present, exists := info.VMXConfig[presentKey]; !exists || present != "TRUE" {
				continue
			}

			disk := DiskInfo{
				Index:      len(info.Disks),
				Name:       fmt.Sprintf("Device %d", len(info.Disks)+1),
				Controller: ctrlPrefix,
			}

			// Get filename
			if filename, ok := info.VMXConfig[fileKey]; ok && filename != "" {
				disk.Filename = filename
				disk.Path = filename

				// Try to get size from VMDK map using basename
				basename := filename
				if strings.Contains(filename, "/") {
					parts := strings.Split(filename, "/")
					basename = parts[len(parts)-1]
				}

				// Look up by basename first, then try full path
				var actualPath string
				if size, ok := vmdkSizeMap[basename]; ok && size > 0 {
					disk.Size = size
					disk.SizeGB = fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
					if p, ok := vmdkPathMap[basename]; ok {
						actualPath = p
					}
				}

				// Extract datastore name from actual path (if found) or disk filename
				datastorePath := actualPath
				if datastorePath == "" {
					datastorePath = filename
				}

				if strings.HasPrefix(datastorePath, "/vmfs/volumes/") {
					parts := strings.Split(datastorePath, "/")
					if len(parts) > 3 {
						disk.Datastore = parts[3]
					}
				}
			}

			// Get device type
			if devType, ok := info.VMXConfig[deviceKey+".deviceType"]; ok && devType != "" {
				disk.DeviceType = devType
			}

			info.Disks = append(info.Disks, disk)
		}
	}
}

// ============================================================================
// PARSING HELPER FUNCTIONS
// ============================================================================

// parseConfigValueFlexible tries multiple key patterns to find a value
func parseConfigValueFlexible(output string, keys []string, target *string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		for _, key := range keys {
			fieldName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(key), "="), ":")

			if matchesFieldAtStart(line, fieldName) {
				// Try to extract value after = sign
				if strings.Contains(line, "=") {
					idx := strings.Index(line, "=")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" && !isPlaceholderValue(value) {
							*target = value
							return
						}
					}
				}
				// Try to extract value after : sign
				if strings.Contains(line, ":") && !strings.Contains(line, "://") {
					idx := strings.Index(line, ":")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" && !isPlaceholderValue(value) {
							*target = value
							return
						}
					}
				}
				return
			}
		}
	}
}

// parseIntValueFlexible tries multiple key patterns to find an integer value
func parseIntValueFlexible(output string, keys []string, target *int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		for _, key := range keys {
			fieldName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(key), "="), ":")

			if matchesFieldAtStart(line, fieldName) {
				if strings.Contains(line, "=") {
					idx := strings.Index(line, "=")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" {
							fmt.Sscanf(value, "%d", target)
							return
						}
					}
				}
				if strings.Contains(line, ":") && !strings.Contains(line, "://") {
					idx := strings.Index(line, ":")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" {
							fmt.Sscanf(value, "%d", target)
							return
						}
					}
				}
				return
			}
		}
	}
}

// matchesFieldAtStart checks if a line starts with a field name (respecting word boundaries)
func matchesFieldAtStart(line, fieldName string) bool {
	lineLower := strings.ToLower(line)
	fieldLower := strings.ToLower(fieldName)

	if !strings.HasPrefix(lineLower, fieldLower) {
		return false
	}

	if len(lineLower) > len(fieldLower) {
		nextChar := lineLower[len(fieldLower)]
		return nextChar == ' ' || nextChar == '\t' || nextChar == '=' || nextChar == ':' || nextChar == '.'
	}

	return true
}

// convertVMwarePath converts VMware format "[datastore] path/to/file" to actual filesystem path
func convertVMwarePath(vmwarePath string, mgr *utils.SSHManager) string {
	if !strings.HasPrefix(vmwarePath, "[") || !strings.Contains(vmwarePath, "]") {
		return ""
	}

	endBracket := strings.Index(vmwarePath, "]")
	if endBracket <= 1 {
		return ""
	}
	datastoreName := vmwarePath[1:endBracket]
	relativePath := strings.TrimSpace(vmwarePath[endBracket+1:])

	output, err := mgr.RunCommand(fmt.Sprintf("ls -d /vmfs/volumes/* 2>/dev/null | grep -i '%s' | head -1", datastoreName))
	if err != nil || strings.TrimSpace(output) == "" {
		return ""
	}

	dsPath := strings.TrimSpace(output)
	return dsPath + "/" + relativePath
}

// extractValue cleans up a value extracted from vim-cmd output
func extractValue(raw string) string {
	value := strings.TrimSpace(raw)

	for strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		value = value[1:]
	}

	for strings.HasSuffix(value, "\"") || strings.HasSuffix(value, "'") || strings.HasSuffix(value, ",") {
		if strings.HasSuffix(value, "\"") {
			value = strings.TrimSuffix(value, "\"")
		} else if strings.HasSuffix(value, "'") {
			value = strings.TrimSuffix(value, "'")
		} else {
			value = strings.TrimSuffix(value, ",")
		}
	}

	value = strings.TrimSpace(value)
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		value = value[1 : len(value)-1]
	}

	value = strings.TrimSpace(value)
	value = strings.Join(strings.Fields(value), " ")

	return value
}

// isPlaceholderValue checks if a value is a placeholder like <unset>
func isPlaceholderValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "<unset>" || value == "(unset)" || value == "unset" ||
		value == "<unknown>" || value == "(unknown)" || value == "unknown"
}

// parseSizeValue parses human-readable size (1K, 1M, 1G) to bytes
func parseSizeValue(sizeStr string, target *int64) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var size float64
	var multiplier int64 = 1

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

// ============================================================================
// COMMAND REGISTRATION
// ============================================================================

// Register registers the inspect command
func init() {
	config.Register("vm-inspect", func(params *config.Params) config.CommandInterface {
		return NewInspectVM(params)
	})
}
