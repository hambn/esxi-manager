package common

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

var (
	loggerOnce sync.Once
	logger     *slog.Logger
)

// Logger returns the global logger instance
func Logger() *slog.Logger {
	loggerOnce.Do(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	})
	return logger
}

// WithContext returns the logger (context is used for cancellation semantics elsewhere)
func WithContext(ctx context.Context) *slog.Logger {
	return Logger()
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}

// Fatalf logs a fatal error and exits
func Fatalf(msg string, args ...any) {
	Logger().Error(msg, args...)
	os.Exit(1)
}
