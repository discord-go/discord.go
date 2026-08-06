package http

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestDefaultRetrier_ShouldRetry(t *testing.T) {
	retrier := NewDefaultRetrier(3)

	// Test nil error, nil response
	if retrier.ShouldRetry(nil, nil) {
		t.Error("should not retry nil error and nil response")
	}

	// Test non-nil error
	if !retrier.ShouldRetry(nil, errors.New("network err")) {
		t.Error("should retry on network error")
	}

	// Test status codes
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusNotFound, false},
		{http.StatusRequestTimeout, true},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		resp := &http.Response{StatusCode: tt.statusCode}
		if retrier.ShouldRetry(resp, nil) != tt.expected {
			t.Errorf("expected ShouldRetry to be %v for status %d", tt.expected, tt.statusCode)
		}
	}
}

func TestDefaultRetrier_Backoff(t *testing.T) {
	retrier := NewDefaultRetrier(3)

	if retrier.Backoff(-1) != 10*time.Millisecond {
		t.Errorf("expected 10ms for attempt -1")
	}
	if retrier.Backoff(0) != 10*time.Millisecond {
		t.Errorf("expected 10ms for attempt 0")
	}
	if retrier.Backoff(1) != 20*time.Millisecond {
		t.Errorf("expected 20ms for attempt 1")
	}
	if retrier.Backoff(2) != 40*time.Millisecond {
		t.Errorf("expected 40ms for attempt 2")
	}
}

func TestDefaultRetrier_MaxRetries(t *testing.T) {
	retrier := NewDefaultRetrier(5)
	if retrier.MaxRetries() != 5 {
		t.Errorf("expected max retries 5, got %d", retrier.MaxRetries())
	}
}
