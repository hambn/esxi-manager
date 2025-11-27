package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

var (
	loggerOnce sync.Once
	logger     *slog.Logger
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorDim    = "\033[2m"
)

// ColoredHandler is a custom slog handler that outputs colored text
type ColoredHandler struct {
	level slog.Level
}

// Handle implements slog.Handler
func (h *ColoredHandler) Handle(ctx context.Context, record slog.Record) error {
	// Get color based on level
	levelColor := colorGreen
	levelStr := "INFO"
	switch record.Level {
	case slog.LevelDebug:
		levelColor = colorCyan
		levelStr = "DEBUG"
	case slog.LevelWarn:
		levelColor = colorYellow
		levelStr = "WARN"
	case slog.LevelError:
		levelColor = colorRed
		levelStr = "ERROR"
	}

	// Format time
	timeStr := record.Time.Format("2006-01-02 15:04:05")

	// Build attributes string
	attrs := ""
	record.Attrs(func(a slog.Attr) bool {
		val := a.Value.Any()
		attrs += fmt.Sprintf(" %s=%v", a.Key, val)
		return true
	})

	// Output format: [TIME COLORED_LEVEL] message attributes
	fmt.Printf("[%s%s%s %s%s%s] %s%s\n",
		colorDim, timeStr, colorReset,
		levelColor, levelStr, colorReset,
		record.Message, attrs)

	return nil
}

// WithAttrs implements slog.Handler
func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler
func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

// Enabled implements slog.Handler
func (h *ColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Logger returns the global logger instance
func Logger() *slog.Logger {
	loggerOnce.Do(func() {
		logger = slog.New(&ColoredHandler{
			level: slog.LevelInfo,
		})
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
