package firewall

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// FirewallService represents a firewall service from service.xml
type FirewallService struct {
	ID    string `xml:"id"`
	Rules []struct {
		Direction string `xml:"direction"`
		Protocol  string `xml:"protocol"`
		PortType  string `xml:"porttype"`
		Port      string `xml:"port"`
	} `xml:"rule"`
	Enabled bool `xml:"enabled"`
}

// Friendly names mapping for firewall services
var friendlyNames = map[string]string{
	"sshServer":             "SSH Server",
	"sshClient":             "SSH Client",
	"dhcp":                  "DHCP Client",
	"dns":                   "DNS",
	"ntpClient":             "NTP Client",
	"nfsClient":             "NFS Client",
	"iSCSI":                 "iSCSI",
	"vMotion":               "vMotion",
	"vSphereClient":         "vSphere Client",
	"webAccess":             "Web Access",
	"updateManager":         "Update Manager",
	"faultTolerance":        "Fault Tolerance",
	"NFC":                   "NFC",
	"HBR":                   "Host Backup and Restore",
	"activeDirectoryAll":    "Active Directory",
	"snmp":                  "SNMP",
	"CIMHttpServer":         "CIM HTTP Server",
	"CIMHttpsServer":        "CIM HTTPS Server",
	"CIMSLP":                "CIM SLP",
	"vpxHeartbeats":         "vCenter Heartbeats",
	"ftpClient":             "FTP Client",
	"httpClient":            "HTTP Client",
	"gdbserver":             "GDB Server",
	"DVFilter":              "DV Filter",
	"DHCPv6":                "DHCP v6",
	"DVSSync":               "DVS Sync",
	"syslog":                "Syslog",
	"WOL":                   "Wake On LAN",
	"vSPC":                  "vSphere HA Servi",
	"remoteSerialPort":      "Remote Serial Port",
	"rdt":                   "Remote Desktop",
	"cmmds":                 "CMMDS",
}

// RuleDetail represents a firewall rule with all details
type RuleDetail struct {
	Name           string `json:"name"`
	Key            string `json:"key"`
	IncomingPorts  string `json:"incoming_ports,omitempty"`
	OutgoingPorts  string `json:"outgoing_ports,omitempty"`
	Protocols      string `json:"protocols,omitempty"`
	Service        string `json:"service,omitempty"`
	Daemon         string `json:"daemon,omitempty"`
}

// listNetworkingFirewall lists all firewall rules on the ESXi host
func listNetworkingFirewall(params *config.Params) (string, error) {
	mgr, err := utils.NewSSHManager(params)
	if err != nil {
		return "", err
	}
	defer mgr.Close()

	// Read service.xml to get firewall rule details
	xmlOutput, err := mgr.RunCommand("cat /etc/vmware/firewall/service.xml 2>/dev/null")
	if err != nil || strings.TrimSpace(xmlOutput) == "" {
		return utils.FormatAsJSON([]RuleDetail{})
	}

	// Parse XML
	type ConfigRoot struct {
		Services []FirewallService `xml:"service"`
	}

	var config ConfigRoot
	if err := xml.Unmarshal([]byte(xmlOutput), &config); err != nil {
		return utils.FormatAsJSON([]RuleDetail{})
	}

	var rules []RuleDetail

	// Process each service
	for _, service := range config.Services {
		rule := RuleDetail{
			Key:     service.ID,
			Service: "N/A",
			Daemon:  "None",
		}

		// Get friendly name or use service key
		if friendlyName, exists := friendlyNames[service.ID]; exists {
			rule.Name = friendlyName
		} else {
			rule.Name = service.ID
		}

		// Extract ports and protocols from rules
		var inboundPorts, outboundPorts []string
		var protocols []string
		protocolMap := make(map[string]bool)

		for _, r := range service.Rules {
			protocol := strings.ToUpper(r.Protocol)
			if !protocolMap[protocol] {
				protocols = append(protocols, protocol)
				protocolMap[protocol] = true
			}

			// Extract port from port element (could be range with begin/end or single value)
			port := strings.TrimSpace(r.Port)
			if port != "" {
				if r.Direction == "inbound" || r.Direction == "Inbound" {
					inboundPorts = append(inboundPorts, port)
				} else if r.Direction == "outbound" || r.Direction == "Outbound" {
					outboundPorts = append(outboundPorts, port)
				}
			}
		}

		// Set ports and protocols
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

	// Convert to JSON manually to ensure proper output
	jsonData, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return utils.FormatAsJSON(rules)
	}

	return string(jsonData), nil
}

func init() {
	config.RegisterFunc("list-networking-firewall", listNetworkingFirewall)
}
