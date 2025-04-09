package logging

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMiddlewareTestLogger() (out *strings.Builder, logger *slog.Logger) {
	out = new(strings.Builder)
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	// Add a fixed time attribute to make testing easier
	return out, slog.New(handler.WithAttrs([]slog.Attr{slog.String("time", "2025-04-09T12:00:00Z")}))
}

// testWriter is a custom ResponseWriter for testing error scenarios
type testWriter struct {
	*httptest.ResponseRecorder
	err error
}

func (w *testWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.ResponseRecorder.Write(b)
}

// testLoggedWriter wraps a testWriter to implement LoggedWriter
type testLoggedWriter struct {
	*testWriter
}

func (w *testLoggedWriter) Attr() slog.Attr {
	return slog.Group("response",
		slog.Int("status", w.Code),
		slog.Int("written", w.Body.Len()),
	)
}

func (w *testLoggedWriter) Err() error {
	return w.err
}

// newTestLoggedWriter creates a LoggedWriter for testing
func newTestLoggedWriter(w http.ResponseWriter) LoggedWriter {
	tw, ok := w.(*testWriter)
	if !ok {
		panic("expected *testWriter")
	}
	return &testLoggedWriter{tw}
}

// newTestWriter creates a testWriter with the given error
func newTestWriter(err error) *testWriter {
	return &testWriter{
		ResponseRecorder: httptest.NewRecorder(),
		err:              err,
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		groupName  string
		want       string
		statusCode int
	}{
		{
			name: "successful request",
			want: `{
				"time": "2025-04-09T12:00:00Z",
				"level": "INFO",
				"msg": "request served",
				"id": "id1",
				"duration": 1000000000,
				"request": {
					"method": "GET",
					"url": "https://example.com/path/"
				},
				"response": {
					"status": 200,
					"written": 13
				}
			}`,
			statusCode: http.StatusOK,
		},
		{
			name:      "request with group",
			groupName: "http",
			want: `{
				"time": "2025-04-09T12:00:00Z",
				"level": "INFO",
				"msg": "request served",
				"http": {
					"id": "id1",
					"duration": 1000000000,
					"request": {
						"method": "GET",
						"url": "https://example.com/path/"
					},
					"response": {
						"status": 200,
						"written": 13
					}
				}
			}`,
			statusCode: http.StatusOK,
		},
		{
			name: "request with error",
			err:  io.ErrClosedPipe,
			want: `{
				"time": "2025-04-09T12:00:00Z",
				"level": "WARN",
				"msg": "write response",
				"error": "io: read/write on closed pipe",
				"id": "id1",
				"duration": 1000000000,
				"request": {
					"method": "GET",
					"url": "https://example.com/path/"
				},
				"response": {
					"status": 200,
					"written": 0
				}
			}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "request with custom status code",
			statusCode: http.StatusNotFound,
			want: `{
				"time": "2025-04-09T12:00:00Z",
				"level": "INFO",
				"msg": "request served",
				"id": "id1",
				"duration": 1000000000,
				"request": {
					"method": "GET",
					"url": "https://example.com/path/"
				},
				"response": {
					"status": 404,
					"written": 13
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger that writes to a string builder
			logOut, logger := newMiddlewareTestLogger()

			// Create middleware options
			options := []MiddlewareOption{
				WithMiddlewareLogger(logger),
				WithIDFunc(func() slog.Attr {
					return slog.String("id", "id1")
				}),
				WithDurationFunc(func(time.Time) time.Duration {
					return time.Second
				}),
			}

			// Add group option if specified
			if tt.groupName != "" {
				options = append(options, WithMiddlewareGroup(tt.groupName))
			}

			// Create the middleware
			// We'll use a custom wrapped middleware below, so this variable isn't used directly

			// Create a handler that writes a response and sets status code if specified
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.statusCode != 0 && tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
				}
				fmt.Fprint(w, "Hello, World!")
			})

			// Create a test writer with the given error
			w := newTestWriter(tt.err)

			// Create a test request
			r := httptest.NewRequest("GET", "https://example.com/path/", nil)

			// Use a custom LoggedWriter for testing
			wrappedMw := Middleware(append(options, WithLoggedWriter(newTestLoggedWriter))...)

			// Execute the middleware
			wrappedMw(next).ServeHTTP(w, r)

			// Verify the log output
			got := logOut.String()
			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestLoggedWriter(t *testing.T) {
	// Test the default LoggedWriter implementation
	rec := httptest.NewRecorder()
	lw := newLoggedWriter(rec)

	// Test WriteHeader
	lw.WriteHeader(http.StatusCreated)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Test Write
	n, err := lw.Write([]byte("test"))
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, "test", rec.Body.String())

	// Test Attr
	attr := lw.Attr()
	require.Equal(t, "response", attr.Key)
	require.Equal(t, slog.KindGroup, attr.Value.Kind())

	// Test default status code
	rec2 := httptest.NewRecorder()
	lw2 := newLoggedWriter(rec2)
	n, err = lw2.Write([]byte("test"))
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, http.StatusOK, rec2.Code)

	// Test Err
	require.NoError(t, lw.Err())
}

func TestMiddlewareWithRequestContext(t *testing.T) {
	// Create a logger that writes to a string builder
	logOut, logger := newMiddlewareTestLogger()

	// Create middleware
	mw := Middleware(
		WithMiddlewareLogger(logger),
	)

	// Create a handler that gets the logger from context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger from context
		ctxLogger, ok := FromContext(r.Context())
		require.True(t, ok, "Logger should be in context")

		// Log something with the context logger
		ctxLogger.Info("handler log")

		fmt.Fprint(w, "Hello, World!")
	})

	// Create a test request and response
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://example.com/path/", nil)

	// Execute the middleware
	mw(next).ServeHTTP(w, r)

	// Verify the log contains both the handler log and the request served log
	got := logOut.String()
	assert.Contains(t, got, "handler log")
	assert.Contains(t, got, "request served")
}
