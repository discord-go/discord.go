package ratelimit

import (
	"net/http"
	"testing"
	"time"
)

func TestParseHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Bucket", "my-bucket")
	headers.Set("X-RateLimit-Remaining", "5")
	headers.Set("X-RateLimit-Reset-After", "1.5")
	headers.Set("X-RateLimit-Reset", "1620000000.5")
	headers.Set("X-RateLimit-Global", "true")

	info := ParseHeaders(headers)

	if info.Bucket != "my-bucket" {
		t.Errorf("expected bucket 'my-bucket', got '%s'", info.Bucket)
	}
	if info.Remaining != 5 {
		t.Errorf("expected remaining 5, got %d", info.Remaining)
	}
	if info.ResetAfter != 1500*time.Millisecond {
		t.Errorf("expected reset after 1.5s, got %v", info.ResetAfter)
	}
	expectedReset := time.Unix(0, int64(1620000000.5*float64(time.Second)))
	if !info.Reset.Equal(expectedReset) {
		t.Errorf("expected reset %v, got %v", expectedReset, info.Reset)
	}
	if !info.Global {
		t.Errorf("expected global true, got %v", info.Global)
	}
}

func TestParseHeaders_Empty(t *testing.T) {
	info := ParseHeaders(http.Header{})
	if info.Bucket != "" {
		t.Errorf("expected empty bucket, got %s", info.Bucket)
	}
	if info.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", info.Remaining)
	}
}

func TestParseHeaders_Invalid(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "abc")
	headers.Set("X-RateLimit-Reset-After", "abc")
	headers.Set("X-RateLimit-Reset", "abc")
	headers.Set("X-RateLimit-Global", "abc")

	info := ParseHeaders(headers)
	if info.Remaining != 0 {
		t.Errorf("expected remaining 0 for invalid input, got %d", info.Remaining)
	}
	if info.ResetAfter != 0 {
		t.Errorf("expected reset after 0 for invalid input, got %v", info.ResetAfter)
	}
	if !info.Reset.IsZero() {
		t.Errorf("expected reset 0 for invalid input, got %v", info.Reset)
	}
	if info.Global {
		t.Errorf("expected global false for invalid input")
	}
}
