package networking

// This file ensures all networking subcommand packages are imported
// which triggers their init() functions for command registration

import (
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/vswitch"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/portgroup"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/network"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking/vmknics"
)
