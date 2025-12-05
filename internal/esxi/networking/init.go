package networking

// This file ensures all networking subcommand packages are imported
// which triggers their init() functions for command registration

import (
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/adapters"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/firewall"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/netstacks"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/portgroups"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/vmknics"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/vswitches"
)
