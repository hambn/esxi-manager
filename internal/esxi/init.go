package esxi

// This file ensures all command packages are imported to trigger their init() functions
// which register commands in the central registry

import (
	// Command package imports - triggers auto-registration via init()
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/host"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/networking"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/virtual-machines"
)
