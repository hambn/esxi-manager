package common

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := Logger()
	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}
}

func TestLoggerSingleton(t *testing.T) {
	logger1 := Logger()
	logger2 := Logger()

	if logger1 != logger2 {
		t.Fatal("Logger should return same instance (singleton)")
	}
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	logger := WithContext(ctx)

	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}
}

func TestDebugLogging(t *testing.T) {
	// Debug logging should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Debug should not panic: %v", r)
		}
	}()
	Debug("test debug message", "key", "value")
}

func TestInfoLogging(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Info should not panic: %v", r)
		}
	}()
	Info("test info message", "key", "value")
}

func TestWarnLogging(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Warn should not panic: %v", r)
		}
	}()
	Warn("test warn message", "key", "value")
}

func TestErrorLogging(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Error should not panic: %v", r)
		}
	}()
	Error("test error message", "key", "value")
}
