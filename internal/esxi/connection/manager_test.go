package connection

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/config"
)

func TestNewManager_ValidConfig(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
		Port:     22,
	}

	manager, err := NewManager(host)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.host != host {
		t.Fatal("manager host not set correctly")
	}
}

func TestNewManager_MissingURI(t *testing.T) {
	host := &config.ESXiHost{
		Username: "root",
		Password: "password",
	}

	_, err := NewManager(host)
	if err == nil {
		t.Fatal("expected error for missing URI")
	}
}

func TestNewManager_MissingUsername(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Password: "password",
	}

	_, err := NewManager(host)
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestNewManager_MissingPassword(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
	}

	_, err := NewManager(host)
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestManagerDefaultPort(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
	}

	// Validate should set default port
	err := host.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if host.Port != 22 {
		t.Fatalf("expected default port 22, got %d", host.Port)
	}
}

func TestManager_IsConnected_NoConnection(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
		Port:     22,
	}

	manager, err := NewManager(host)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager.IsConnected() {
		t.Fatal("expected IsConnected to return false for unconnected manager")
	}
}

func TestManager_Close_NoConnection(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "192.168.1.100",
		Username: "root",
		Password: "password",
		Port:     22,
	}

	manager, err := NewManager(host)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	err = manager.Close()
	if err != nil {
		t.Fatalf("Close should not error on unconnected manager: %v", err)
	}
}

func TestManager_GetClient_NoConnection(t *testing.T) {
	host := &config.ESXiHost{
		URI:      "invalid.host",
		Username: "root",
		Password: "password",
		Port:     22,
	}

	manager, err := NewManager(host)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Trying to get client should fail gracefully
	_, err = manager.GetClient()
	if err == nil {
		t.Fatal("expected error when connecting to invalid host")
	}
}
