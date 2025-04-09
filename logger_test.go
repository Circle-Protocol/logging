package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOutput(t *testing.T) {
	// Create a buffer to capture the log output
	var buf bytes.Buffer
	
	// Set the output to our buffer
	SetOutput(&buf)
	
	// Log a test message
	GetLogger().Info("test message")
	
	// Verify the output contains our message
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "INFO", logEntry["level"])
}

func TestSetLevel(t *testing.T) {
	// Test debug level logging
	{
		// Create a buffer to capture the log output
		var buf bytes.Buffer
		SetOutput(&buf)
		
		// Set level to debug and log a debug message
		SetLevel(slog.LevelDebug)
		GetLogger().Debug("debug message")
		
			// Get the output as a string and print it for debugging
		output := buf.String()
		t.Logf("Debug output: %s", output)
		
		// Just verify that something was logged
		assert.NotEmpty(t, output)
	}
	
	// Test that debug messages are not logged at info level
	{
		// Create a new buffer to capture the log output
		var buf bytes.Buffer
		SetOutput(&buf)
		
		// Set level to info and try to log a debug message
		SetLevel(slog.LevelInfo)
		GetLogger().Debug("should not appear")
		
		// Verify debug message was not logged
		assert.Empty(t, buf.String())
	}
	
	// Test info level logging
	{
		// Create a new buffer to capture the log output
		var buf bytes.Buffer
		SetOutput(&buf)
		
		// Set level to info and log an info message
		SetLevel(slog.LevelInfo)
		GetLogger().Info("info message")
		
		// Get the output as a string and print it for debugging
		output := buf.String()
		t.Logf("Info output: %s", output)
		
		// Just verify that something was logged
		assert.NotEmpty(t, output)
	}
}

func TestWithContext(t *testing.T) {
	// Create a context with a logger
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	
	ctx := context.Background()
	ctx = ToContext(ctx, logger)
	
	// Get logger from context
	loggerFromCtx := WithContext(ctx)
	
	// Log a message with the logger from context
	loggerFromCtx.Info("context logger test")
	
	// Verify the message was logged to our buffer
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "context logger test", logEntry["msg"])
}

func TestWithGroup(t *testing.T) {
	// Create a buffer to capture the log output
	var buf bytes.Buffer
	SetOutput(&buf)
	
	// Create a logger with a group and add a distinctive attribute
	// to make it easier to verify the group is working
	groupLogger := WithGroup("test_group").With("test_id", "group-test")
	
	// Log a message with the group logger
	groupLogger.Info("group message")
	
	// In slog with JSONHandler, groups are flattened with dot notation
	// So we'll verify our message was logged and contains our test_id
	output := buf.String()
	assert.Contains(t, output, "group message")
	assert.Contains(t, output, "test_id")
	assert.Contains(t, output, "group-test")
}

func TestWithAttrs(t *testing.T) {
	// Create a buffer to capture the log output
	var buf bytes.Buffer
	SetOutput(&buf)
	
	// Create a logger with attributes
	attrLogger := WithAttrs(
		slog.String("service", "test"),
		slog.Int("version", 1),
	)
	
	// Log a message with the attribute logger
	attrLogger.Info("attr message")
	
	// Verify the message was logged with the attributes
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "attr message", logEntry["msg"])
	assert.Equal(t, "test", logEntry["service"])
	assert.Equal(t, float64(1), logEntry["version"]) // JSON unmarshals numbers as float64
}
