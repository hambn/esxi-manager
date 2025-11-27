package connection

import "errors"

var (
	ErrNotConnected = errors.New("not connected to ESXi host")
	ErrDialFailed   = errors.New("failed to establish SSH connection")
)
