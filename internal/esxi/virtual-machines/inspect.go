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
	// Either ID or Name must be provided
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

	// Populate hardware info from legacy fields
	info.Hardware = HardwareInfo{
		CPUs:      info.CPUs,
		Memory:    info.Memory,
		MaxCPUs:   info.MaxCPUs,
		MaxMemory: info.MaxMemory,
		BootDelay: info.BootDelay,
	}

	// Reorganize network and storage data
	if info.NICs > 0 && len(info.Networks) == 0 {
		// Auto-generate network adapters if they weren't populated
		for idx := 0; idx < info.NICs; idx++ {
			info.Networks = append(info.Networks, NetworkInfo{
				Index: idx,
				Name:  fmt.Sprintf("Network adapter %d", idx+1),
			})
		}
	}
	if info.DiskCount > 0 && len(info.Disks) == 0 {
		// Auto-generate disks if they weren't populated
		for idx := 0; idx < info.DiskCount; idx++ {
			info.Disks = append(info.Disks, DiskInfo{
				Index: idx,
				Name:  fmt.Sprintf("Hard disk %d", idx+1),
			})
		}
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

// InspectVMInfo holds all gathered VM inspection information
type InspectVMInfo struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	State         string         `json:"state"`
	PowerState    string         `json:"power_state"`
	UUID          string         `json:"uuid"`
	BiosUUID      string         `json:"bios_uuid"`
	ConfigPath    string         `json:"config_path"`
	Annotation    string         `json:"annotation"`
	CreateDate    string         `json:"create_date"`
	UpTime        string         `json:"up_time"`
	Version       string         `json:"version"`
	Firmware      string         `json:"firmware"`
	GuestOS       string         `json:"guest_os"`
	ToolsStatus   string         `json:"tools_status"`
	ToolsVersion  string         `json:"tools_version"`
	Hardware      HardwareInfo   `json:"hardware"`
	Networks      []NetworkInfo  `json:"networks"`
	Disks         []DiskInfo     `json:"disks"`
	Datastores    []DatastoreInfo `json:"datastores"`
	Snapshots     []SnapshotInfo `json:"snapshots"`
	VMXConfig     map[string]string `json:"vmx_config"`
	VMDKConfigs   []VMDKInfo     `json:"vmdk_configs"`

	// Legacy fields kept for backward compatibility (not exported to JSON)
	Uuid           string      `json:"-"`
	BiosUuid       string      `json:"-"`
	CPUs           int         `json:"-"`
	Memory         int         `json:"-"`
	NICs           int         `json:"-"`
	DiskCount      int         `json:"-"`
	MaxCPUs        int         `json:"-"`
	MaxMemory      int         `json:"-"`
	BootDelay      int         `json:"-"`
	ToolsRunning   string      `json:"-"`
	VMXPath        string      `json:"-"`
	CPUInfo        CPUInfo     `json:"-"`
	StorageDevices []DiskInfo  `json:"-"`
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
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Capacity    int64   `json:"capacity_bytes"`
	CapacityGB  string  `json:"capacity_gb"`
	FreeSpace   int64   `json:"free_space_bytes"`
	FreeGB      string  `json:"free_space_gb"`
	UsedSpace   int64   `json:"used_space_bytes"`
	UsedGB      string  `json:"used_space_gb"`
	Type        string  `json:"type"`
	URL         string  `json:"url"`
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

// CPUInfo holds CPU configuration
type CPUInfo struct {
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`
	HZ      string `json:"hz"`
}

// VMDKInfo holds VMDK descriptor file information
type VMDKInfo struct {
	Filename        string            `json:"filename"`
	Version         string            `json:"version"`
	Encoding        string            `json:"encoding"`
	CID             string            `json:"cid"`
	ParentCID       string            `json:"parent_cid"`
	CreateType      string            `json:"create_type"`
	Extents         []ExtentInfo      `json:"extents"`
	DDBParameters   map[string]string `json:"ddb_parameters"`
	Capacity        int64             `json:"capacity_sectors"`
	AdapterType     string            `json:"adapter_type"`
	Geometry        GeometryInfo      `json:"geometry"`
	ThinProvisioned bool              `json:"thin_provisioned"`
	UUID            string            `json:"uuid"`
	VirtualHWVer    string            `json:"virtual_hw_version"`
	LongContentID   string            `json:"long_content_id"`
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

// gatherVMInfo collects comprehensive VM information
func (i *InspectVM) gatherVMInfo(mgr *utils.SSHManager, vmID string) (*InspectVMInfo, error) {
	info := &InspectVMInfo{
		ID:        vmID,
		VMXConfig: make(map[string]string),
		Uuid:      "", // Will be populated by getBasicInfo
	}

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

	// Extract network details from VMX config
	if len(info.VMXConfig) > 0 {
		i.extractNetworkDetailsFromVMX(info)
		i.extractDiskDetailsFromVMX(info)
		i.extractGuestOSFromVMX(info)
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
		// No snapshots - this is not an error
		info.Snapshots = nil
		return nil
	}

	// Clear snapshots array
	info.Snapshots = nil

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "--") {
			continue
		}

		// Parse snapshot lines
		parts := strings.Split(line, "--")
		if len(parts) >= 2 {
			snapshotName := strings.TrimSpace(parts[0])
			if snapshotName == "" {
				continue // Skip empty snapshot names
			}

			snapshot := SnapshotInfo{
				Name: snapshotName,
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

// getDatastores retrieves datastore information from disk file paths
func (i *InspectVM) getDatastores(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	// Collect unique datastores from disk paths
	seenDatastores := make(map[string]bool)
	var datastorePaths []string

	// Extract datastore paths from disk filenames
	for _, disk := range info.Disks {
		if disk.Path != "" {
			// Parse path to extract datastore UUID
			parts := strings.Split(disk.Path, "/")
			if strings.HasPrefix(disk.Path, "/vmfs/volumes/") && len(parts) > 3 {
				// Format: /vmfs/volumes/uuid/path/to/file
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
		// Extract from [datastore] path format using convertVMwarePath
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
		// Get datastore size info
		dfOutput, err := mgr.RunCommand(fmt.Sprintf("df -h %s 2>/dev/null | tail -1", dsPath))
		if err != nil || strings.TrimSpace(dfOutput) == "" {
			continue
		}

		// Parse df output: filesystem total used available use%
		fields := strings.Fields(dfOutput)
		if len(fields) < 4 {
			continue
		}

		// Extract datastore name from UUID
		dsUUID := strings.TrimPrefix(dsPath, "/vmfs/volumes/")

		datastore := DatastoreInfo{
			Path: dsPath,
			Name: dsUUID,
			Type: "VMFS",
		}

		// Parse capacity, used, and free space
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

// parseVMXFile reads and parses the .vmx configuration file
func (i *InspectVM) parseVMXFile(mgr *utils.SSHManager, info *InspectVMInfo) error {
	if info.ConfigPath == "" {
		return fmt.Errorf("no config path available")
	}

	// Convert VMware format [datastore] path/to/file to actual path
	vmxPath := convertVMwarePath(info.ConfigPath, mgr)
	if vmxPath == "" {
		// Try extracting directory from config path as fallback
		parts := strings.Split(info.ConfigPath, "/")
		vmxPath = strings.Join(parts[:len(parts)-1], "/") + "/" + info.Name + ".vmx"
	}

	output, err := mgr.RunCommand(fmt.Sprintf("cat '%s' 2>/dev/null", vmxPath))
	if err != nil {
		// Try to find .vmx file using find command
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

	// Convert VMware format to actual path
	vmxPath := convertVMwarePath(info.ConfigPath, mgr)
	if vmxPath == "" {
		// Fallback: extract directory from config path
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
	parseConfigValueFlexible(output, []string{"state =", "state=", "config.name.state =", "config.name.state="}, &info.State)
	parseConfigValueFlexible(output, []string{"config.annotation =", "config.annotation=", "annotation =", "annotation="}, &info.Annotation)

	// UUID fields
	parseConfigValueFlexible(output, []string{"config.uuid =", "config.uuid=", "uuid =", "uuid="}, &info.UUID)
	info.Uuid = info.UUID // Keep legacy field in sync
	parseConfigValueFlexible(output, []string{"uuid.bios =", "uuid.bios=", "bios.uuid =", "config.uuid.bios ="}, &info.BiosUUID)
	info.BiosUuid = info.BiosUUID // Keep legacy field in sync

	// Guest info
	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname =", "guestOS =", "config.guestFullName ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"toolsRunningStatus =", "toolsRunningStatus=", "tools.runningStatus =", "guest.toolsRunningStatus ="}, &info.ToolsStatus)
	info.ToolsRunning = info.ToolsStatus // Keep legacy field in sync
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "tools.version =", "guest.toolsVersion =", "guestToolsVersion ="}, &info.ToolsVersion)

	// Power state
	parseConfigValueFlexible(output, []string{"powerState =", "powerState=", "runtime.powerState ="}, &info.PowerState)

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
// It's very flexible to handle different vim-cmd output formats
func parseConfigValueFlexible(output string, keys []string, target *string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		for _, key := range keys {
			// Extract field name from key (remove trailing separators)
			fieldName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(key), "="), ":")

			// Check if line starts with this field name (with word boundary)
			if matchesFieldAtStart(line, fieldName) {
				// Try to extract value after = sign
				if strings.Contains(line, "=") {
					idx := strings.Index(line, "=")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" && !isPlaceholderValue(value) {
							*target = value
							return // Found, stop searching
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
							return // Found, stop searching
						}
					}
				}
				return // Found but couldn't extract, stop searching
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

	// Check for word boundary after field name
	if len(lineLower) > len(fieldLower) {
		nextChar := lineLower[len(fieldLower)]
		// Valid boundaries: whitespace, =, :, .
		return nextChar == ' ' || nextChar == '\t' || nextChar == '=' || nextChar == ':' || nextChar == '.'
	}

	return true // Line is exactly the field name
}

// extractNetworkDetailsFromVMX extracts network adapter info from VMX config
func (i *InspectVM) extractNetworkDetailsFromVMX(info *InspectVMInfo) {
	// Clear existing networks from simple generation
	info.Networks = nil

	// Ethernet adapters are named ethernet0, ethernet1, etc.
	for idx := 0; idx < 10; idx++ {
		prefix := fmt.Sprintf("ethernet%d", idx)
		presentKey := prefix + ".present"

		// Check if this adapter is present
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

		// For connected status, check if device has a backing
		network.Connected = network.Network != ""

		info.Networks = append(info.Networks, network)
	}
}

// extractDiskDetailsFromVMX extracts disk info from VMX config
func (i *InspectVM) extractDiskDetailsFromVMX(info *InspectVMInfo) {
	// Clear existing disks from simple generation
	info.Disks = nil

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
				Index: len(info.Disks),
				Name:  fmt.Sprintf("Hard disk %d", len(info.Disks)+1),
			}

			// Get filename
			if filename, ok := info.VMXConfig[fileKey]; ok && filename != "" {
				disk.Filename = filename
				disk.Path = filename
			}

			// Get device type
			if devType, ok := info.VMXConfig[deviceKey+".deviceType"]; ok && devType != "" {
				disk.DeviceType = devType
			}

			// Set controller type
			disk.Controller = ctrlPrefix

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
				Index: len(info.Disks),
				Name:  fmt.Sprintf("Device %d", len(info.Disks)+1),
			}

			// Get filename
			if filename, ok := info.VMXConfig[fileKey]; ok && filename != "" {
				disk.Filename = filename
				disk.Path = filename
			}

			// Get device type
			if devType, ok := info.VMXConfig[deviceKey+".deviceType"]; ok && devType != "" {
				disk.DeviceType = devType
			}

			// Set controller type
			disk.Controller = ctrlPrefix

			info.Disks = append(info.Disks, disk)
		}
	}
}

// extractGuestOSFromVMX extracts guest OS from VMX config
func (i *InspectVM) extractGuestOSFromVMX(info *InspectVMInfo) {
	if info.GuestOS == "" {
		if guestOS, ok := info.VMXConfig["guestOS"]; ok && guestOS != "" {
			info.GuestOS = guestOS
		}
	}
}

// convertVMwarePath converts VMware format "[datastore] path/to/file" to actual filesystem path
func convertVMwarePath(vmwarePath string, mgr *utils.SSHManager) string {
	// Check if it's in VMware format
	if !strings.HasPrefix(vmwarePath, "[") || !strings.Contains(vmwarePath, "]") {
		return "" // Not in VMware format
	}

	// Extract datastore name: [datastore] -> datastore
	endBracket := strings.Index(vmwarePath, "]")
	if endBracket <= 1 {
		return ""
	}
	datastoreName := vmwarePath[1:endBracket]
	relativePath := strings.TrimSpace(vmwarePath[endBracket+1:])

	// Get actual datastore path by searching /vmfs/volumes
	output, err := mgr.RunCommand(fmt.Sprintf("ls -d /vmfs/volumes/* 2>/dev/null | grep -i '%s' | head -1", datastoreName))
	if err != nil || strings.TrimSpace(output) == "" {
		return ""
	}

	dsPath := strings.TrimSpace(output)
	return dsPath + "/" + relativePath
}

// extractValue cleans up a value extracted from vim-cmd output
func extractValue(raw string) string {
	// Remove leading/trailing whitespace
	value := strings.TrimSpace(raw)

	// Remove leading quotes
	for strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		value = value[1:]
	}

	// Remove trailing quotes and commas
	for strings.HasSuffix(value, "\"") || strings.HasSuffix(value, "'") || strings.HasSuffix(value, ",") {
		if strings.HasSuffix(value, "\"") {
			value = strings.TrimSuffix(value, "\"")
		} else if strings.HasSuffix(value, "'") {
			value = strings.TrimSuffix(value, "'")
		} else {
			value = strings.TrimSuffix(value, ",")
		}
	}

	// If the value starts and ends with matching quotes, remove them
	value = strings.TrimSpace(value)
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		value = value[1 : len(value)-1]
	}

	// Clean up multiple spaces
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
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		for _, key := range keys {
			// Extract field name from key (remove trailing separators)
			fieldName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(key), "="), ":")

			// Check if line starts with this field name (with word boundary)
			if matchesFieldAtStart(line, fieldName) {
				// Try to extract value after = sign
				if strings.Contains(line, "=") {
					idx := strings.Index(line, "=")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" {
							fmt.Sscanf(value, "%d", target)
							return // Found, stop searching
						}
					}
				}
				// Try to extract value after : sign
				if strings.Contains(line, ":") && !strings.Contains(line, "://") {
					idx := strings.Index(line, ":")
					if idx >= 0 {
						value := extractValue(line[idx+1:])
						if value != "" {
							fmt.Sscanf(value, "%d", target)
							return // Found, stop searching
						}
					}
				}
				return // Found but couldn't extract, stop searching
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
