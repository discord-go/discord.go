package http

import (
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client interface defines the methods for an HTTP client.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// sleep is used for testing delays.
var sleep = time.Sleep

// DefaultClient is the default HTTP client implementation.
type DefaultClient struct {
	HTTPClient *http.Client
	UserAgent  string
	Retrier    Retrier
}

// NewClient creates a new DefaultClient with the specified User-Agent.
func NewClient(userAgent string) *DefaultClient {
	return &DefaultClient{
		HTTPClient: &http.Client{
			Transport: &LoggingTransport{
				Base: http.DefaultTransport,
			},
		},
		UserAgent: userAgent,
		Retrier:   NewDefaultRetrier(3),
	}
}

// Do executes an HTTP request, adding the User-Agent and handling retries automatically.
func (c *DefaultClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	var resp *http.Response
	var err error

	maxRetries := 0
	if c.Retrier != nil {
		maxRetries = c.Retrier.MaxRetries()
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Attempt to reconstruct the body if we're retrying
		if attempt > 0 && req.Body != nil {
			if req.GetBody != nil {
				var bErr error
				req.Body, bErr = req.GetBody()
				if bErr != nil {
					return resp, bErr
				}
			} else {
				// We can't retry if we can't reconstruct the request body
				break
			}
		}

		resp, err = c.HTTPClient.Do(req)

		if c.Retrier != nil && c.Retrier.ShouldRetry(resp, err) {
			if attempt < maxRetries {
				if resp != nil && resp.Body != nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				delay := c.Retrier.Backoff(attempt)
				if resp != nil {
					if retryAfter, parseErr := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64); parseErr == nil && retryAfter > 0 {
						delay = time.Duration(retryAfter * float64(time.Second))
					}
				}
				if req.Context() != nil {
					select {
					case <-req.Context().Done():
						return resp, req.Context().Err()
					default:
					}
				}
				if delay > 0 {
					if req.Context() == nil {
						sleep(delay)
					} else {
						timer := time.NewTimer(delay)
						select {
						case <-timer.C:
						case <-req.Context().Done():
							if !timer.Stop() {
								select {
								case <-timer.C:
								default:
								}
							}
							return resp, req.Context().Err()
						}
					}
				}
				continue
			}
		}
		break
	}

	return resp, err
}

// Get issues a GET to the specified URL.
func (c *DefaultClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post issues a POST to the specified URL.
func (c *DefaultClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
