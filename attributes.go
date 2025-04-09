package logging

import (
	"fmt"
	"log/slog"
	"net/http"
)

// StringerValuer returns a Valuer that forces the logger to use the type's String
// method, even in json output mode. By wrapping the type we defer String
// being called to the point we actually log.
func StringerValuer(s fmt.Stringer) slog.LogValuer {
	if s == nil {
		return nil
	}
	return stringerValuer{s}
}

type stringerValuer struct {
	fmt.Stringer
}

func (v stringerValuer) LogValue() slog.Value {
	return slog.StringValue(v.String())
}

// RequestToAttr converts an HTTP request to a structured slog attribute
// containing method and URL information
func RequestToAttr(req *http.Request) slog.Attr {
	if req == nil {
		return slog.Group("request")
	}

	attrs := []any{
		slog.String("method", req.Method),
	}

	if req.URL != nil {
		attrs = append(attrs, slog.Any("url", StringerValuer(req.URL)))
	}

	return slog.Group("request", attrs...)
}

// ResponseToAttr converts an HTTP response to a structured slog attribute
// containing status and content length information
func ResponseToAttr(resp *http.Response) slog.Attr {
	if resp == nil {
		return slog.Group("response")
	}

	return slog.Group("response",
		slog.String("status", resp.Status),
		slog.Int64("content_length", resp.ContentLength),
	)
}

// LoggedWriter stores information regarding the HTTP response.
// This might be status code, amount of data written or headers.
type LoggedWriter interface {
	http.ResponseWriter

	// Attr is called after the next handler in the Middleware returns and
	// the complete response should have been written.
	//
	// The returned Attribute should be a [slog.Group]
	// containing response Attributes.
	Attr() slog.Attr

	// Err is called by the middleware to check if the underlying writer
	// returned an error. If so, the middleware will print an ERROR line.
	Err() error
}
