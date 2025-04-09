package logging

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errRoundTripper is a RoundTripper that always returns an error
type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, io.ErrClosedPipe
}

func TestEnableHTTPClient(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		fromCtx   bool
		wantErr   error
		wantLog   string
	}{
		{
			name:      "nil transport / default",
			transport: nil,
			wantLog: `{
				"time":"not",
				"level":"INFO",
				"msg":"request roundtrip",
				"request":{"method":"GET","url":"%s"},
				"duration":1000000000,
				"response":{
					"status":"200 OK",
					"content_length":14
				}
			}`,
		},
		{
			name:      "transport set",
			transport: http.DefaultTransport,
			wantLog: `{
				"time":"not",
				"level":"INFO",
				"msg":"request roundtrip",
				"request":{"method":"GET","url":"%s"},
				"duration":1000000000,
				"response":{
					"status":"200 OK",
					"content_length":14
				}
			}`,
		},
		{
			name:      "roundtrip error",
			transport: errRoundTripper{},
			wantErr:   io.ErrClosedPipe,
			wantLog: `{
				"time":"not",
				"level":"ERROR",
				"msg":"request roundtrip",
				"request":{"method":"GET","url":"%s"},
				"duration":1000000000,
				"error":"io: read/write on closed pipe"
			}`,
		},
		{
			name:      "logger from ctx",
			transport: http.DefaultTransport,
			fromCtx:   true,
			wantLog: `{
				"time":"not",
				"level":"INFO",
				"msg":"request roundtrip",
				"ctx":{
					"request":{"method":"GET","url":"%s"},
					"duration":1000000000,
					"response":{
						"status":"200 OK",
						"content_length":14
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test logger that writes to a string builder
			out, logger := newTestLogger()

			// Create a client with the specified transport
			c := &http.Client{
				Transport: tt.transport,
			}

			// Enable logging for the client
			err := EnableHTTPClient(c,
				WithFallbackLogger(logger),
				WithClientDurationFunc(func(time.Time) time.Duration {
					return time.Second
				}),
				WithClientRequestAttr(RequestToAttr),
				WithClientResponseAttr(ResponseToAttr),
			)
			require.NoError(t, err)

			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
			}))
			defer ts.Close()

			// Create a request
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)

			// Add logger to context if needed
			if tt.fromCtx {
				req = req.WithContext(ToContext(req.Context(), logger.WithGroup("ctx")))
			}

			// Execute the request
			_, err = c.Do(req)

			// Check for expected error
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			// Format the expected log with the actual server URL
			wantLog := fmt.Sprintf(tt.wantLog, ts.URL)

			// Compare the actual log with the expected log
			assert.JSONEq(t, wantLog, out.String())
		})
	}
}

func TestEnableHTTPClientNilClient(t *testing.T) {
	// Test that EnableHTTPClient returns an error when given a nil client
	err := EnableHTTPClient(nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilClient.Error(), err.Error())
}
