package rest

import (
	"bytes"
	"context"
	"io"
	gohttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"errors"
	"github.com/discord-go/discord.go/json"
	"github.com/discord-go/discord.go/ratelimit"
)

// ErrInvalidRequestLimitExceeded is returned when too many invalid requests are made.
var ErrInvalidRequestLimitExceeded = errors.New("blocked to prevent Cloudflare IP ban: too many invalid requests")

func (c *Client) checkInvalidRequests() error {
	c.invalidMu.Lock()
	defer c.invalidMu.Unlock()

	oldest := c.invalidTimestamps[c.invalidIdx]
	if !oldest.IsZero() && time.Since(oldest) < 10*time.Minute {
		return ErrInvalidRequestLimitExceeded
	}
	return nil
}

func (c *Client) markInvalidRequest() {
	c.invalidMu.Lock()
	defer c.invalidMu.Unlock()
	c.invalidTimestamps[c.invalidIdx] = time.Now()
	c.invalidIdx = (c.invalidIdx + 1) % 9500
}

// Request performs an HTTP request to the Discord API.
func (c *Client) Request(ctx context.Context, method, path string, body any, v any) error {
	return c.request(ctx, method, path, body, v, true)
}

// RequestNoAuth performs a request without an authorization header.
func (c *Client) RequestNoAuth(ctx context.Context, method, path string, body any, v any) error {
	return c.request(ctx, method, path, body, v, false)
}

func (c *Client) request(ctx context.Context, method, path string, body any, v any, authenticate bool) error {
	requestURL := c.BaseURL + path

	var reqBodyBytes []byte
	if body != nil {
		var err error
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	bucket := path
	if idx := strings.Index(path, "?"); idx != -1 {
		bucket = path[:idx]
	}

	for {
		if err := c.checkInvalidRequests(); err != nil {
			return err
		}

		var reqBody io.Reader
		if reqBodyBytes != nil {
			reqBody = bytes.NewReader(reqBodyBytes)
		}

		req, err := gohttp.NewRequestWithContext(ctx, method, requestURL, reqBody)
		if err != nil {
			return err
		}

		if authenticate && c.AuthMode != AuthNone && c.token != "" {
			req.Header.Set("Authorization", string(c.AuthMode)+" "+c.token)
		}
		if reqBodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if reason, ok := ReasonFromContext(ctx); ok {
			req.Header.Set("X-Audit-Log-Reason", url.QueryEscape(reason))
		}

		if err := c.Limiter.Wait(ctx, bucket); err != nil {
			return err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 429 {
			c.markInvalidRequest()
		}

		info := ratelimit.ParseHeaders(resp.Header)
		c.Limiter.Update(bucket, info)

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode == 429 {
			var rateLimitErr struct {
				RetryAfter float64 `json:"retry_after"`
			}
			waitDuration := info.ResetAfter
			if err := json.Unmarshal(respBody, &rateLimitErr); err == nil && rateLimitErr.RetryAfter > 0 {
				waitDuration = time.Duration(rateLimitErr.RetryAfter * float64(time.Second))
			} else if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if parsed, err := strconv.ParseFloat(retryAfter, 64); err == nil {
					waitDuration = time.Duration(parsed * float64(time.Second))
				}
			}

			if waitDuration > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(waitDuration):
				}
				continue // Retry the request
			}
		}

		if resp.StatusCode >= 400 {
			var captchaErr CaptchaError
			if err := json.Unmarshal(respBody, &captchaErr); err == nil && (captchaErr.CaptchaSitekey != "" || captchaErr.CaptchaService != "" || captchaErr.CaptchaKey != nil) {
				captchaErr.HTTPStatus = resp.StatusCode
				return &captchaErr
			}

			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				apiErr.HTTPStatus = resp.StatusCode
				return &apiErr
			}
			return &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    string(respBody),
			}
		}

		if v != nil && len(respBody) > 0 {
			return json.Unmarshal(respBody, v)
		}

		return nil
	}
}
