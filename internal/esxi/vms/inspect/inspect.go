package inspect

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
	if info.VSwitches == nil {
		info.VSwitches = []VSwitchInfo{}
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
	// Basic identification
	ID   string `json:"id"`
	Name string `json:"name"`

	// Power and connection state
	PowerState       string `json:"power_state,omitempty"`
	ConnectionState  string `json:"connection_state,omitempty"`
	BootTime         string `json:"boot_time,omitempty"`
	FaultTolerance   string `json:"fault_tolerance_state,omitempty"`

	// UUIDs and identifiers
	UUID         string `json:"uuid,omitempty"`
	InstanceUUID string `json:"instance_uuid,omitempty"`
	BiosUUID     string `json:"bios_uuid,omitempty"`

	// Configuration
	ConfigPath  string `json:"config_path,omitempty"`
	Version     string `json:"version,omitempty"`
	Firmware    string `json:"firmware,omitempty"`
	GuestOS     string `json:"guest_os,omitempty"`
	GuestID     string `json:"guest_id,omitempty"`

	// VM metadata
	Annotation string `json:"annotation,omitempty"`
	CreateDate string `json:"create_date,omitempty"`
	Template   bool   `json:"template,omitempty"`

	// Tools and runtime
	ToolsStatus  string `json:"tools_status,omitempty"`
	ToolsVersion string `json:"tools_version,omitempty"`
	ToolsType    string `json:"tools_type,omitempty"`

	// Hardware and resources
	Hardware config.HardwareInfo `json:"hardware"`

	// Usage metrics
	MaxCpuUsage   int `json:"max_cpu_usage,omitempty"`
	MaxMemoryUsage int `json:"max_memory_usage,omitempty"`

	// Infrastructure
	Networks    []NetworkInfo   `json:"networks"`
	VSwitches   []VSwitchInfo   `json:"vswitches,omitempty"`
	Disks       []DiskInfo      `json:"disks"`
	Datastores  []DatastoreInfo `json:"datastores"`
	Snapshots   []SnapshotInfo  `json:"snapshots,omitempty"`

	// Raw configuration files
	VMXConfig   map[string]string `json:"vmx_config,omitempty"`
	VMDKConfigs []VMDKInfo        `json:"vmdk_configs,omitempty"`

	// Internal fields not exported to JSON
	CPUs         int    `json:"-"`
	Memory       int    `json:"-"`
	NICs         int    `json:"-"`
	DiskCount    int    `json:"-"`
	MaxCPUs      int    `json:"-"`
	MaxMemory    int    `json:"-"`
	BootDelay    int    `json:"-"`
	ToolsRunning string `json:"-"`
	VMXPath      string `json:"-"`
}


// NetworkInfo holds NIC information with detailed vswitch/portgroup details
type NetworkInfo struct {
	Index          int                      `json:"index"`
	Name           string                   `json:"name"`
	MacAddress     string                   `json:"mac_address"`
	Network        string                   `json:"network"`
	Connected      bool                     `json:"connected"`
	VSwitch        string                   `json:"vswitch,omitempty"`
	VLANID         int                      `json:"vlan_id,omitempty"`
	ActiveClients  int                      `json:"active_clients,omitempty"`
	Accessible     bool                     `json:"accessible,omitempty"`
	VMCount        int                      `json:"vm_count,omitempty"`
	ActivePorts    int                      `json:"active_ports,omitempty"`
	Security       *config.SecurityPolicy   `json:"security_policy,omitempty"`
	NICTeaming     *config.NICTeamingPolicy `json:"nic_teaming_policy,omitempty"`
	Shaping        *config.ShapingPolicy    `json:"shaping_policy,omitempty"`
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

// DatastoreInfo holds datastore information with comprehensive VMFS/NFS metadata
type DatastoreInfo struct {
	Name            string  `json:"name"`
	Path            string  `json:"path"`
	Type            string  `json:"type"`
	Capacity        int64   `json:"capacity_bytes"`
	CapacityGB      string  `json:"capacity_gb"`
	FreeSpace       int64   `json:"free_space_bytes"`
	FreeGB          string  `json:"free_space_gb"`
	UsedSpace       int64   `json:"used_space_bytes"`
	UsedGB          string  `json:"used_space_gb"`
	UsagePercent    float64 `json:"usage_percent"`
	UUID            string  `json:"uuid,omitempty"`
	MountPoint      string  `json:"mount_point,omitempty"`
	Version         string  `json:"version,omitempty"`
	Local           bool    `json:"local,omitempty"`
	BlockSize       string  `json:"block_size,omitempty"`
	HostCount       int     `json:"host_count,omitempty"`
	VMCount         int     `json:"vm_count,omitempty"`
	Mounted         bool    `json:"mounted,omitempty"`
	Accessible      bool    `json:"accessible,omitempty"`
	Extents         []string `json:"extents,omitempty"`
	URL             string  `json:"url,omitempty"`
}

// VSwitchInfo holds virtual switch information
type VSwitchInfo struct {
	Name           string            `json:"name"`
	Type           string            `json:"type,omitempty"`
	PortGroupCount int               `json:"port_group_count,omitempty"`
	Uplinks        []string          `json:"uplinks,omitempty"`
	MTU            int               `json:"mtu,omitempty"`
	Ports          int               `json:"ports,omitempty"`
	AvailablePorts int               `json:"available_ports,omitempty"`
	LinkDiscovery  string            `json:"link_discovery,omitempty"`
	AttachedVMs    int               `json:"attached_vms,omitempty"`
	ActiveVMs      int               `json:"active_vms,omitempty"`
	Security       *config.SecurityPolicy   `json:"security_policy,omitempty"`
	NICTeaming     *config.NICTeamingPolicy `json:"nic_teaming_policy,omitempty"`
	Shaping        *config.ShapingPolicy    `json:"shaping_policy,omitempty"`
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

	// Phase 5b: Enrich network details with infrastructure info (vswitch, VLAN, active clients)
	i.enrichNetworkDetailsWithInfrastructure(mgr, info)

	// Phase 6: Get snapshots
	if err := i.getSnapshots(mgr, vmID, info); err != nil {
		common.Debug("get snapshots", "error", err.Error())
	}

	// Phase 7: Get datastores
	if err := i.getDatastores(mgr, vmID, info); err != nil {
		common.Debug("get datastores", "error", err.Error())
	}

	// Phase 7b: Enrich datastore details with metadata (UUID, mount point, type)
	i.enrichDatastoreDetailsWithMetadata(mgr, info)

	// Phase 8: Enrich vswitch details from network vswitches
	i.enrichVSwitchDetailsWithInfrastructure(mgr, info)

	// Phase 9: Populate hardware info
	info.Hardware = config.HardwareInfo{
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

// getBasicInfo retrieves basic VM information from vim-cmd get.summary
func (i *InspectVM) getBasicInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.summary %s", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get basic VM info")
	}

	// Basic identification
	parseConfigValueFlexible(output, []string{"name =", "name="}, &info.Name)

	// Power and connection state
	parseConfigValueFlexible(output, []string{"powerState =", "powerState=", "runtime.powerState ="}, &info.PowerState)
	parseConfigValueFlexible(output, []string{"connectionState =", "connectionState=", "runtime.connectionState ="}, &info.ConnectionState)
	parseConfigValueFlexible(output, []string{"bootTime =", "bootTime=", "runtime.bootTime ="}, &info.BootTime)
	parseConfigValueFlexible(output, []string{"faultToleranceState =", "faultToleranceState=", "runtime.faultToleranceState ="}, &info.FaultTolerance)

	// UUIDs
	parseConfigValueFlexible(output, []string{"uuid =", "uuid=", "config.uuid =", "config.uuid="}, &info.UUID)
	parseConfigValueFlexible(output, []string{"uuid.bios =", "uuid.bios=", "bios.uuid =", "config.uuid.bios ="}, &info.BiosUUID)

	// Guest and tools from summary
	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname =", "guestOS ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"toolsRunningStatus =", "toolsRunningStatus=", "tools.runningStatus ="}, &info.ToolsStatus)
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "tools.version ="}, &info.ToolsVersion)

	// Usage metrics
	parseIntValueFlexible(output, []string{"maxCpuUsage =", "maxCpuUsage=", "runtime.maxCpuUsage ="}, &info.MaxCpuUsage)
	parseIntValueFlexible(output, []string{"maxMemoryUsage =", "maxMemoryUsage=", "runtime.maxMemoryUsage ="}, &info.MaxMemoryUsage)

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

// getHardwareInfo retrieves hardware configuration and metadata from vim-cmd get.config
func (i *InspectVM) getHardwareInfo(mgr *utils.SSHManager, vmID string, info *InspectVMInfo) error {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s 2>/dev/null", vmID))
	if err != nil {
		return common.WrapError(err, "failed to get hardware info")
	}

	// Hardware specifications
	parseIntValueFlexible(output, []string{"memoryMB =", "memoryMB=", "config.hardware.memoryMB ="}, &info.Memory)
	parseIntValueFlexible(output, []string{"numCPU =", "numCPU=", "config.hardware.numCPU ="}, &info.CPUs)
	parseConfigValueFlexible(output, []string{"version =", "version=", "config.version ="}, &info.Version)
	parseConfigValueFlexible(output, []string{"firmware =", "firmware=", "config.hardware.firmware ="}, &info.Firmware)
	parseIntValueFlexible(output, []string{"bootDelay =", "bootDelay=", "config.hardware.bootDelay ="}, &info.BootDelay)
	parseIntValueFlexible(output, []string{"maxCpus =", "maxCpus=", "config.hardware.maxCpus ="}, &info.MaxCPUs)
	parseIntValueFlexible(output, []string{"maxMemory =", "maxMemory=", "config.hardware.maxMemory ="}, &info.MaxMemory)

	// Hardware details
	parseIntValueFlexible(output, []string{"numCoresPerSocket =", "numCoresPerSocket=", "config.hardware.numCoresPerSocket ="}, &info.Hardware.CoresPerSocket)
	parseIntValueFlexible(output, []string{"simultaneousThreads =", "simultaneousThreads=", "config.hardware.simultaneousThreads ="}, &info.Hardware.SimultaneousThreads)
	parseConfigValueFlexible(output, []string{"motherboardLayout =", "motherboardLayout=", "config.hardware.motherboardLayout ="}, &info.Hardware.MotherboardLayout)

	// VM metadata
	parseConfigValueFlexible(output, []string{"createDate =", "createDate=", "config.createDate ="}, &info.CreateDate)
	parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "config.guestFullName ="}, &info.GuestOS)
	parseConfigValueFlexible(output, []string{"guestId =", "guestId=", "config.guestId ="}, &info.GuestID)
	parseConfigValueFlexible(output, []string{"uuid =", "uuid=", "config.uuid ="}, &info.UUID)
	parseConfigValueFlexible(output, []string{"instanceUuid =", "instanceUuid=", "config.instanceUuid ="}, &info.InstanceUUID)

	// Annotation and metadata
	parseConfigValueFlexible(output, []string{"annotation =", "annotation=", "config.annotation ="}, &info.Annotation)

	// Tools info from config
	parseConfigValueFlexible(output, []string{"toolsVersion =", "toolsVersion=", "config.tools.toolsVersion =", "tools.toolsVersion ="}, &info.ToolsVersion)
	parseConfigValueFlexible(output, []string{"toolsInstallType =", "toolsInstallType=", "config.tools.toolsInstallType ="}, &info.ToolsType)

	// Network and disk counts
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

// enrichNetworkDetailsWithInfrastructure populates network info with comprehensive vswitch and portgroup details
func (i *InspectVM) enrichNetworkDetailsWithInfrastructure(mgr *utils.SSHManager, info *InspectVMInfo) {
	if len(info.Networks) == 0 {
		return
	}

	// Get all port groups with esxcli
	output, err := mgr.RunCommand("esxcli network vswitch standard portgroup list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		common.Debug("enrich networks", "error", "failed to get portgroups from esxcli")
		return
	}

	// Parse esxcli output - format: Name VSwitch ActiveClients VLANID
	// Build a map of portgroup name to vswitch info
	portgroupInfo := make(map[string]map[string]string)
	lines := strings.Split(output, "\n")

	// Skip header line
	for idx, line := range lines {
		if idx == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pgName := fields[0]
		vswitchName := fields[1]
		activeClients := fields[2]
		vlanID := fields[3]

		portgroupInfo[pgName] = map[string]string{
			"vswitch":       vswitchName,
			"vlan":          vlanID,
			"activeClients": activeClients,
		}
	}

	// Enrich each network with infrastructure details
	for idx, network := range info.Networks {
		if pgInfo, ok := portgroupInfo[network.Network]; ok {
			// Basic info
			if vswitchName, ok := pgInfo["vswitch"]; ok {
				info.Networks[idx].VSwitch = vswitchName
			}
			if vlanStr, ok := pgInfo["vlan"]; ok {
				vlanID := 0
				fmt.Sscanf(vlanStr, "%d", &vlanID)
				if vlanID >= 0 {
					info.Networks[idx].VLANID = vlanID
				}
			}
			if clientsStr, ok := pgInfo["activeClients"]; ok {
				clients := 0
				fmt.Sscanf(clientsStr, "%d", &clients)
				if clients > 0 {
					info.Networks[idx].ActiveClients = clients
					info.Networks[idx].VMCount = clients
				}
			}

			// Query detailed portgroup info
			pgDetailCmd := fmt.Sprintf("esxcli network vswitch standard portgroup get -p '%s' 2>/dev/null", network.Network)
			pgDetailOutput, _ := mgr.RunCommand(pgDetailCmd)

			// Parse portgroup details
			if pgDetailOutput != "" {
				info.Networks[idx].Accessible = parsePortGroupDetail(pgDetailOutput, "Accessible") == "Yes"
				vmCountStr := parsePortGroupDetail(pgDetailOutput, "Virtual machines")
				fmt.Sscanf(vmCountStr, "%d", &info.Networks[idx].VMCount)

				activePorts := 0
				portsStr := parsePortGroupDetail(pgDetailOutput, "Active ports")
				fmt.Sscanf(portsStr, "%d", &activePorts)
				if activePorts > 0 {
					info.Networks[idx].ActivePorts = activePorts
				}
			}

			// Query security policy
			securityCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy security get -p '%s' 2>/dev/null", network.Network)
			secOutput, _ := mgr.RunCommand(securityCmd)
			if secOutput != "" {
				info.Networks[idx].Security = &config.SecurityPolicy{
					AllowPromiscuous: parseBoolValue(parseSecurityPolicyField(secOutput, "Promiscuous")),
					AllowForgedTx:    parseBoolValue(parseSecurityPolicyField(secOutput, "Forged")),
					AllowMACChanges:  parseBoolValue(parseSecurityPolicyField(secOutput, "MAC")),
				}
			}

			// Query NIC teaming policy
			teamingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy failover get -p '%s' 2>/dev/null", network.Network)
			teamOutput, _ := mgr.RunCommand(teamingCmd)
			if teamOutput != "" {
				info.Networks[idx].NICTeaming = &config.NICTeamingPolicy{
					NotifySwitches: parseBoolValue(parseTeamingPolicyField(teamOutput, "Notify")),
					Policy:         parseTeamingPolicyField(teamOutput, "Policy"),
					ReversePolicy:  parseBoolValue(parseTeamingPolicyField(teamOutput, "Reverse")),
					Failback:       parseBoolValue(parseTeamingPolicyField(teamOutput, "Failback")),
				}
			}

			// Query shaping policy
			shapingCmd := fmt.Sprintf("esxcli network vswitch standard portgroup policy shaping get -p '%s' 2>/dev/null", network.Network)
			shapOutput, _ := mgr.RunCommand(shapingCmd)
			if shapOutput != "" {
				info.Networks[idx].Shaping = &config.ShapingPolicy{
					Enabled: parseBoolValue(parseShapingPolicyField(shapOutput, "Enabled")),
				}
			}
		}
	}
}

// enrichDatastoreDetailsWithMetadata populates datastore info with comprehensive metadata
func (i *InspectVM) enrichDatastoreDetailsWithMetadata(mgr *utils.SSHManager, info *InspectVMInfo) {
	if len(info.Datastores) == 0 {
		return
	}

	// Get all filesystem info with esxcli
	output, err := mgr.RunCommand("esxcli storage filesystem list 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		common.Debug("enrich datastores", "error", "failed to get filesystems from esxcli")
		return
	}

	// Parse esxcli output - format includes Mount Point, UUID, Type, and more
	// Build a map of mount path to metadata
	filesystemInfo := make(map[string]map[string]string)
	lines := strings.Split(output, "\n")

	var currentMount string
	var currentUUID string
	var currentType string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Mount Point:") {
			currentMount = strings.TrimSpace(strings.TrimPrefix(line, "Mount Point:"))
		} else if strings.HasPrefix(line, "UUID:") {
			currentUUID = strings.TrimSpace(strings.TrimPrefix(line, "UUID:"))
		} else if strings.HasPrefix(line, "Type:") {
			currentType = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
			if currentMount != "" && currentUUID != "" {
				filesystemInfo[currentMount] = map[string]string{
					"uuid": currentUUID,
					"type": currentType,
				}
			}
			// Reset for next entry
			currentMount = ""
			currentUUID = ""
			currentType = ""
		}
	}

	// Enrich each datastore with metadata
	for idx, ds := range info.Datastores {
		// Mark as mounted and set mount point
		info.Datastores[idx].Mounted = true
		info.Datastores[idx].MountPoint = ds.Path
		info.Datastores[idx].Accessible = true

		// Try to match by mount path and get detailed metadata
		if fsInfo, ok := filesystemInfo[ds.Path]; ok {
			if uuid, ok := fsInfo["uuid"]; ok && uuid != "" {
				info.Datastores[idx].UUID = uuid
			}
			if fsType, ok := fsInfo["type"]; ok && fsType != "" {
				info.Datastores[idx].Type = fsType
			}
		}

		// Query detailed datastore info using esxcli
		if ds.UUID != "" {
			detailCmd := fmt.Sprintf("esxcli storage filesystem info -l '%s' 2>/dev/null", ds.UUID)
			detailOutput, _ := mgr.RunCommand(detailCmd)

			if detailOutput != "" {
				// Parse detailed information
				typeVal := parseDatastoreDetail(detailOutput, "Type")
				if typeVal != "" && typeVal != "Unknown" {
					info.Datastores[idx].Type = typeVal
				}

				version := parseDatastoreDetail(detailOutput, "Version")
				if version != "" && version != "Unknown" {
					info.Datastores[idx].Version = version
				}

				localVal := parseDatastoreDetail(detailOutput, "Local")
				if localVal == "Yes" || localVal == "true" {
					info.Datastores[idx].Local = true
				}

				blockSize := parseDatastoreDetail(detailOutput, "Block size")
				if blockSize != "" && blockSize != "Unknown" {
					info.Datastores[idx].BlockSize = blockSize
				}

				hostsStr := parseDatastoreDetail(detailOutput, "Hosts")
				var hostCount int
				if _, err := fmt.Sscanf(hostsStr, "%d", &hostCount); err == nil && hostCount > 0 {
					info.Datastores[idx].HostCount = hostCount
				}

				vmsStr := parseDatastoreDetail(detailOutput, "Virtual Machines")
				var vmCount int
				if _, err := fmt.Sscanf(vmsStr, "%d", &vmCount); err == nil && vmCount > 0 {
					info.Datastores[idx].VMCount = vmCount
				}

				// Parse extents if available
				extentStr := parseDatastoreDetail(detailOutput, "Extent")
				if extentStr != "" && extentStr != "Unknown" {
					info.Datastores[idx].Extents = []string{extentStr}
				}

				accessibleVal := parseDatastoreDetail(detailOutput, "Accessible")
				info.Datastores[idx].Accessible = accessibleVal == "Yes" || accessibleVal == "true"
			}
		}
	}
}

// enrichVSwitchDetailsWithInfrastructure populates vswitch info from the networks' vswitches
func (i *InspectVM) enrichVSwitchDetailsWithInfrastructure(mgr *utils.SSHManager, info *InspectVMInfo) {
	if len(info.Networks) == 0 {
		return
	}

	// Collect unique vswitches from networks
	vswitchSet := make(map[string]bool)
	for _, network := range info.Networks {
		if network.VSwitch != "" {
			vswitchSet[network.VSwitch] = true
		}
	}

	if len(vswitchSet) == 0 {
		return
	}

	// Get details for each vswitch
	for vswitchName := range vswitchSet {
		vswitch := VSwitchInfo{
			Name: vswitchName,
		}

		// Query vswitch details
		detailCmd := fmt.Sprintf("esxcli network vswitch standard list -v '%s' 2>/dev/null", vswitchName)
		detailOutput, _ := mgr.RunCommand(detailCmd)

		if detailOutput != "" {
			vswitch.Type = parseVSwitchDetail(detailOutput, "Type")

			portGroupsStr := parseVSwitchDetail(detailOutput, "Port groups")
			var pgCount int
			if _, err := fmt.Sscanf(portGroupsStr, "%d", &pgCount); err == nil && pgCount > 0 {
				vswitch.PortGroupCount = pgCount
			}

			uplinksStr := parseVSwitchDetail(detailOutput, "Uplinks")
			if uplinksStr != "" && uplinksStr != "Unknown" {
				// uplinks might be comma or space separated
				vswitch.Uplinks = strings.FieldsFunc(uplinksStr, func(r rune) bool {
					return r == ',' || r == ' '
				})
			}

			// Additional details
			mtuStr := parseVSwitchDetail(detailOutput, "MTU")
			var mtu int
			if _, err := fmt.Sscanf(mtuStr, "%d", &mtu); err == nil && mtu > 0 {
				vswitch.MTU = mtu
			}

			portsStr := parseVSwitchDetail(detailOutput, "Ports")
			parsePortsField(portsStr, &vswitch.Ports, &vswitch.AvailablePorts)

			vswitch.LinkDiscovery = parseVSwitchDetail(detailOutput, "Link discovery")

			vmsStr := parseVSwitchDetail(detailOutput, "Attached VMs")
			parseVMsField(vmsStr, &vswitch.AttachedVMs, &vswitch.ActiveVMs)
		}

		// Query vswitch security policy
		securityCmd := fmt.Sprintf("esxcli network vswitch standard policy security get -v '%s' 2>/dev/null", vswitchName)
		secOutput, _ := mgr.RunCommand(securityCmd)
		if secOutput != "" {
			vswitch.Security = &config.SecurityPolicy{
				AllowPromiscuous: parseBoolValue(parseSecurityPolicyField(secOutput, "Promiscuous")),
				AllowForgedTx:    parseBoolValue(parseSecurityPolicyField(secOutput, "Forged")),
				AllowMACChanges:  parseBoolValue(parseSecurityPolicyField(secOutput, "MAC")),
			}
		}

		// Query vswitch NIC teaming policy
		teamingCmd := fmt.Sprintf("esxcli network vswitch standard policy failover get -v '%s' 2>/dev/null", vswitchName)
		teamOutput, _ := mgr.RunCommand(teamingCmd)
		if teamOutput != "" {
			vswitch.NICTeaming = &config.NICTeamingPolicy{
				NotifySwitches: parseBoolValue(parseTeamingPolicyField(teamOutput, "Notify")),
				Policy:         parseTeamingPolicyField(teamOutput, "Policy"),
				ReversePolicy:  parseBoolValue(parseTeamingPolicyField(teamOutput, "Reverse")),
				Failback:       parseBoolValue(parseTeamingPolicyField(teamOutput, "Failback")),
			}
		}

		// Query vswitch shaping policy
		shapingCmd := fmt.Sprintf("esxcli network vswitch standard policy shaping get -v '%s' 2>/dev/null", vswitchName)
		shapOutput, _ := mgr.RunCommand(shapingCmd)
		if shapOutput != "" {
			vswitch.Shaping = &config.ShapingPolicy{
				Enabled: parseBoolValue(parseShapingPolicyField(shapOutput, "Enabled")),
			}
		}

		info.VSwitches = append(info.VSwitches, vswitch)
	}
}

// ============================================================================
// HELPER PARSING FUNCTIONS FOR ENRICHMENT
// ============================================================================

// parsePortGroupDetail extracts a field value from esxcli portgroup get output
func parsePortGroupDetail(output, fieldName string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldName)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseSecurityPolicyField extracts a security policy value
func parseSecurityPolicyField(output, fieldType string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldType)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseTeamingPolicyField extracts a teaming policy value
func parseTeamingPolicyField(output, fieldType string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldType)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseShapingPolicyField extracts a shaping policy value
func parseShapingPolicyField(output, fieldType string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldType)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseBoolValue converts Yes/No or true/false to boolean
func parseBoolValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "yes" || value == "true" || value == "1"
}

// parseDatastoreDetail extracts a field value from esxcli storage filesystem info output
func parseDatastoreDetail(output, fieldName string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldName)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseVSwitchDetail extracts a field value from esxcli vswitch standard list output
func parseVSwitchDetail(output, fieldName string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(fieldName)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parsePortsField parses "Ports: 1536 (1529 available)" format
func parsePortsField(portsStr string, totalPorts, availablePorts *int) {
	// Format: "1536 (1529 available)"
	if portsStr == "" {
		return
	}

	// Extract total ports
	var total int
	if _, err := fmt.Sscanf(portsStr, "%d", &total); err == nil && total > 0 {
		*totalPorts = total
	}

	// Extract available ports from (XXXX available)
	if idx := strings.Index(portsStr, "("); idx != -1 {
		endIdx := strings.Index(portsStr[idx:], ")")
		if endIdx != -1 {
			availStr := portsStr[idx+1 : idx+endIdx]
			var avail int
			if _, err := fmt.Sscanf(availStr, "%d", &avail); err == nil && avail > 0 {
				*availablePorts = avail
			}
		}
	}
}

// parseVMsField parses "X (Y active)" format for attached VMs
func parseVMsField(vmsStr string, totalVMs, activeVMs *int) {
	// Format: "1 (0 active)"
	if vmsStr == "" {
		return
	}

	// Extract total VMs
	var total int
	if _, err := fmt.Sscanf(vmsStr, "%d", &total); err == nil && total > 0 {
		*totalVMs = total
	}

	// Extract active VMs from (X active)
	if idx := strings.Index(vmsStr, "("); idx != -1 {
		endIdx := strings.Index(vmsStr[idx:], ")")
		if endIdx != -1 {
			activeStr := vmsStr[idx+1 : idx+endIdx]
			var active int
			if _, err := fmt.Sscanf(activeStr, "%d", &active); err == nil && active >= 0 {
				*activeVMs = active
			}
		}
	}
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
