package storage

// This file ensures all storage subcommand packages are imported
// which triggers their init() functions for command registration

import (
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage/adapters"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage/datastores"
	_ "github.com/esxi-manager/esxi-manager/internal/esxi/storage/devices"
)
