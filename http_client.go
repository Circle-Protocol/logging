package logging

import (
	"context"
	"net/http"
	"time"

	"log/slog"
)

// ClientLoggerOption is a function that configures a logRoundTripper
type ClientLoggerOption func(*logRoundTripper)

// WithFallbackLogger uses the passed logger if none was
// found in the context.
func WithFallbackLogger(logger *slog.Logger) ClientLoggerOption {
	return func(lrt *logRoundTripper) {
		lrt.fallback = logger
	}
}

// WithClientDurationFunc allows overriding the request duration
// for testing.
func WithClientDurationFunc(df func(time.Time) time.Duration) ClientLoggerOption {
	return func(lrt *logRoundTripper) {
		lrt.duration = df
	}
}

// WithClientGroup groups the log attributes
// produced by the client.
func WithClientGroup(name string) ClientLoggerOption {
	return func(lrt *logRoundTripper) {
		lrt.group = name
	}
}

// WithClientRequestAttr allows customizing the information used
// from a request as request attributes.
func WithClientRequestAttr(requestToAttr func(*http.Request) slog.Attr) ClientLoggerOption {
	return func(lrt *logRoundTripper) {
		lrt.reqToAttr = requestToAttr
	}
}

// WithClientResponseAttr allows customizing the information used
// from a response as response attributes.
func WithClientResponseAttr(responseToAttr func(*http.Response) slog.Attr) ClientLoggerOption {
	return func(lrt *logRoundTripper) {
		lrt.resToAttr = responseToAttr
	}
}

// EnableHTTPClient adds slog functionality to the HTTP client.
// It attempts to obtain a logger with WithContext.
// If no logger is in the context, it tries to use a fallback logger,
// which might be set by WithFallbackLogger.
// If no logger was found finally, the Transport is
// executed without logging.
func EnableHTTPClient(c *http.Client, opts ...ClientLoggerOption) error {
	if c == nil {
		return ErrNilClient
	}

	lrt := &logRoundTripper{
		next:      c.Transport,
		duration:  time.Since,
		reqToAttr: RequestToAttr,
		resToAttr: ResponseToAttr,
	}

	if lrt.next == nil {
		lrt.next = http.DefaultTransport
	}

	for _, opt := range opts {
		opt(lrt)
	}

	c.Transport = lrt
	return nil
}

// ErrNilClient is returned when a nil client is provided to EnableHTTPClient
var ErrNilClient = &clientError{"nil client provided"}

type clientError struct {
	msg string
}

func (e *clientError) Error() string {
	return e.msg
}

// logRoundTripper is an http.RoundTripper that logs requests and responses
type logRoundTripper struct {
	next      http.RoundTripper
	duration  func(time.Time) time.Duration
	fallback  *slog.Logger
	group     string
	reqToAttr func(*http.Request) slog.Attr
	resToAttr func(*http.Response) slog.Attr
}

// RoundTrip implements http.RoundTripper
func (l *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get logger from context or fallback
	logger, ok := l.fromContextOrFallback(req.Context())
	if !ok {
		// No logger available, just execute the transport
		return l.next.RoundTrip(req)
	}

	// Record start time
	start := time.Now()

	// Execute the request
	resp, err := l.next.RoundTrip(req)

	// Add request info to logger
	if l.group != "" {
		logger = logger.WithGroup(l.group)
	}

	// Create attributes for the log entry
	attrs := []any{
		l.reqToAttr(req),
		slog.Duration("duration", l.duration(start)),
	}

	// Log the result
	if err != nil {
		logger.With(attrs...).Error("request roundtrip", "error", err)
		return resp, err
	}

	// Include response attributes if we have a response
	if resp != nil {
		attrs = append(attrs, l.resToAttr(resp))
	}

	logger.With(attrs...).Info("request roundtrip")
	return resp, nil
}

// fromContextOrFallback gets a logger from context or uses the fallback
func (l *logRoundTripper) fromContextOrFallback(ctx context.Context) (*slog.Logger, bool) {
	logger := WithContext(ctx)
	if logger != defaultLogger {
		return logger, true
	}
	return l.fallback, l.fallback != nil
}
