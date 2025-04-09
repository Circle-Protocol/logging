package logging

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// Entry represents a log entry with additional functionality
// for conditional logging and field management
type Entry struct {
	logger    *slog.Logger
	isOnError bool
	err       error
	ctx       context.Context
}

var idKey = "logID"

// SetIDKey sets the key used for log entry IDs
func SetIDKey(key string) {
	idKey = key
}

// Log creates a new entry with an ID (maintained for backward compatibility)
func Log(id string) *Entry {
	return New().WithField(idKey, id)
}

// LogWithFields creates a new entry with an ID and the given fields
// (maintained for backward compatibility)
func LogWithFields(id string, fields ...interface{}) *Entry {
	return Log(id).SetFields(fields...)
}

// New instantiates a new entry with the default logger
func New() *Entry {
	return &Entry{
		logger: GetLogger(),
		ctx:    context.Background(),
	}
}

// NewWithContext instantiates a new entry with the default logger and context
func NewWithContext(ctx context.Context) *Entry {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = GetLogger()
	}
	return &Entry{
		logger: logger,
		ctx:    ctx,
	}
}

// OnError sets the error. The log will only be printed if err is not nil
func OnError(err error) *Entry {
	return New().OnError(err)
}

// WithError creates a new entry with the given error
func WithError(err error) *Entry {
	return New().WithError(err)
}

// WithFields creates a new entry with the given fields
func WithFields(fields ...interface{}) *Entry {
	return New().SetFields(fields...)
}

// EntryWithContext creates a new entry with the given context
func EntryWithContext(ctx context.Context) *Entry {
	return NewWithContext(ctx)
}

// OnError sets the error. The log will only be printed if err is not nil
func (e *Entry) OnError(err error) *Entry {
	e.err = err
	e.isOnError = true
	return e
}

// SetFields sets the given fields on the entry
// It panics if the length of fields is odd
func (e *Entry) SetFields(fields ...interface{}) *Entry {
	if len(fields)%2 != 0 {
		panic(fmt.Sprintf("odd number of arguments passed as key-value pairs for fields: %d", len(fields)))
	}

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			panic(fmt.Sprintf("key is not a string: %v", fields[i]))
		}
		e = e.WithField(key, fields[i+1])
	}
	return e
}

// WithField adds a field to the entry
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return &Entry{
		logger:    e.logger.With(key, value),
		isOnError: e.isOnError,
		err:       e.err,
		ctx:       e.ctx,
	}
}

// WithFields adds multiple fields to the entry
func (e *Entry) WithFields(fields map[string]interface{}) *Entry {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return &Entry{
		logger:    e.logger.With(attrs...),
		isOnError: e.isOnError,
		err:       e.err,
		ctx:       e.ctx,
	}
}

// WithError adds an error field to the entry
func (e *Entry) WithError(err error) *Entry {
	if err == nil {
		return e
	}
	return &Entry{
		logger:    e.logger.With("error", err),
		isOnError: e.isOnError,
		err:       err, // Store the error in the Entry
		ctx:       e.ctx,
	}
}

// WithTime adds a time field to the entry
func (e *Entry) WithTime(t time.Time) *Entry {
	return &Entry{
		logger:    e.logger.With("time", t),
		isOnError: e.isOnError,
		err:       e.err,
		ctx:       e.ctx,
	}
}

// WithContext sets the context for the entry
func (e *Entry) WithContext(ctx context.Context) *Entry {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = e.logger
	}
	return &Entry{
		logger:    logger,
		isOnError: e.isOnError,
		err:       e.err,
		ctx:       ctx,
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Debug(fmt.Sprint(args...)) })
}

// Debug logs a debug message
func (e *Entry) Debug(args ...interface{}) {
	e.log(func() { e.logger.Debug(fmt.Sprint(args...)) })
}

// Debugln logs a debug message with a newline
func Debugln(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Debug(fmt.Sprintln(args...)) })
}

// Debugln logs a debug message with a newline
func (e *Entry) Debugln(args ...interface{}) {
	e.log(func() { e.logger.Debug(fmt.Sprintln(args...)) })
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Debug(fmt.Sprintf(format, args...)) })
}

// Debugf logs a formatted debug message
func (e *Entry) Debugf(format string, args ...interface{}) {
	e.log(func() { e.logger.Debug(fmt.Sprintf(format, args...)) })
}

// Info logs an info message
func Info(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Info(fmt.Sprint(args...)) })
}

// Info logs an info message
func (e *Entry) Info(args ...interface{}) {
	e.log(func() { e.logger.Info(fmt.Sprint(args...)) })
}

// Infoln logs an info message with a newline
func Infoln(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Info(fmt.Sprintln(args...)) })
}

// Infoln logs an info message with a newline
func (e *Entry) Infoln(args ...interface{}) {
	e.log(func() { e.logger.Info(fmt.Sprintln(args...)) })
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Info(fmt.Sprintf(format, args...)) })
}

// Infof logs a formatted info message
func (e *Entry) Infof(format string, args ...interface{}) {
	e.log(func() { e.logger.Info(fmt.Sprintf(format, args...)) })
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Warn(fmt.Sprint(args...)) })
}

// Warn logs a warning message
func (e *Entry) Warn(args ...interface{}) {
	e.log(func() { e.logger.Warn(fmt.Sprint(args...)) })
}

// Warnln logs a warning message with a newline
func Warnln(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Warn(fmt.Sprintln(args...)) })
}

// Warnln logs a warning message with a newline
func (e *Entry) Warnln(args ...interface{}) {
	e.log(func() { e.logger.Warn(fmt.Sprintln(args...)) })
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Warn(fmt.Sprintf(format, args...)) })
}

// Warnf logs a formatted warning message
func (e *Entry) Warnf(format string, args ...interface{}) {
	e.log(func() { e.logger.Warn(fmt.Sprintf(format, args...)) })
}

// Warning is an alias for Warn (maintained for backward compatibility)
func Warning(args ...interface{}) {
	Warn(args...)
}

// Warning is an alias for Warn (maintained for backward compatibility)
func (e *Entry) Warning(args ...interface{}) {
	e.Warn(args...)
}

// Warningln is an alias for Warnln (maintained for backward compatibility)
func Warningln(args ...interface{}) {
	Warnln(args...)
}

// Warningln is an alias for Warnln (maintained for backward compatibility)
func (e *Entry) Warningln(args ...interface{}) {
	e.Warnln(args...)
}

// Warningf is an alias for Warnf (maintained for backward compatibility)
func Warningf(format string, args ...interface{}) {
	Warnf(format, args...)
}

// Warningf is an alias for Warnf (maintained for backward compatibility)
func (e *Entry) Warningf(format string, args ...interface{}) {
	e.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Error(fmt.Sprint(args...)) })
}

// Error logs an error message
func (e *Entry) Error(args ...interface{}) {
	e.log(func() { e.logger.Error(fmt.Sprint(args...)) })
}

// Errorln logs an error message with a newline
func Errorln(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Error(fmt.Sprintln(args...)) })
}

// Errorln logs an error message with a newline
func (e *Entry) Errorln(args ...interface{}) {
	e.log(func() { e.logger.Error(fmt.Sprintln(args...)) })
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Error(fmt.Sprintf(format, args...)) })
}

// Errorf logs a formatted error message
func (e *Entry) Errorf(format string, args ...interface{}) {
	e.log(func() { e.logger.Error(fmt.Sprintf(format, args...)) })
}

// Fatal logs a fatal message and exits with status code 1
func Fatal(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Error(fmt.Sprint(args...)) })
	panic("fatal: " + fmt.Sprint(args...))
}

// Fatal logs a fatal message and exits with status code 1
func (e *Entry) Fatal(args ...interface{}) {
	e.log(func() { e.logger.Error(fmt.Sprint(args...)) })
	panic("fatal: " + fmt.Sprint(args...))
}

// Fatalln logs a fatal message with a newline and exits with status code 1
func Fatalln(args ...interface{}) {
	e := New()
	e.log(func() { e.logger.Error(fmt.Sprintln(args...)) })
	panic("fatal: " + fmt.Sprintln(args...))
}

// Fatalln logs a fatal message with a newline and exits with status code 1
func (e *Entry) Fatalln(args ...interface{}) {
	e.log(func() { e.logger.Error(fmt.Sprintln(args...)) })
	panic("fatal: " + fmt.Sprintln(args...))
}

// Fatalf logs a formatted fatal message and exits with status code 1
func Fatalf(format string, args ...interface{}) {
	e := New()
	msg := fmt.Sprintf(format, args...)
	e.log(func() { e.logger.Error(msg) })
	panic("fatal: " + msg)
}

// Fatalf logs a formatted fatal message and exits with status code 1
func (e *Entry) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.log(func() { e.logger.Error(msg) })
	panic("fatal: " + msg)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	e := New()
	msg := fmt.Sprint(args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Panic logs a panic message and panics
func (e *Entry) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Panicln logs a panic message with a newline and panics
func Panicln(args ...interface{}) {
	e := New()
	msg := fmt.Sprintln(args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Panicln logs a panic message with a newline and panics
func (e *Entry) Panicln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	e := New()
	msg := fmt.Sprintf(format, args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Panicf logs a formatted panic message and panics
func (e *Entry) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.log(func() { e.logger.Error(msg) })
	panic(msg)
}

// Log logs a message with the specified level
func (e *Entry) Log(level slog.Level, args ...interface{}) {
	e.log(func() {
		switch level {
		case slog.LevelDebug:
			e.logger.Debug(fmt.Sprint(args...))
		case slog.LevelInfo:
			e.logger.Info(fmt.Sprint(args...))
		case slog.LevelWarn:
			e.logger.Warn(fmt.Sprint(args...))
		case slog.LevelError:
			e.logger.Error(fmt.Sprint(args...))
		default:
			e.logger.Info(fmt.Sprint(args...))
		}
	})
}

// Logf logs a formatted message with the specified level
func Logf(level slog.Level, format string, args ...interface{}) {
	e := New()
	e.Logf(level, format, args...)
}

// Logf logs a formatted message with the specified level
func (e *Entry) Logf(level slog.Level, format string, args ...interface{}) {
	e.log(func() {
		msg := fmt.Sprintf(format, args...)
		switch level {
		case slog.LevelDebug:
			e.logger.Debug(msg)
		case slog.LevelInfo:
			e.logger.Info(msg)
		case slog.LevelWarn:
			e.logger.Warn(msg)
		case slog.LevelError:
			e.logger.Error(msg)
		default:
			e.logger.Info(msg)
		}
	})
}

// Logln logs a message with a newline and the specified level
func Logln(level slog.Level, args ...interface{}) {
	e := New()
	e.Logln(level, args...)
}

// Logln logs a message with a newline and the specified level
func (e *Entry) Logln(level slog.Level, args ...interface{}) {
	e.log(func() {
		msg := fmt.Sprintln(args...)
		switch level {
		case slog.LevelDebug:
			e.logger.Debug(msg)
		case slog.LevelInfo:
			e.logger.Info(msg)
		case slog.LevelWarn:
			e.logger.Warn(msg)
		case slog.LevelError:
			e.logger.Error(msg)
		default:
			e.logger.Info(msg)
		}
	})
}

// log executes the logging function if the entry is not conditional
// or if the condition is met (error is not nil for OnError entries)
func (e *Entry) log(logFunc func()) {
	// Check if this is a conditional log (OnError) and handle accordingly
	if e.isOnError {
		// If no error, don't log anything
		if e.err == nil {
			return
		}
		// If there's an error, add it to the logger
		e.logger = e.logger.With("error", e.err)
	}
	
	// Add caller information
	_, file, line, ok := runtime.Caller(2)
	if ok {
		e.logger = e.logger.With("caller", fmt.Sprintf("%s:%d", file, line))
	}
	
	// Execute the log function
	logFunc()
}

// checkOnError returns nil if the entry is conditional (isOnError is true)
// and the condition is not met (err is nil)
// This method is deprecated and kept for compatibility
// The logic has been moved to the log method
func (e *Entry) checkOnError() *Entry {
	return e
}
