package logging

import (
	"log/slog"
	"net/http"
	"time"
)

// MiddlewareOption is a function that configures a middleware instance
type MiddlewareOption func(*middleware)

// WithLogger sets the passed logger with request attributes
// into the Request's context.
func WithMiddlewareLogger(logger *slog.Logger) MiddlewareOption {
	return func(m *middleware) {
		m.logger = logger
	}
}

// WithMiddlewareGroup groups the log attributes
// produced by the middleware.
func WithMiddlewareGroup(name string) MiddlewareOption {
	return func(m *middleware) {
		m.group = name
	}
}

// WithIDFunc enables the creation of request IDs
// in the middleware, which are then attached to
// the logger.
func WithIDFunc(nextID func() slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.nextID = nextID
	}
}

// WithDurationFunc allows overriding the request duration for testing.
func WithDurationFunc(df func(time.Time) time.Duration) MiddlewareOption {
	return func(m *middleware) {
		m.duration = df
	}
}

// WithRequestAttr allows customizing the information used
// from a request as request attributes.
func WithRequestAttr(requestToAttr func(*http.Request) slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.reqAttr = requestToAttr
	}
}

// WithLoggedWriter allows customizing the writer from
// which post-request attributes are taken.
func WithLoggedWriter(wrap func(w http.ResponseWriter) LoggedWriter) MiddlewareOption {
	return func(m *middleware) {
		m.wrapWriter = wrap
	}
}

// Middleware enables request logging and sets a logger
// to the request context.
// Use FromContext to obtain the logger anywhere in the request lifetime.
//
// The default logger is slog.Default(), with the request's URL and Method
// as preset attributes.
// When the request terminates, an INFO line with the Status Code and
// amount written to the client is printed.
// This behavior can be modified with options.
func Middleware(options ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := &middleware{
			logger:     slog.Default(),
			duration:   time.Since,
			next:       next,
			reqAttr:    RequestToAttr,
			wrapWriter: newLoggedWriter,
		}
		for _, opt := range options {
			opt(mw)
		}
		return mw
	}
}

// middleware is the internal implementation of the HTTP middleware
type middleware struct {
	logger     *slog.Logger
	group      string
	nextID     func() slog.Attr
	next       http.Handler
	duration   func(time.Time) time.Duration
	reqAttr    func(*http.Request) slog.Attr
	wrapWriter func(http.ResponseWriter) LoggedWriter
}

// ServeHTTP implements the http.Handler interface
func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Create a base logger
	logger := m.logger

	// Get request attributes
	requestAttr := m.reqAttr(r)

	// Create a logger with attributes
	if m.group != "" {
		// For grouped logging, we'll collect all attributes at the end
		// Just store the base logger for now
	} else {
		// For non-grouped logging, add request attributes directly
		logger = logger.With(requestAttr)

		// Add ID if configured
		if m.nextID != nil {
			logger = logger.With(m.nextID())
		}
	}

	// Add logger to request context
	r = r.WithContext(ToContext(r.Context(), logger))

	// Wrap the response writer to capture status code and bytes written
	lw := m.wrapWriter(w)

	// Call the next handler
	m.next.ServeHTTP(lw, r)

	// Prepare response attributes
	durationAttr := slog.Duration("duration", m.duration(start))
	responseAttr := lw.Attr()

	// Add response details to logger
	if m.group != "" {
		// For grouped logging, add all attributes under the group name
		groupAttrs := []any{}

		// Add ID if configured
		if m.nextID != nil {
			groupAttrs = append(groupAttrs, m.nextID())
		}

		// Add request attributes
		groupAttrs = append(groupAttrs, requestAttr)

		// Add duration and response
		groupAttrs = append(groupAttrs, durationAttr, responseAttr)

		// Create a new logger with the group
		logger = logger.With(slog.Group(m.group, groupAttrs...))
	} else {
		// For non-grouped logging, just add the response attributes
		logger = logger.With(durationAttr, responseAttr)
	}

	// Check for errors
	if err := lw.Err(); err != nil {
		logger.WarnContext(r.Context(), "write response", "error", err)
		return
	}

	// Log successful request
	logger.InfoContext(r.Context(), "request served")
}

// loggedWriter is a wrapper around http.ResponseWriter that captures
// status code and bytes written
type loggedWriter struct {
	http.ResponseWriter

	statusCode int
	written    int
	err        error
}

// newLoggedWriter creates a new LoggedWriter
func newLoggedWriter(w http.ResponseWriter) LoggedWriter {
	return &loggedWriter{
		ResponseWriter: w,
	}
}

// WriteHeader captures the status code and calls the underlying WriteHeader
func (w *loggedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written and any error
func (w *loggedWriter) Write(b []byte) (int, error) {
	// If WriteHeader was not called, set default status code to 200 OK
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(b)
	w.written += n
	w.err = err
	return n, err
}

// Attr returns a slog.Attr with response information
func (lw *loggedWriter) Attr() slog.Attr {
	return slog.Group("response",
		slog.Int("status", lw.statusCode),
		slog.Int("written", lw.written),
	)
}

// Err returns any error that occurred during writing
func (lw *loggedWriter) Err() error {
	return lw.err
}
