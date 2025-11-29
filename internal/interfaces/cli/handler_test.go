package cli

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/esxi"
)

// ValidateConnection Tests

func TestParamsValidateConnection_Success(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := params.ValidateConnection()
	if err != nil {
		t.Fatalf("ValidateConnection failed: %v", err)
	}
}

func TestParamsValidateConnection_MissingURI(t *testing.T) {
	params := &esxi.Params{
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := params.ValidateConnection()
	if err == nil {
		t.Fatal("expected error for missing URI")
	}
}

func TestParamsValidateConnection_MissingUsername(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := params.ValidateConnection()
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestParamsValidateConnection_MissingPassword(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPort:     22,
	}

	err := params.ValidateConnection()
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestParamsValidateConnection_InvalidPort(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     99999,
	}

	err := params.ValidateConnection()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestParamsValidateConnection_PortZero(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     0,
	}

	err := params.ValidateConnection()
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

// Execute Tests

func TestExecute_DispatchError(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	// Try to execute unknown command
	err := Execute("unknown-command", params)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
