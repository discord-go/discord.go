package http

import (
	"net/http"
)

// LoggingTransport is an http.RoundTripper that allows wrapping another transport
// and could be extended to log requests or track metrics.
type LoggingTransport struct {
	Base http.RoundTripper
}

// RoundTrip executes a single HTTP transaction.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	// Pass-through for now; this can be extended with logging or metrics tracking.
	return base.RoundTrip(req)
}
