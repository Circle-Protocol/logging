package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config defines the configuration for the logging system
type Config struct {
	// Level sets the minimum log level (debug, info, warn, error)
	Level string `json:"level" yaml:"level"`

	// Formatter configures the log output format and options
	Formatter Formatter `json:"formatter" yaml:"formatter"`

	// AddSource adds source code information to log entries
	AddSource bool `json:"addSource" yaml:"addSource"`

	// Output specifies where logs should be written (default: stderr)
	Output string `json:"output" yaml:"output"`
}

// Formatter defines the log format configuration
type Formatter struct {
	// Format specifies the output format (json or text)
	Format string `json:"format" yaml:"format"`

	// Data contains additional formatter-specific options
	Data map[string]interface{} `json:"data" yaml:"data"`
}

// Format constants
const (
	FormatterText = "text"
	FormatterJSON = "json"
)

// Default output destinations
const (
	OutputStdout = "stdout"
	OutputStderr = "stderr"
	OutputFile   = "file"
)

// ConfigOption is a function that modifies a Config
type ConfigOption func(*Config)

// WithLevel sets the log level
func WithLevel(level string) ConfigOption {
	return func(c *Config) {
		c.Level = level
	}
}

// WithJSONFormatter sets the formatter to JSON
func WithJSONFormatter() ConfigOption {
	return func(c *Config) {
		c.Formatter.Format = FormatterJSON
	}
}

// WithTextFormatter sets the formatter to text
func WithTextFormatter() ConfigOption {
	return func(c *Config) {
		c.Formatter.Format = FormatterText
	}
}

// WithAddSource enables source code information in logs
func WithAddSource(enable bool) ConfigOption {
	return func(c *Config) {
		c.AddSource = enable
	}
}

// WithOutput sets the output destination
func WithOutput(output string) ConfigOption {
	return func(c *Config) {
		c.Output = output
	}
}

// WithFormatterData adds custom data to the formatter
func WithFormatterData(key string, value interface{}) ConfigOption {
	return func(c *Config) {
		if c.Formatter.Data == nil {
			c.Formatter.Data = make(map[string]interface{})
		}
		c.Formatter.Data[key] = value
	}
}

// NewConfig creates a new Config with default values
func NewConfig(opts ...ConfigOption) *Config {
	// Default configuration
	config := &Config{
		Level: "info",
		Formatter: Formatter{
			Format: FormatterJSON,
			Data:   make(map[string]interface{}),
		},
		AddSource: false,
		Output:    OutputStderr,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	return config
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type configAlias Config
	alias := configAlias(*NewConfig())

	if err := unmarshal(&alias); err != nil {
		return err
	}

	*c = Config(alias)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Config) UnmarshalJSON(data []byte) error {
	type configAlias Config
	alias := configAlias(*NewConfig())

	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	*c = Config(alias)
	return nil
}

// Apply applies the configuration to create a new logger
func (c *Config) Apply() error {
	logger, err := c.CreateLogger()
	if err != nil {
		return err
	}

	// Update the global logger
	slog.SetDefault(logger)
	return nil
}

// CreateLogger creates a new logger based on the configuration
func (c *Config) CreateLogger() (*slog.Logger, error) {
	// Parse log level
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(c.Level))); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", c.Level, err)
	}

	// Configure handler options
	opts := &slog.HandlerOptions{
		AddSource: c.AddSource,
		Level:     level,
	}

	// Add field mapping if specified
	if fieldMap := c.fieldMapToReplaceAttr(); fieldMap != nil {
		opts.ReplaceAttr = fieldMap
	}

	// Get output writer
	writer, err := c.getOutputWriter()
	if err != nil {
		return nil, err
	}

	// Create handler based on format
	var handler slog.Handler
	switch c.Formatter.Format {
	case FormatterJSON:
		handler = slog.NewJSONHandler(writer, opts)
	case FormatterText, "":
		handler = slog.NewTextHandler(writer, opts)
	default:
		return nil, fmt.Errorf("unsupported formatter format: %s", c.Formatter.Format)
	}

	// Create and return logger
	return slog.New(handler), nil
}

// getOutputWriter returns an io.Writer based on the Output configuration
func (c *Config) getOutputWriter() (io.Writer, error) {
	switch c.Output {
	case OutputStdout, "":
		return os.Stdout, nil
	case OutputStderr:
		return os.Stderr, nil
	case OutputFile:
		// If file path is specified in formatter data
		if filePath, ok := c.Formatter.Data["file"].(string); ok && filePath != "" {
			// Validate file path to prevent path traversal
			cleanPath := filepath.Clean(filePath)
			if !filepath.IsAbs(cleanPath) {
				// Convert relative paths to absolute using current directory
				cwd, err := os.Getwd()
				if err != nil {
					return nil, fmt.Errorf("failed to get current directory: %w", err)
				}
				cleanPath = filepath.Join(cwd, cleanPath)
			}

			// Use more restrictive file permissions (0600 instead of 0666)
			file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file %q: %w", cleanPath, err)
			}
			return file, nil
		}
		return nil, fmt.Errorf("file output specified but no file path provided in formatter data")
	default:
		// Assume it's a file path
		// Validate file path to prevent path traversal
		cleanPath := filepath.Clean(c.Output)
		if !filepath.IsAbs(cleanPath) {
			// Convert relative paths to absolute using current directory
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			cleanPath = filepath.Join(cwd, cleanPath)
		}

		// Use more restrictive file permissions (0600 instead of 0666)
		file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", cleanPath, err)
		}
		return file, nil
	}
}

// fieldMapToReplaceAttr creates a ReplaceAttr function for field mapping
func (c *Config) fieldMapToReplaceAttr() func(groups []string, a slog.Attr) slog.Attr {
	fieldMap, ok := c.Formatter.Data["fieldmap"].(map[string]interface{})
	if !ok {
		return nil
	}

	return func(groups []string, a slog.Attr) slog.Attr {
		// If we have a mapping for this key, replace it
		if newKey, ok := fieldMap[a.Key]; ok {
			if strKey, ok := newKey.(string); ok {
				a.Key = strKey
			}
		}
		return a
	}
}
