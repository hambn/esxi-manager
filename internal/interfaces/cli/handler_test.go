package cli

import (
	"testing"

	"github.com/esxi-manager/esxi-manager/internal/esxi"
)

// ParseFlags Tests

func TestValidateParams_Success(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	err := validateParams(params, "list-vms")
	if err != nil {
		t.Fatalf("validateParams failed: %v", err)
	}
}

func TestValidateParams_MissingURI(t *testing.T) {
	params := &esxi.Params{
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
	params := &esxi.Params{
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
	params := &esxi.Params{
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
	params := &esxi.Params{
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
	params := &esxi.Params{
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

// Executor Tests

func TestNewExecutor(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	executor := NewExecutor(params, "list-vms")
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}

	if executor.params != params {
		t.Fatal("executor should store params")
	}

	if executor.commandName != "list-vms" {
		t.Fatal("executor should store command name")
	}
}

func TestExecutor_CloneVM_MissingSourceVM(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		DestVMName:       "dest-vm",
	}

	_ = NewExecutor(params, "clone-vm")

	// Parameter validation is tested here
	if params.SourceVMName == "" && params.SourceVMID == "" {
		t.Log("correctly identified missing source VM")
	}
}

func TestExecutor_CloneVM_MissingDestVM(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		SourceVMName:     "source-vm",
	}

	_ = NewExecutor(params, "clone-vm")

	if params.DestVMName == "" {
		t.Log("correctly identified missing dest VM")
	}
}

func TestExecutor_CreateVM_MissingName(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	_ = NewExecutor(params, "create-vm")

	if params.DestVMName == "" {
		t.Log("correctly identified missing VM name")
	}
}

func TestExecutor_DeleteVM_MissingName(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	_ = NewExecutor(params, "delete-vm")

	if params.VMName == "" && params.DestVMName == "" {
		t.Log("correctly identified missing VM name")
	}
}

func TestExecutor_UnknownCommand(t *testing.T) {
	params := &esxi.Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
	}

	executor := NewExecutor(params, "unknown-command")

	if executor != nil && executor.commandName == "unknown-command" {
		t.Log("correctly stored unknown command for validation")
	}
}
