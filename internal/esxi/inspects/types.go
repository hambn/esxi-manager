package inspects

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
