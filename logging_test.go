package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTest() (*bytes.Buffer, *Entry) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)
	
	// Create a handler that writes to the buffer with Debug level
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	
	// Create a logger with the handler
	logger := slog.New(handler)
	
	// Set the logger as the default and set level to Debug
	SetOutput(buf)
	SetLevel(slog.LevelDebug) // This is crucial for Debug logs to appear
	
	// Create an entry with the logger
	entry := &Entry{
		logger: logger,
		ctx:    context.Background(),
	}
	
	return buf, entry
}

func TestEntryWithField(t *testing.T) {
	buf, entry := setupTest()
	
	// Add a field and log a message
	entry.WithField("key", "value").Info("test message")
	
	// Check that the output contains the field and message
	output := buf.String()
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "test message")
}

func TestEntryWithFields(t *testing.T) {
	buf, entry := setupTest()
	
	// Add multiple fields and log a message
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	entry.WithFields(fields).Info("test message")
	
	// Check that the output contains the fields and message
	output := buf.String()
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "test message")
}

func TestEntryWithError(t *testing.T) {
	buf, entry := setupTest()
	
	// Add an error and log a message
	err := assert.AnError
	entry.WithError(err).Info("test message")
	
	// Check that the output contains the error and message
	output := buf.String()
	assert.Contains(t, output, "error")
	assert.Contains(t, output, err.Error())
	assert.Contains(t, output, "test message")
}

func TestEntryOnError(t *testing.T) {
	buf, entry := setupTest()
	
	// Test with error
	err := assert.AnError
	entry.OnError(err).Info("test message")
	
	// Check that the output contains the error and message
	output := buf.String()
	assert.Contains(t, output, "error")
	assert.Contains(t, output, err.Error())
	assert.Contains(t, output, "test message")
	
	// Clear the buffer
	buf.Reset()
	
	// Test with nil error (should not log)
	entry.OnError(nil).Info("should not appear")
	
	// Check that nothing was logged
	assert.Empty(t, buf.String())
}

func TestEntrySetFields(t *testing.T) {
	buf, entry := setupTest()
	
	// Add fields using SetFields and log a message
	entry.SetFields("key1", "value1", "key2", 42).Info("test message")
	
	// Check that the output contains the fields and message
	output := buf.String()
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "test message")
}

func TestEntrySetFieldsOddArgs(t *testing.T) {
	_, entry := setupTest()
	
	// Test with odd number of arguments (should panic)
	assert.Panics(t, func() {
		entry.SetFields("key1", "value1", "key2")
	})
}

func TestEntrySetFieldsNonStringKey(t *testing.T) {
	_, entry := setupTest()
	
	// Test with non-string key (should panic)
	assert.Panics(t, func() {
		entry.SetFields(42, "value1")
	})
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(*Entry)
		expected string
	}{
		{
			name: "Debug",
			logFunc: func(e *Entry) {
				e.Debug("debug message")
			},
			expected: "DEBUG",
		},
		{
			name: "Info",
			logFunc: func(e *Entry) {
				e.Info("info message")
			},
			expected: "INFO",
		},
		{
			name: "Warn",
			logFunc: func(e *Entry) {
				e.Warn("warn message")
			},
			expected: "WARN",
		},
		{
			name: "Error",
			logFunc: func(e *Entry) {
				e.Error("error message")
			},
			expected: "ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, entry := setupTest()
			
			// Call the log function
			tt.logFunc(entry)
			
			// Check that the output contains the expected level
			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestLogWithContext(t *testing.T) {
	// Create a context with a logger
	buf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	ctx := ToContext(context.Background(), logger)
	
	// Create an entry with the context
	entry := WithContext(ctx)
	
	// Log a message
	entry.Info("test message")
	
	// Check that the output contains the message
	output := buf.String()
	assert.Contains(t, output, "test message")
}

func TestLogWithID(t *testing.T) {
	// Set the ID key
	SetIDKey("requestID")
	
	buf, _ := setupTest()
	
	// Create an entry with an ID
	Log("12345").Info("test message")
	
	// Check that the output contains the ID
	output := buf.String()
	assert.Contains(t, output, "requestID")
	assert.Contains(t, output, "12345")
	assert.Contains(t, output, "test message")
}

func TestLogWithFieldsFunc(t *testing.T) {
	buf, _ := setupTest()
	
	// Create an entry with an ID and fields
	LogWithFields("12345", "key1", "value1", "key2", 42).Info("test message")
	
	// Check that the output contains the ID and fields
	output := buf.String()
	assert.Contains(t, output, idKey)
	assert.Contains(t, output, "12345")
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "test message")
}

func TestGlobalLogFunctions(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Debug",
			logFunc: func() {
				Debug("debug message")
			},
			expected: "DEBUG",
		},
		{
			name: "Debugf",
			logFunc: func() {
				Debugf("debug %s", "formatted")
			},
			expected: "debug formatted",
		},
		{
			name: "Debugln",
			logFunc: func() {
				Debugln("debug", "message")
			},
			expected: "debug message",
		},
		{
			name: "Info",
			logFunc: func() {
				Info("info message")
			},
			expected: "INFO",
		},
		{
			name: "Infof",
			logFunc: func() {
				Infof("info %s", "formatted")
			},
			expected: "info formatted",
		},
		{
			name: "Infoln",
			logFunc: func() {
				Infoln("info", "message")
			},
			expected: "info message",
		},
		{
			name: "Warn",
			logFunc: func() {
				Warn("warn message")
			},
			expected: "WARN",
		},
		{
			name: "Warnf",
			logFunc: func() {
				Warnf("warn %s", "formatted")
			},
			expected: "warn formatted",
		},
		{
			name: "Warnln",
			logFunc: func() {
				Warnln("warn", "message")
			},
			expected: "warn message",
		},
		{
			name: "Warning",
			logFunc: func() {
				Warning("warning message")
			},
			expected: "WARN",
		},
		{
			name: "Error",
			logFunc: func() {
				Error("error message")
			},
			expected: "ERROR",
		},
		{
			name: "Errorf",
			logFunc: func() {
				Errorf("error %s", "formatted")
			},
			expected: "error formatted",
		},
		{
			name: "Errorln",
			logFunc: func() {
				Errorln("error", "message")
			},
			expected: "error message",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, _ := setupTest()
			
			// Call the log function
			tt.logFunc()
			
			// Check that the output contains the expected content
			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestPanicFunctions(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func()
	}{
		{
			name: "Panic",
			logFunc: func() {
				Panic("panic message")
			},
		},
		{
			name: "Panicf",
			logFunc: func() {
				Panicf("panic %s", "formatted")
			},
		},
		{
			name: "Panicln",
			logFunc: func() {
				Panicln("panic", "message")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, _ := setupTest()
			
			// Call the log function (should panic)
			assert.Panics(t, tt.logFunc)
			
			// Check that the output contains ERROR level
			output := buf.String()
			assert.Contains(t, output, "ERROR")
		})
	}
}

func TestFatalFunctions(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func()
	}{
		{
			name: "Fatal",
			logFunc: func() {
				Fatal("fatal message")
			},
		},
		{
			name: "Fatalf",
			logFunc: func() {
				Fatalf("fatal %s", "formatted")
			},
		},
		{
			name: "Fatalln",
			logFunc: func() {
				Fatalln("fatal", "message")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, _ := setupTest()
			
			// Call the log function (should panic)
			func() {
				defer func() {
					r := recover()
					assert.NotNil(t, r, "Expected function to panic")
					
					// Check that the panic message starts with "fatal: "
					if r != nil {
						panicMsg, ok := r.(string)
						assert.True(t, ok, "Expected panic value to be a string")
						assert.Contains(t, panicMsg, "fatal: ", "Expected panic message to contain 'fatal: '")
					}
				}()
				
				// This should panic
				tt.logFunc()
			}()
			
			// Check that the output contains ERROR level
			output := buf.String()
			assert.Contains(t, output, "ERROR")
		})
	}
}

func TestLogFunctions(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected string
	}{
		{
			name:     "Debug",
			level:    slog.LevelDebug,
			expected: "DEBUG",
		},
		{
			name:     "Info",
			level:    slog.LevelInfo,
			expected: "INFO",
		},
		{
			name:     "Warn",
			level:    slog.LevelWarn,
			expected: "WARN",
		},
		{
			name:     "Error",
			level:    slog.LevelError,
			expected: "ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, entry := setupTest()
			
			// Test Log
			entry.Log(tt.level, "test message")
			output := buf.String()
			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, "test message")
			
			// Clear the buffer
			buf.Reset()
			
			// Test Logf
			Logf(tt.level, "test %s", "formatted")
			output = buf.String()
			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, "test formatted")
			
			// Clear the buffer
			buf.Reset()
			
			// Test Logln
			Logln(tt.level, "test", "message")
			output = buf.String()
			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, "test message")
		})
	}
}

func TestCallerInfo(t *testing.T) {
	buf, entry := setupTest()
	
	// Log a message
	entry.Info("test message")
	
	// Check that the output contains caller information
	output := buf.String()
	assert.Contains(t, output, "caller")
	assert.Contains(t, output, "logging_test.go")
}

func TestWithGlobalFunctions(t *testing.T) {
	buf, _ := setupTest()
	
	// Test WithError
	WithError(assert.AnError).Info("test message")
	output := buf.String()
	assert.Contains(t, output, "error")
	assert.Contains(t, output, assert.AnError.Error())
	
	// Clear the buffer
	buf.Reset()
	
	// Test WithFields
	WithFields("key1", "value1", "key2", 42).Info("test message")
	output = buf.String()
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
	
	// Clear the buffer
	buf.Reset()
	
	// Test OnError with error
	OnError(assert.AnError).Info("test message")
	output = buf.String()
	assert.Contains(t, output, "error")
	assert.Contains(t, output, assert.AnError.Error())
	
	// Clear the buffer
	buf.Reset()
	
	// Test OnError with nil (should not log)
	OnError(nil).Info("should not appear")
	assert.Empty(t, buf.String())
}
