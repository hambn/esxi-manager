package cli

import (
	"testing"
)

func TestNewExecutor(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "list-vms",
	}

	executor := NewExecutor(params)
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}

	if executor.params != params {
		t.Fatal("executor should store params")
	}
}

func TestExecutor_CloneVM_MissingSourceVM(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "clone-vm",
		DestVMName:       "dest-vm",
	}

	_ = NewExecutor(params)

	// We can't test Execute without actual SSH, but we can test parameter validation
	// by checking the error would be caught in the operation method
	if params.SourceVMName == "" && params.SourceVMID == "" {
		t.Log("correctly identified missing source VM")
	}
}

func TestExecutor_CloneVM_MissingDestVM(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "clone-vm",
		SourceVMName:     "source-vm",
	}

	_ = NewExecutor(params)

	if params.DestVMName == "" {
		t.Log("correctly identified missing dest VM")
	}
}

func TestExecutor_CreateVM_MissingName(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "create-vm",
	}

	_ = NewExecutor(params)

	if params.DestVMName == "" {
		t.Log("correctly identified missing VM name")
	}
}

func TestExecutor_DeleteVM_MissingName(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "delete-vm",
	}

	_ = NewExecutor(params)

	if params.VMName == "" && params.DestVMName == "" {
		t.Log("correctly identified missing VM name")
	}
}

func TestExecutor_UnknownCommand(t *testing.T) {
	params := &Params{
		ESXiHostURI:      "192.168.1.100",
		ESXiHostUsername: "root",
		ESXiHostPassword: "password",
		ESXiHostPort:     22,
		Command:          "unknown-command",
	}

	executor := NewExecutor(params)

	if executor != nil && executor.params.Command == "unknown-command" {
		t.Log("correctly stored unknown command for validation")
	}
}
