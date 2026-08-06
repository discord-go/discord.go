package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	// Mock sleep to make tests execute instantly
	sleep = func(d time.Duration) {}
}

func TestDefaultClient_Do_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("expected User-Agent 'test-agent', got '%s'", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDefaultClient_Do_ExistingUserAgent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "custom-agent" {
			t.Errorf("expected User-Agent 'custom-agent', got '%s'", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("User-Agent", "custom-agent")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
}

func TestDefaultClient_Do_Retry_502(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 3 {
			w.WriteHeader(http.StatusBadGateway) // 502
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 4 {
		t.Errorf("expected 4 attempts, got %d", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDefaultClient_Do_NoRetry_404(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound) // 404
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestDefaultClient_Do_NoRetrier(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway) // 502
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	client.Retrier = nil // Remove retrier completely

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode)
	}
}

func TestDefaultClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestDefaultClient_Get_InvalidURL(t *testing.T) {
	client := NewClient("test-agent")
	_, err := client.Get(":\x00invalid")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDefaultClient_Post(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test" {
			t.Errorf("expected body 'test', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient("test-agent")
	resp, err := client.Post(ts.URL, "application/json", strings.NewReader("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestDefaultClient_Post_InvalidURL(t *testing.T) {
	client := NewClient("test-agent")
	_, err := client.Post(":\x00invalid", "application/json", nil)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDefaultClient_Do_NoBodyRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway) // 502
	}))
	defer ts.Close()

	client := NewClient("test-agent")

	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader("test"))
	req.GetBody = nil // Ensure GetBody is nil to simulate non-reusable body

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
func (errReader) Close() error { return nil }

func TestDefaultClient_Do_GetBodyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	client := NewClient("test-agent")

	req, _ := http.NewRequest(http.MethodPost, ts.URL, errReader{})
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, errors.New("get body error")
	}

	_, err := client.Do(req)
	if err == nil || err.Error() != "get body error" {
		t.Errorf("expected 'get body error', got: %v", err)
	}
}

func TestDefaultClient_Do_NetworkError(t *testing.T) {
	client := NewClient("test-agent")
	// Port 0 should immediately fail on most platforms
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:0", nil)

	_, err := client.Do(req)
	if err == nil {
		t.Error("expected network error")
	}
}
