package config

import (
	"testing"
)

func TestESXiHost_Validate_Success(t *testing.T) {
	host := &ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
		Port:     22,
	}

	err := host.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestESXiHost_Validate_MissingURI(t *testing.T) {
	host := &ESXiHost{
		Username: "root",
		Password: "password",
	}

	err := host.Validate()
	if err != ErrMissingURI {
		t.Fatalf("expected ErrMissingURI, got %v", err)
	}
}

func TestESXiHost_Validate_MissingUsername(t *testing.T) {
	host := &ESXiHost{
		URI:      "192.168.1.100",
		Password: "password",
	}

	err := host.Validate()
	if err != ErrMissingUsername {
		t.Fatalf("expected ErrMissingUsername, got %v", err)
	}
}

func TestESXiHost_Validate_MissingPassword(t *testing.T) {
	host := &ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
	}

	err := host.Validate()
	if err != ErrMissingPassword {
		t.Fatalf("expected ErrMissingPassword, got %v", err)
	}
}

func TestESXiHost_DefaultPort(t *testing.T) {
	host := &ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
		Port:     0,
	}

	err := host.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if host.Port != 22 {
		t.Fatalf("expected default port 22, got %d", host.Port)
	}
}
