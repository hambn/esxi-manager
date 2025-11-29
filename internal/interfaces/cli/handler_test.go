package cli

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/esxi"
)

// ValidateConnectionParams Tests

func TestValidateConnectionParams_Success(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := esxi.ValidateConnectionParams(params)
	if err != nil {
		t.Fatalf("ValidateConnectionParams failed: %v", err)
	}
}

func TestValidateConnectionParams_MissingURI(t *testing.T) {
	params := &esxi.Params{
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := esxi.ValidateConnectionParams(params)
	if err == nil {
		t.Fatal("expected error for missing URI")
	}
	if !esxi.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestValidateConnectionParams_MissingUsername(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := esxi.ValidateConnectionParams(params)
	if err == nil {
		t.Fatal("expected error for missing username")
	}
	if !esxi.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestValidateConnectionParams_MissingPassword(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPort:     22,
	}

	err := esxi.ValidateConnectionParams(params)
	if err == nil {
		t.Fatal("expected error for missing password")
	}
	if !esxi.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestValidateConnectionParams_InvalidPort(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     99999,
	}

	err := esxi.ValidateConnectionParams(params)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !esxi.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestValidateConnectionParams_PortZero(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     0,
	}

	err := esxi.ValidateConnectionParams(params)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
	if !esxi.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
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
