package http

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"
)

// Retrier defines the interface for retrying HTTP requests.
type Retrier interface {
	ShouldRetry(resp *http.Response, err error) bool
	Backoff(attempt int) time.Duration
	MaxRetries() int
}

// DefaultRetrier implements default retry logic for the Discord API.
type DefaultRetrier struct {
	maxRetries int
}

// NewDefaultRetrier creates a new DefaultRetrier.
func NewDefaultRetrier(maxRetries int) *DefaultRetrier {
	return &DefaultRetrier{maxRetries: maxRetries}
}

// ShouldRetry determines if a request should be retried based on errors or status codes.
func (r *DefaultRetrier) ShouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	}
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// Backoff returns the duration to wait before the next attempt using exponential backoff.
func (r *DefaultRetrier) Backoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := math.Pow(2, float64(attempt))
	return time.Duration(delay) * 10 * time.Millisecond // fast backoff for standard usage
}

// MaxRetries returns the maximum number of retries.
func (r *DefaultRetrier) MaxRetries() int {
	return r.maxRetries
}
