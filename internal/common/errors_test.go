package common

import (
	"errors"
	"testing"
)

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := WrapError(originalErr, "context")

	if wrapped == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	if !errors.Is(wrapped, originalErr) {
		t.Fatal("wrapped error should contain original error")
	}

	expectedMsg := "context: original error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, wrapped.Error())
	}
}

func TestWrapError_NilError(t *testing.T) {
	wrapped := WrapError(nil, "context")
	if wrapped != nil {
		t.Fatal("wrapping nil error should return nil")
	}
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := WrapErrorf(originalErr, "operation %s failed", "test")

	if wrapped == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	if !errors.Is(wrapped, originalErr) {
		t.Fatal("wrapped error should contain original error")
	}

	expectedMsg := "operation test failed: original error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, wrapped.Error())
	}
}

func TestIs(t *testing.T) {
	err := WrapError(ErrInvalidConfig, "validation")

	if !Is(err, ErrInvalidConfig) {
		t.Fatal("Is should find wrapped error type")
	}

	if Is(err, ErrNotFound) {
		t.Fatal("Is should not match different error types")
	}
}

func TestAs(t *testing.T) {
	customErr := &testError{msg: "test"}
	wrapped := WrapError(customErr, "context")

	var target *testError
	if !As(wrapped, &target) {
		t.Fatal("As should find custom error type")
	}

	if target.msg != "test" {
		t.Errorf("expected msg 'test', got %q", target.msg)
	}
}

// Test helper type
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
