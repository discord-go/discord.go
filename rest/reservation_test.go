package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/discord-go/discord.go/ratelimit"
)

func TestUnknownRouteAllowsOneDiscoveryRequest(t *testing.T) {
	var calls atomic.Int32
	started := make(chan struct{})
	finish := make(chan struct{})
	c := New("token", nil, &mockHTTPClient{DoFunc: func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-finish
		}
		return &http.Response{StatusCode: 204, Header: http.Header{
			"X-Ratelimit-Bucket": {"bucket"}, "X-Ratelimit-Remaining": {"0"},
			"X-Ratelimit-Reset-After": {"1"},
		}, Body: io.NopCloser(strings.NewReader(""))}, nil
	}})
	first := make(chan error, 1)
	go func() { first <- c.Request(context.Background(), "GET", "/channels/1/messages", nil, nil) }()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := c.Request(ctx, "GET", "/channels/1/messages", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting request: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("sent %d requests before discovery completed", got)
	}
	close(finish)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
}

func TestProbeReleasedOnTransportFailure(t *testing.T) {
	var calls atomic.Int32
	c := New("token", nil, &mockHTTPClient{DoFunc: func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("connection failed")
		}
		return &http.Response{StatusCode: 204, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	}})
	if err := c.Request(context.Background(), "GET", "/test", nil, nil); err == nil {
		t.Fatal("expected transport error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Request(ctx, "GET", "/test", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredBucketAllowsOneProbe(t *testing.T) {
	l := ratelimit.NewLimiter(nil)
	route := l.(ratelimit.RouteLimiter)
	reserver := l.(ratelimit.RouteReserver)
	route.UpdateRoute("GET", "/test", ratelimit.Info{Bucket: "b", Reset: time.Now().Add(-time.Second)})
	release, err := reserver.AcquireRoute(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			if done, err := reserver.AcquireRoute(ctx, "GET", "/test"); !errors.Is(err, context.DeadlineExceeded) {
				if done != nil {
					done()
				}
				t.Errorf("waiting for probe: %v", err)
			}
		}()
	}
	wg.Wait()
}

type failingUpdateLimiter struct{ err error }

func (l *failingUpdateLimiter) Wait(context.Context, string) error { return nil }
func (l *failingUpdateLimiter) Update(string, ratelimit.Info)      {}
func (l *failingUpdateLimiter) UpdateRouteContext(context.Context, string, string, ratelimit.Info) error {
	return l.err
}

func TestResponseUpdateErrorIsReturned(t *testing.T) {
	backendErr := errors.New("coordination unavailable")
	client := New("token", &failingUpdateLimiter{err: backendErr}, &mockHTTPClient{DoFunc: func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 204, Header: http.Header{"X-Ratelimit-Remaining": {"2"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	}})
	if err := client.Request(context.Background(), "POST", "/test", nil, nil); !errors.Is(err, backendErr) {
		t.Fatalf("got %v, want backend error", err)
	}
}
