package config

import (
	"flag"
	"sync"
)

// Params holds all application parameters - both connection and command parameters unified in one place
type Params struct {
	// ESXi Host Connection Parameters
	// These are required for all commands to establish SSH connectivity
	ESXiHostURI      string
	ESXiHostUsername string
	ESXiHostPassword string
	ESXiHostPort     int

	// Virtual Machine Parameters
	// Used by VM-related commands (clone, create, delete, list)
	VMName        string
	SourceVMName  string
	SourceVMID    string
	DestVMName    string
	DestDiskStore string
	DestRAM       int
	DestCPU       int
	DestNetwork   string

	// VM Inspect Parameters
	// Used by vm-inspect command to show all details about a specific VM
	VMInspectID   string
	VMInspectName string

	// Network Parameters
	// Used by networking commands (vswitch, portgroup management)
	VSwitchName   string
	PortgroupName string
	VLAN          int
	MTU           int
	Uplinks       []string

	// Storage Parameters
	// Used by storage commands (datastore operations)
	DatastoreName string
}

// CommandInterface defines the contract that all commands must implement
// Each command must be able to validate its specific parameters and execute its operation
type CommandInterface interface {
	Validate() error
	Execute() error
}

// CommandFactory is a factory function that creates a command instance
// It receives the unified Params and returns a CommandInterface ready to execute
type CommandFactory func(params *Params) CommandInterface

// CommandRegistry manages all registered commands with thread-safe access
// Commands auto-register themselves via init() functions in their packages
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]CommandFactory
}

// RegisterFlags registers all parameter flags with the flag package
// This centralizes all CLI flag definitions in one place
func (p *Params) RegisterFlags() {
	// ESXi Host Connection Flags
	flag.StringVar(&p.ESXiHostURI, "esxi-host-uri", "", "ESXi host URI/IP address")
	flag.StringVar(&p.ESXiHostUsername, "esxi-host-username", "", "ESXi host username")
	flag.StringVar(&p.ESXiHostPassword, "esxi-host-password", "", "ESXi host password")
	flag.IntVar(&p.ESXiHostPort, "esxi-host-port", 22, "ESXi host SSH port")

	// Virtual Machine Operation Flags
	flag.StringVar(&p.VMName, "vm-name", "", "Virtual machine name")
	flag.StringVar(&p.SourceVMName, "source-vm-name", "", "Source VM name for cloning")
	flag.StringVar(&p.SourceVMID, "source-vm-id", "", "Source VM ID for cloning")
	flag.StringVar(&p.DestVMName, "dest-vm-name", "", "Destination VM name")
	flag.StringVar(&p.DestDiskStore, "dest-vm-disk-store", "", "Destination datastore for VM")
	flag.IntVar(&p.DestRAM, "dest-vm-ram", 0, "Destination VM RAM in MB")
	flag.IntVar(&p.DestCPU, "dest-vm-cpu", 0, "Destination VM CPU count")
	flag.StringVar(&p.DestNetwork, "dest-vm-network", "", "Destination VM network/portgroup")

	// VM Inspect Flags
	flag.StringVar(&p.VMInspectID, "vm-inspect-id", "", "VM ID to inspect")
	flag.StringVar(&p.VMInspectName, "vm-inspect-name", "", "VM name to inspect")

	// Network Operation Flags
	flag.StringVar(&p.VSwitchName, "vswitch-name", "", "Virtual switch name")
	flag.StringVar(&p.PortgroupName, "portgroup-name", "", "Port group name")
	flag.IntVar(&p.VLAN, "vlan", 0, "VLAN ID")
	flag.IntVar(&p.MTU, "mtu", 1500, "MTU size")

	// Storage Operation Flags
	flag.StringVar(&p.DatastoreName, "datastore-name", "", "Datastore name")
}

// ============================================================================
// INSPECT DATA TYPES - Shared across all inspect commands
// ============================================================================

// SecurityPolicy holds security settings for port groups and vswitches
type SecurityPolicy struct {
	AllowPromiscuous bool `json:"allow_promiscuous_mode,omitempty"`
	AllowForgedTx    bool `json:"allow_forged_transmits,omitempty"`
	AllowMACChanges  bool `json:"allow_mac_changes,omitempty"`
}

// NICTeamingPolicy holds NIC teaming/failover settings
type NICTeamingPolicy struct {
	NotifySwitches bool   `json:"notify_switches,omitempty"`
	Policy         string `json:"policy,omitempty"`
	ReversePolicy  bool   `json:"reverse_policy,omitempty"`
	Failback       bool   `json:"failback,omitempty"`
}

// ShapingPolicy holds traffic shaping settings
type ShapingPolicy struct {
	Enabled           bool   `json:"enabled,omitempty"`
	AverageBandwidth  int64  `json:"average_bandwidth,omitempty"`
	PeakBandwidth     int64  `json:"peak_bandwidth,omitempty"`
	BurstSize         int64  `json:"burst_size,omitempty"`
}

// DatastoreInfo holds datastore information with comprehensive VMFS/NFS metadata
type DatastoreInfo struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Type         string  `json:"type"`
	Capacity     int64   `json:"capacity_bytes"`
	FreeSpace    int64   `json:"free_space_bytes"`
	UsedSpace    int64   `json:"used_space_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	UUID         string  `json:"uuid,omitempty"`
	MountPoint   string  `json:"mount_point,omitempty"`
	Version      string  `json:"version,omitempty"`
	Local        bool    `json:"local,omitempty"`
	BlockSize    string  `json:"block_size,omitempty"`
	HostCount    int     `json:"host_count,omitempty"`
	VMCount      int     `json:"vm_count,omitempty"`
	Mounted      bool    `json:"mounted,omitempty"`
	Accessible   bool    `json:"accessible,omitempty"`
	Extents      []string `json:"extents,omitempty"`
	URL          string  `json:"url,omitempty"`
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
	Security       *SecurityPolicy   `json:"security_policy,omitempty"`
	NICTeaming     *NICTeamingPolicy `json:"nic_teaming_policy,omitempty"`
	Shaping        *ShapingPolicy    `json:"shaping_policy,omitempty"`
}

// PortGroupInfo holds port group information
type PortGroupInfo struct {
	Name           string            `json:"name"`
	VSwitch        string            `json:"vswitch,omitempty"`
	VLANID         int               `json:"vlan_id,omitempty"`
	ActiveClients  int               `json:"active_clients,omitempty"`
	Accessible     bool              `json:"accessible,omitempty"`
	VMCount        int               `json:"vm_count,omitempty"`
	ActivePorts    int               `json:"active_ports,omitempty"`
	Security       *SecurityPolicy   `json:"security_policy,omitempty"`
	NICTeaming     *NICTeamingPolicy `json:"nic_teaming_policy,omitempty"`
	Shaping        *ShapingPolicy    `json:"shaping_policy,omitempty"`
}

// NetworkAdapterInfo holds NIC information with detailed vswitch/portgroup details
type NetworkAdapterInfo struct {
	Index          int              `json:"index"`
	Name           string           `json:"name"`
	MacAddress     string           `json:"mac_address"`
	Network        string           `json:"network"`
	Connected      bool             `json:"connected"`
	VSwitch        string           `json:"vswitch,omitempty"`
	VLANID         int              `json:"vlan_id,omitempty"`
	ActiveClients  int              `json:"active_clients,omitempty"`
	Accessible     bool             `json:"accessible,omitempty"`
	VMCount        int              `json:"vm_count,omitempty"`
	ActivePorts    int              `json:"active_ports,omitempty"`
	Security       *SecurityPolicy  `json:"security_policy,omitempty"`
	NICTeaming     *NICTeamingPolicy `json:"nic_teaming_policy,omitempty"`
	Shaping        *ShapingPolicy   `json:"shaping_policy,omitempty"`
}

// HardwareInfo contains hardware specifications
type HardwareInfo struct {
	CPUs              int `json:"cpus"`
	Memory            int `json:"memory_mb"`
	MaxCPUs           int `json:"max_cpus,omitempty"`
	MaxMemory         int `json:"max_memory_mb,omitempty"`
	BootDelay         int `json:"boot_delay_ms,omitempty"`
	CoresPerSocket    int `json:"cores_per_socket,omitempty"`
	SimultaneousThreads int `json:"simultaneous_threads,omitempty"`
	MotherboardLayout string `json:"motherboard_layout,omitempty"`
}

// DiskInfo holds disk information
type DiskInfo struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size_bytes"`
	Datastore  string `json:"datastore"`
	Controller string `json:"controller"`
	DeviceType string `json:"device_type"`
	Filename   string `json:"filename"`
}

// SnapshotInfo holds snapshot information
type SnapshotInfo struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	CreateTime  string   `json:"create_time,omitempty"`
	State       string   `json:"state,omitempty"`
	ParentKey   string   `json:"parent_key,omitempty"`
	ChildKeys   []string `json:"child_keys,omitempty"`
}

// VMDKInfo holds information about a VMDK file
type VMDKInfo struct {
	Filename  string            `json:"filename"`
	Capacity  int64             `json:"capacity_sectors"`
	Extents   []ExtentInfo      `json:"extents,omitempty"`
	DDB       map[string]string `json:"ddb,omitempty"`
}

// ExtentInfo holds information about a VMDK extent
type ExtentInfo struct {
	Type    string `json:"type"`
	Sectors int64  `json:"sectors"`
	Path    string `json:"path"`
}

// ============================================================================
// STORAGE TYPES
// ============================================================================

// StorageAdapterInfo holds storage adapter information
type StorageAdapterInfo struct {
	Name        string `json:"name"`
	Driver      string `json:"driver,omitempty"`
	Type        string `json:"type,omitempty"`
	Model       string `json:"model,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	Status      string `json:"status,omitempty"`
	PathCount   int    `json:"path_count,omitempty"`
	Queue       int    `json:"queue_length,omitempty"`
	Bus         string `json:"bus,omitempty"`
	Slot        string `json:"slot,omitempty"`
	Target      string `json:"target,omitempty"`
	LUN         string `json:"lun,omitempty"`
}

// StorageDeviceInfo holds storage device/disk information
type StorageDeviceInfo struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name,omitempty"`
	Size              int64  `json:"size_bytes,omitempty"`
	SizeGB            string `json:"size_gb,omitempty"`
	Vendor            string `json:"vendor,omitempty"`
	Model             string `json:"model,omitempty"`
	SerialNumber      string `json:"serial_number,omitempty"`
	Status            string `json:"status,omitempty"`
	Type              string `json:"type,omitempty"` // SSD, HDD, Unknown
	LocalDisk         bool   `json:"local_disk,omitempty"`
	PhysicalLocation  string `json:"physical_location,omitempty"`
	ScsiLevel         int    `json:"scsi_level,omitempty"`
	MaxQueueDepth     int    `json:"max_queue_depth,omitempty"`
	SectorSize        int    `json:"sector_size,omitempty"`
	LUN               string `json:"lun,omitempty"`
	Adapter           string `json:"adapter,omitempty"`
	DevfsPath         string `json:"devfs_path,omitempty"`
}

// NVDimmInfo holds NVDIMM (Non-Volatile DIMM) information
type NVDimmInfo struct {
	DimmID            string `json:"dimm_id"`
	Health            string `json:"health,omitempty"`
	HealthStatus      string `json:"health_status,omitempty"`
	Capacity          int64  `json:"capacity_bytes,omitempty"`
	CapacityGB        string `json:"capacity_gb,omitempty"`
	Location          string `json:"location,omitempty"`
	ProductName       string `json:"product_name,omitempty"`
	ManufacturerID    string `json:"manufacturer_id,omitempty"`
	FirmwareVersion   string `json:"firmware_version,omitempty"`
	VoltaileSize      int64  `json:"volatile_size_bytes,omitempty"`
	PersistentSize    int64  `json:"persistent_size_bytes,omitempty"`
	ActionRequired    bool   `json:"action_required,omitempty"`
	Temperature       int    `json:"temperature_celsius,omitempty"`
	PowerLoss         bool   `json:"power_loss_protection,omitempty"`
}

// ============================================================================
// NETWORKING TYPES
// ============================================================================

// VMKnicInfo holds virtual machine kernel NIC information
type VMKnicInfo struct {
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled,omitempty"`
	Portgroup      string `json:"portgroup,omitempty"`
	MAC            string `json:"mac_address,omitempty"`
	IPv4Address    string `json:"ipv4_address,omitempty"`
	IPv4Netmask    string `json:"ipv4_netmask,omitempty"`
	IPv4Gateway    string `json:"ipv4_gateway,omitempty"`
	IPv6Address    string `json:"ipv6_address,omitempty"`
	IPv6Prefix     string `json:"ipv6_prefix,omitempty"`
	IPv4DHCP       bool   `json:"ipv4_dhcp,omitempty"`
	IPv6DHCP       bool   `json:"ipv6_dhcp,omitempty"`
	MTU            int    `json:"mtu,omitempty"`
	Speed          string `json:"speed,omitempty"`
	Duplex         string `json:"duplex,omitempty"`
	LinkStatus     string `json:"link_status,omitempty"`
	VLANID         int    `json:"vlan_id,omitempty"`
	VSwitch        string `json:"vswitch,omitempty"`
}

// NetworkStackInfo holds network stack information
type NetworkStackInfo struct {
	Name               string `json:"name"`
	Instance           int    `json:"instance,omitempty"`
	MaximumMTU         int    `json:"maximum_mtu,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
	VMotionEnabled     bool   `json:"vmotion_enabled,omitempty"`
	ProvisioningEnabled bool  `json:"provisioning_enabled,omitempty"`
	DNSResolver        string `json:"dns_resolver,omitempty"`
	IPRouting          bool   `json:"ip_routing_enabled,omitempty"`
}

// FirewallRuleInfo holds individual firewall rule information
type FirewallRuleInfo struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled,omitempty"`
	Inbound     bool   `json:"inbound,omitempty"`
	Outbound    bool   `json:"outbound,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Direction   string `json:"direction,omitempty"`
	PortType    string `json:"port_type,omitempty"`
	AllPorts    bool   `json:"all_ports,omitempty"`
	BeginPort   int    `json:"begin_port,omitempty"`
	EndPort     int    `json:"end_port,omitempty"`
	SourceIP    string `json:"source_ip,omitempty"`
	DestIP      string `json:"dest_ip,omitempty"`
	SourceMask  string `json:"source_mask,omitempty"`
	DestMask    string `json:"dest_mask,omitempty"`
}

// FirewallInfo holds firewall configuration information
type FirewallInfo struct {
	Name               string `json:"name"`
	DefaultInbound     string `json:"default_inbound,omitempty"` // allow, deny, etc.
	DefaultOutbound    string `json:"default_outbound,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
	LoadedRuleCount    int    `json:"loaded_rule_count,omitempty"`
	EnabledRuleCount   int    `json:"enabled_rule_count,omitempty"`
	Rules              []FirewallRuleInfo `json:"rules,omitempty"`
}

