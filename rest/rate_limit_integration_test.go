package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitResponseBlocksOtherRequests(t *testing.T) {
	for _, global := range []bool{false, true} {
		t.Run(map[bool]string{false: "route", true: "global"}[global], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Bucket", "abc")
				w.Header().Set("X-RateLimit-Remaining", "5")
				w.WriteHeader(http.StatusTooManyRequests)
				if global {
					io.WriteString(w, `{"retry_after":1,"global":true}`)
				} else {
					io.WriteString(w, `{"retry_after":1}`)
				}
			}))
			defer server.Close()
			c := New("token", nil, &testHTTPClient{})
			c.BaseURL = server.URL
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
			defer cancel()
			if err := c.Request(ctx, "GET", "/channels/1/messages", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("first request: %v", err)
			}
			path := "/channels/1/messages"
			if global {
				path = "/guilds/2"
			}
			ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel2()
			if err := c.waitRoute(ctx2, "GET", path); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("other request did not observe cooldown: %v", err)
			}
		})
	}
}

func TestMultipartRetriesPreserveBody(t *testing.T) {
	for _, form := range []bool{false, true} {
		t.Run(map[bool]string{false: "payload", true: "form"}[form], func(t *testing.T) {
			attempts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if err := r.ParseMultipartForm(1 << 20); err != nil {
					t.Error(err)
					return
				}
				file, _, err := r.FormFile("files[0]")
				if err != nil {
					t.Error(err)
					return
				}
				content, _ := io.ReadAll(file)
				file.Close()
				if string(content) != "test" {
					t.Errorf("file content %q", content)
				}
				if attempts == 1 {
					w.WriteHeader(http.StatusTooManyRequests)
					io.WriteString(w, `{"retry_after":0.001}`)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			c := New("token", nil, &testHTTPClient{})
			c.BaseURL = server.URL
			c.MaxRetries = 1
			files := []File{FileFromBytes("a.txt", []byte("test"))}
			var err error
			if form {
				err = c.RequestMultipartForm(context.Background(), "POST", "/test", nil, files, nil)
			} else {
				err = c.RequestMultipart(context.Background(), "POST", "/test", nil, files, nil)
			}
			if err != nil || attempts != 2 {
				t.Fatalf("attempts=%d err=%v", attempts, err)
			}
		})
	}
}
