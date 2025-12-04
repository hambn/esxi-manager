package vms

// This file ensures all VM subcommand packages are imported
// which triggers their init() functions for command registration

import (
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/vms/inspect"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/vms/list"
)
