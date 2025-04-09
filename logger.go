package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// loggerKey is the key used to store and retrieve the logger from context
const loggerKey contextKey = "logger"

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

// WithContext returns a logger from the context if present,
// otherwise returns the default logger
func WithContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return defaultLogger
	}
	
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	
	return defaultLogger
}

// ToContext adds the logger to the provided context
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	
	if logger == nil {
		logger = defaultLogger
	}
	
	return context.WithValue(ctx, loggerKey, logger)
}

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
