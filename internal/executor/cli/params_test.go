package cli

import (
	"testing"
)

func TestParams_ToESXiHost(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	host := params.ToESXiHost()

	if host.URI != "192.168.1.100" {
		t.Errorf("expected URI 192.168.1.100, got %s", host.URI)
	}
	if host.Username != "root" {
		t.Errorf("expected username root, got %s", host.Username)
	}
	if host.Password != "password" {
		t.Errorf("expected password password, got %s", host.Password)
	}
	if host.Port != 22 {
		t.Errorf("expected port 22, got %d", host.Port)
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
