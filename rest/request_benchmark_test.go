package rest

import (
	"context"
	"net/http"
	"testing"
)

// BenchmarkRequestNoContent measures local request overhead for an HTTP 204
// with http.NoBody. The mock limiter does not enforce Discord rate limits and
// the mock HTTP client performs no network I/O.
func BenchmarkRequestNoContent(b *testing.B) {
	limiter := &mockLimiter{WaitFunc: func(context.Context, string) error { return nil }}
	response := &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: http.NoBody}
	client := New("benchmark", limiter, &mockHTTPClient{DoFunc: func(*http.Request) (*http.Response, error) {
		return response, nil
	}})
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := client.Request(ctx, http.MethodDelete, "/channels/1/messages/2", nil, nil); err != nil {
			b.Fatal(err)
		}
	}
}
