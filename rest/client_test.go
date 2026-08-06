package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/ratelimit"
)

type mockLimiter struct {
	WaitFunc func(ctx context.Context, bucket string) error
}

func (m *mockLimiter) Wait(ctx context.Context, bucket string) error {
	return m.WaitFunc(ctx, bucket)
}

func (m *mockLimiter) Update(bucket string, info ratelimit.Info) {}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
func (m *mockHTTPClient) Get(url string) (*http.Response, error) { return nil, nil }
func (m *mockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	c := New("token", nil, nil)
	if c.Token != "token" {
		t.Errorf("Expected token to be 'token', got %s", c.Token)
	}
	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
	if c.Limiter == nil {
		t.Error("Expected Limiter to be initialized")
	}
	if c.BaseURL != "https://discord.com/api/v10" {
		t.Errorf("Expected BaseURL, got %s", c.BaseURL)
	}
}

func TestRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot token" {
			t.Errorf("Expected auth header, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("X-RateLimit-Bucket", "test-bucket")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	var v struct {
		ID string `json:"id"`
	}
	err := c.Request(context.Background(), "GET", "/test", nil, &v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if v.ID != "123" {
		t.Errorf("Expected id 123, got %s", v.ID)
	}
}

func TestRequest_BodyError(t *testing.T) {
	c := New("token", nil, nil)
	err := c.Request(context.Background(), "POST", "/test", make(chan int), nil)
	if err == nil {
		t.Error("Expected marshal error")
	}
}

func TestRequest_NewRequestError(t *testing.T) {
	c := New("token", nil, nil)
	err := c.Request(context.Background(), "INVALID\x00METHOD", "/test", nil, nil)
	if err == nil {
		t.Error("Expected NewRequest error")
	}
}

func TestRequest_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": 10001, "message": "Unknown account"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	err := c.Request(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Code != 10001 {
		t.Errorf("Expected code 10001, got %d", apiErr.Code)
	}
	if apiErr.HTTPStatus != 400 {
		t.Errorf("Expected status 400, got %d", apiErr.HTTPStatus)
	}
	expectedMsg := "discord api error: 10001 (http 400): Unknown account"
	if apiErr.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, apiErr.Error())
	}
}

func TestRequest_HTTPError_NotJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	err := c.Request(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Expected plain text message, got %s", apiErr.Message)
	}
}

func TestRequest_DoError(t *testing.T) {
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	})

	err := c.Request(context.Background(), "GET", "/test", nil, nil)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestRequest_ContextDone(t *testing.T) {
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.Canceled
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Request(ctx, "GET", "/test", nil, nil)
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestRequest_QueryBucket(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL
	err := c.Request(context.Background(), "GET", "/test?query=1", nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequest_LimiterError(t *testing.T) {
	c := New("token", &mockLimiter{
		WaitFunc: func(ctx context.Context, bucket string) error {
			return io.EOF
		},
	}, nil)
	err := c.Request(context.Background(), "GET", "/test", nil, nil)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
func (e errorReader) Close() error { return nil }

func TestRequest_ReadAllError(t *testing.T) {
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       errorReader{},
				Header:     make(http.Header),
			}, nil
		},
	})
	err := c.Request(context.Background(), "GET", "/test", nil, nil)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected io.ErrUnexpectedEOF, got %v", err)
	}
}
