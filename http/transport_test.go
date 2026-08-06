package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockTransport struct {
	Called bool
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.Called = true
	return nil, errors.New("mock error")
}

func TestLoggingTransport(t *testing.T) {
	mt := &mockTransport{}
	transport := &LoggingTransport{Base: mt}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	_, err := transport.RoundTrip(req)

	if !mt.Called {
		t.Error("expected base transport to be called")
	}
	if err == nil || err.Error() != "mock error" {
		t.Errorf("expected mock error, got %v", err)
	}
}

func TestLoggingTransport_NilBase(t *testing.T) {
	transport := &LoggingTransport{Base: nil}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
