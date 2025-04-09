package logging

import (
	"context"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	// Test with nil context
	got, ok := FromContext(nil)
	assert.False(t, ok)
	assert.Nil(t, got)

	// Test with empty context
	got, ok = FromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)

	// Test with logger in context
	want := slog.Default()
	ctx := ToContext(context.Background(), want)
	got, ok = FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}

func TestToContext(t *testing.T) {
	// Test with nil context
	logger := slog.Default()
	ctx := ToContext(nil, logger)
	assert.NotNil(t, ctx)

	got, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, logger, got)

	// Test with nil logger
	ctx = ToContext(context.Background(), nil)
	assert.NotNil(t, ctx)

	got, ok = FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, GetLogger(), got)
}

func TestWithContextLogger(t *testing.T) {
	// Test with nil context
	logger := WithContext(nil)
	assert.Equal(t, GetLogger(), logger)

	// Test with empty context
	logger = WithContext(context.Background())
	assert.Equal(t, GetLogger(), logger)

	// Test with logger in context
	want := slog.Default()
	ctx := ToContext(context.Background(), want)
	logger = WithContext(ctx)
	assert.Equal(t, want, logger)
}
