package logging

import (
	"io"
	"log/slog"
	"os"
)

// This file contains the core logger functionality

// Default logger instance with JSON handler
var defaultOutput io.Writer = os.Stdout
var defaultLevel = slog.LevelInfo

var defaultLogger = slog.New(slog.NewJSONHandler(defaultOutput, &slog.HandlerOptions{
	Level: defaultLevel,
}))

// SetOutput configures the output destination for the default logger
// while preserving the existing log level
func SetOutput(out io.Writer) {
	// Update the default output
	defaultOutput = out

	// Create a new handler with the current level but new output
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: defaultLevel,
	})

	// Replace the default logger
	defaultLogger = slog.New(handler)
}

// SetLevel configures the minimum log level for the default logger
// while preserving the existing output writer
func SetLevel(level slog.Level) {
	// Update the default level
	defaultLevel = level

	// Create a new handler with the current output but new level
	handler := slog.NewJSONHandler(defaultOutput, &slog.HandlerOptions{
		Level: level,
	})

	// Replace the default logger
	defaultLogger = slog.New(handler)
}

// GetLogger returns the default logger instance
func GetLogger() *slog.Logger {
	return defaultLogger
}

// Context-related functions have been moved to context.go

// WithGroup returns a new logger with the specified group name
func WithGroup(name string) *slog.Logger {
	return defaultLogger.WithGroup(name)
}

// WithAttrs returns a new logger with the specified attributes
func WithAttrs(attrs ...slog.Attr) *slog.Logger {
	// Convert []slog.Attr to []any for compatibility with slog.With
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	return defaultLogger.With(anyAttrs...)
}
