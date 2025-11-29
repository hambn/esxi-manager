package cli

import (
	"testing"
)

func TestParams_ToConfigParams(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		VMName:           "test-vm",
	}

	configParams := params.ToConfigParams()

	if configParams.URI != "192.168.1.100" {
		t.Errorf("expected URI 192.168.1.100, got %s", configParams.URI)
	}
	if configParams.Username != "root" {
		t.Errorf("expected username root, got %s", configParams.Username)
	}
	if configParams.Password != "password" {
		t.Errorf("expected password password, got %s", configParams.Password)
	}
	if configParams.Port != 22 {
		t.Errorf("expected port 22, got %d", configParams.Port)
	}
	if configParams.VMName != "test-vm" {
		t.Errorf("expected VMName test-vm, got %s", configParams.VMName)
	}
}

func TestParams_Validate_Success(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "list-vms",
	}

	err := params.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestParams_Validate_MissingURI(t *testing.T) {
	params := &Params{
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "list-vms",
	}

	err := params.Validate()
	if err == nil {
		t.Fatal("expected error for missing URI")
	}
}

func TestParams_Validate_MissingUsername(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "list-vms",
	}

	err := params.Validate()
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestParams_Validate_MissingPassword(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPort:     22,
		Command:          "list-vms",
	}

	err := params.Validate()
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestParams_Validate_MissingCommand(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := params.Validate()
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestParams_Validate_InvalidPort(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     99999,
		Command:          "list-vms",
	}

	err := params.Validate()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
