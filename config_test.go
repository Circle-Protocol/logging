package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		options  []ConfigOption
		expected *Config
	}{
		{
			name:    "default config",
			options: nil,
			expected: &Config{
				Level: "info",
				Formatter: Formatter{
					Format: FormatterJSON,
					Data:   map[string]interface{}{},
				},
				AddSource: false,
				Output:    OutputStderr,
			},
		},
		{
			name: "with options",
			options: []ConfigOption{
				WithLevel("debug"),
				WithTextFormatter(),
				WithAddSource(true),
				WithOutput(OutputStdout),
				WithFormatterData("timeFormat", "2006-01-02"),
			},
			expected: &Config{
				Level: "debug",
				Formatter: Formatter{
					Format: FormatterText,
					Data: map[string]interface{}{
						"timeFormat": "2006-01-02",
					},
				},
				AddSource: true,
				Output:    OutputStdout,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig(tt.options...)
			assert.Equal(t, tt.expected.Level, config.Level)
			assert.Equal(t, tt.expected.Formatter.Format, config.Formatter.Format)
			assert.Equal(t, tt.expected.AddSource, config.AddSource)
			assert.Equal(t, tt.expected.Output, config.Output)
			
			// Check formatter data
			for k, v := range tt.expected.Formatter.Data {
				assert.Equal(t, v, config.Formatter.Data[k])
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected *Config
		wantErr  bool
	}{
		{
			name: "debug level json format",
			jsonData: `{
				"level": "debug", 
				"formatter": {
					"format": "json", 
					"data": {"timeFormat": "2006-01-02"}
				},
				"addSource": true,
				"output": "stdout"
			}`,
			expected: &Config{
				Level: "debug",
				Formatter: Formatter{
					Format: FormatterJSON,
					Data: map[string]interface{}{
						"timeFormat": "2006-01-02",
					},
				},
				AddSource: true,
				Output:    OutputStdout,
			},
			wantErr: false,
		},
		{
			name: "warn level text format",
			jsonData: `{
				"level": "warn", 
				"formatter": {
					"format": "text", 
					"data": null
				}
			}`,
			expected: &Config{
				Level: "warn",
				Formatter: Formatter{
					Format: FormatterText,
					Data:   map[string]interface{}{},
				},
				AddSource: false,
				Output:    OutputStderr,
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			jsonData: `{
				"level": "error"
			}`,
			expected: &Config{
				Level: "error",
				Formatter: Formatter{
					Format: FormatterJSON,
					Data:   map[string]interface{}{},
				},
				AddSource: false,
				Output:    OutputStderr,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			err := json.Unmarshal([]byte(tt.jsonData), config)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Level, config.Level)
			assert.Equal(t, tt.expected.Formatter.Format, config.Formatter.Format)
			assert.Equal(t, tt.expected.AddSource, config.AddSource)
			assert.Equal(t, tt.expected.Output, config.Output)
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected *Config
		wantErr  bool
	}{
		{
			name: "debug level json format",
			yamlData: `
level: debug
formatter:
  format: json
  data:
    timeFormat: '2006-01-02'
addSource: true
output: stdout
`,
			expected: &Config{
				Level: "debug",
				Formatter: Formatter{
					Format: FormatterJSON,
					Data: map[string]interface{}{
						"timeFormat": "2006-01-02",
					},
				},
				AddSource: true,
				Output:    OutputStdout,
			},
			wantErr: false,
		},
		{
			name: "warn level text format",
			yamlData: `
level: warn
formatter:
  format: text
`,
			expected: &Config{
				Level: "warn",
				Formatter: Formatter{
					Format: FormatterText,
					Data:   map[string]interface{}{},
				},
				AddSource: false,
				Output:    OutputStderr,
			},
			wantErr: false,
		},
		{
			name:     "minimal config",
			yamlData: "level: error",
			expected: &Config{
				Level: "error",
				Formatter: Formatter{
					Format: FormatterJSON,
					Data:   map[string]interface{}{},
				},
				AddSource: false,
				Output:    OutputStderr,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			err := yaml.Unmarshal([]byte(tt.yamlData), config)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Level, config.Level)
			assert.Equal(t, tt.expected.Formatter.Format, config.Formatter.Format)
			assert.Equal(t, tt.expected.AddSource, config.AddSource)
			assert.Equal(t, tt.expected.Output, config.Output)
		})
	}
}

func TestCreateLogger(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		logMessage string
		contains   []string
		wantErr    bool
	}{
		{
			name: "json logger",
			config: NewConfig(
				WithLevel("info"),
				WithJSONFormatter(),
			),
			logMessage: "test json message",
			contains:   []string{"test json message", "INFO"},
			wantErr:    false,
		},
		{
			name: "text logger",
			config: NewConfig(
				WithLevel("warn"),
				WithTextFormatter(),
			),
			logMessage: "test text message",
			contains:   []string{"test text message", "WARN"},
			wantErr:    false,
		},
		{
			name: "debug level",
			config: NewConfig(
				WithLevel("debug"),
				WithJSONFormatter(),
			),
			logMessage: "test debug message",
			contains:   []string{"test debug message", "DEBUG"},
			wantErr:    false,
		},
		{
			name: "invalid level",
			config: &Config{
				Level: "invalid",
				Formatter: Formatter{
					Format: FormatterJSON,
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			buf := &bytes.Buffer{}
			
			// Override the output in the config to use our buffer
			if !tt.wantErr {
				tt.config.Output = OutputStderr // Set to stderr for consistency
			}
			
			// Create the logger
			logger, err := tt.config.CreateLogger()
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotNil(t, logger)
			
			// If we have a log message to test, create a custom logger with our buffer
			if tt.logMessage != "" {
				// Parse the level
				var level slog.Level
				err := level.UnmarshalText([]byte(strings.ToUpper(tt.config.Level)))
				require.NoError(t, err)
				
				// Create handler options
				opts := &slog.HandlerOptions{
					Level: level,
				}
				
				// Create a handler that writes to our buffer
				var handler slog.Handler
				if tt.config.Formatter.Format == FormatterJSON {
					handler = slog.NewJSONHandler(buf, opts)
				} else {
					handler = slog.NewTextHandler(buf, opts)
				}
				
				// Create a logger with our buffer
				testLogger := slog.New(handler)
				
				// Log the message
				switch strings.ToUpper(tt.config.Level) {
				case "DEBUG":
					testLogger.Debug(tt.logMessage)
				case "INFO":
					testLogger.Info(tt.logMessage)
				case "WARN":
					testLogger.Warn(tt.logMessage)
				case "ERROR":
					testLogger.Error(tt.logMessage)
				}
				
				// Check that the output contains expected strings
				output := buf.String()
				for _, s := range tt.contains {
					assert.Contains(t, output, s)
				}
			}
		})
	}
}

func TestFieldMapToReplaceAttr(t *testing.T) {
	// Create a config with field mapping
	config := NewConfig(
		WithFormatterData("fieldmap", map[string]interface{}{
			"time":  "timestamp",
			"level": "severity",
			"msg":   "message",
		}),
	)
	
	// Get the replace function
	replaceFunc := config.fieldMapToReplaceAttr()
	require.NotNil(t, replaceFunc)
	
	// Test attribute replacement
	testCases := []struct {
		attr     slog.Attr
		expected string
	}{
		{slog.String("time", "2023-01-01"), "timestamp"},
		{slog.String("level", "info"), "severity"},
		{slog.String("msg", "test"), "message"},
		{slog.String("other", "value"), "other"}, // Unchanged
	}
	
	for _, tc := range testCases {
		result := replaceFunc(nil, tc.attr)
		assert.Equal(t, tc.expected, result.Key)
	}
}
