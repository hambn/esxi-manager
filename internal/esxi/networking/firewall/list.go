package firewall

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// FirewallRule represents a firewall rule entry
type FirewallRule struct {
	Name           string `json:"name"`
	Key            string `json:"key"`
	IncomingPorts  string `json:"incoming_ports,omitempty"`
	OutgoingPorts  string `json:"outgoing_ports,omitempty"`
	Protocols      string `json:"protocols,omitempty"`
	Service        string `json:"service,omitempty"`
	Daemon         string `json:"daemon,omitempty"`
}

// Friendly names mapping for firewall services
var friendlyNames = map[string]string{
	"sshServer":          "SSH Server",
	"sshClient":          "SSH Client",
	"dhcp":               "DHCP Client",
	"dns":                "DNS",
	"ntpClient":          "NTP Client",
	"nfsClient":          "NFS Client",
	"iSCSI":              "iSCSI",
	"vMotion":            "vMotion",
	"vSphereClient":      "vSphere Client",
	"webAccess":          "Web Access",
	"updateManager":      "Update Manager",
	"faultTolerance":     "Fault Tolerance",
	"NFC":                "NFC",
	"HBR":                "Host Backup and Restore",
	"activeDirectoryAll": "Active Directory",
	"snmp":               "SNMP",
	"CIMHttpServer":      "CIM HTTP Server",
	"CIMHttpsServer":     "CIM HTTPS Server",
	"CIMSLP":             "CIM SLP",
	"vpxHeartbeats":      "vCenter Heartbeats",
	"ftpClient":          "FTP Client",
	"httpClient":         "HTTP Client",
	"gdbserver":          "GDB Server",
	"DVFilter":           "DV Filter",
	"DHCPv6":             "DHCP v6",
	"DVSSync":            "DVS Sync",
	"syslog":             "Syslog",
	"WOL":                "Wake On LAN",
	"vSPC":               "vSphere HA Service",
	"remoteSerialPort":   "Remote Serial Port",
	"rdt":                "Remote Desktop",
	"cmmds":              "CMMDS",
}

// listNetworkingFirewall lists all firewall rules on the ESXi host
func listNetworkingFirewall(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Read service.xml
	output, err := mgr.RunCommand("cat /etc/vmware/firewall/service.xml 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return utils.FormatAsJSON([]FirewallRule{})
	}

	// Parse XML
	type Service struct {
		ID    string `xml:"id"`
		Rules []struct {
			Direction string `xml:"direction"`
			Protocol  string `xml:"protocol"`
			Port      string `xml:"port"`
		} `xml:"rule"`
	}

	type Root struct {
		Services []Service `xml:"service"`
	}

	var root Root
	if err := xml.Unmarshal([]byte(output), &root); err != nil {
		return utils.FormatAsJSON([]FirewallRule{})
	}

	var rules []FirewallRule

	// Process each service
	for _, svc := range root.Services {
		rule := FirewallRule{
			Key:     svc.ID,
			Service: "N/A",
			Daemon:  "None",
		}

		// Get friendly name or use service key
		if friendlyName, ok := friendlyNames[svc.ID]; ok {
			rule.Name = friendlyName
		} else {
			rule.Name = svc.ID
		}

		// Extract ports and protocols
		var inboundPorts, outboundPorts []string
		protocolMap := make(map[string]bool)
		var protocols []string

		for _, r := range svc.Rules {
			// Collect protocols
			proto := strings.ToUpper(r.Protocol)
			if !protocolMap[proto] {
				protocols = append(protocols, proto)
				protocolMap[proto] = true
			}

			// Collect ports
			port := strings.TrimSpace(r.Port)
			if port != "" {
				if r.Direction == "inbound" || r.Direction == "Inbound" {
					inboundPorts = append(inboundPorts, port)
				} else if r.Direction == "outbound" || r.Direction == "Outbound" {
					outboundPorts = append(outboundPorts, port)
				}
			}
		}

		// Set fields
		if len(inboundPorts) > 0 {
			rule.IncomingPorts = strings.Join(inboundPorts, ", ")
		}
		if len(outboundPorts) > 0 {
			rule.OutgoingPorts = strings.Join(outboundPorts, ", ")
		}
		if len(protocols) > 0 {
			rule.Protocols = strings.Join(protocols, ", ")
		}

		rules = append(rules, rule)
	}

	// Return JSON
	jsonData, _ := json.MarshalIndent(rules, "", "  ")
	return string(jsonData), nil
}

func init() {
	config.RegisterFunc("list-networking-firewall", listNetworkingFirewall)
}
