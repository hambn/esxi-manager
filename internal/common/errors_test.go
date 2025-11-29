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
