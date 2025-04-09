package logging

import (
	"context"

	"log/slog"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// loggerKey is the key used to store and retrieve the logger from context
const loggerKey contextKey = "logger"

// FromContext takes a Logger from the context, if it was
// previously set by ToContext
func FromContext(ctx context.Context) (logger *slog.Logger, ok bool) {
	if ctx == nil {
		return nil, false
	}

	logger, ok = ctx.Value(loggerKey).(*slog.Logger)
	return logger, ok
}

// ToContext sets a Logger to the context.
// If ctx is nil, a new background context is created.
// If logger is nil, the default logger is used.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if logger == nil {
		logger = GetLogger()
	}

	return context.WithValue(ctx, loggerKey, logger)
}

// WithContext returns a logger from the context if present,
// otherwise returns the default logger
func WithContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return GetLogger()
	}

	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}

	return GetLogger()
}
