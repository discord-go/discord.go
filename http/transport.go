package http

import "net/http"

// Transport is an http.RoundTripper that delegates to a base transport.
// It exists as an extension point for applications that want to wrap
// the default transport with logging, metrics, or tracing.
type Transport struct {
	Base http.RoundTripper
}

// RoundTrip executes a single HTTP transaction by delegating to the Base
// transport. If Base is nil, http.DefaultTransport is used.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// LoggingTransport is a deprecated alias for Transport. It does not perform
// any logging; use Transport instead.
//
// Deprecated: Use Transport.
type LoggingTransport = Transport
