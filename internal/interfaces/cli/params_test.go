package cli

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/config"
)

func TestParseFlags_ValidParams(t *testing.T) {
	// Note: Can't easily test ParseFlags without mocking flag.Parse()
	// This would require refactoring flag parsing into a separate package
	// For now, we test the validateParams function directly
}

func TestValidateParams_Success(t *testing.T) {
	params := &config.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "list-vms")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestValidateParams_MissingURI(t *testing.T) {
	params := &config.Params{
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "list-vms")
	if err == nil {
		t.Fatal("expected error for missing URI")
	}
}

func TestValidateParams_MissingUsername(t *testing.T) {
	params := &config.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "list-vms")
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestValidateParams_MissingPassword(t *testing.T) {
	params := &config.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "list-vms")
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestValidateParams_MissingCommand(t *testing.T) {
	params := &config.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "")
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestValidateParams_InvalidPort(t *testing.T) {
	params := &config.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     99999,
	}

	err := validateParams(params, "list-vms")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
