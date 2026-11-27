package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	gohttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

func (c *Client) waitRoute(ctx context.Context, method, path string) error {
	if limiter, ok := c.Limiter.(ratelimit.RouteLimiter); ok {
		return limiter.WaitRoute(ctx, method, path)
	}
	return c.Limiter.Wait(ctx, path)
}

func (c *Client) acquireRoute(ctx context.Context, method, path string) (func(), error) {
	if limiter, ok := c.Limiter.(ratelimit.RouteReserver); ok {
		return limiter.AcquireRoute(ctx, method, path)
	}
	if err := c.waitRoute(ctx, method, path); err != nil {
		return nil, err
	}
	return func() {}, nil
}

func (c *Client) updateRoute(ctx context.Context, method, path string, info ratelimit.Info) error {
	if limiter, ok := c.Limiter.(ratelimit.ContextRouteUpdater); ok {
		return limiter.UpdateRouteContext(ctx, method, path, info)
	}
	if limiter, ok := c.Limiter.(ratelimit.RouteLimiter); ok {
		limiter.UpdateRoute(method, path, info)
	} else {
		c.Limiter.Update(path, info)
	}
	return nil
}

// readResponseBody avoids allocating an empty buffer for responses that the
// HTTP transport identifies as having no body, such as successful 204s.
func readResponseBody(body io.ReadCloser) ([]byte, error) {
	if body == gohttp.NoBody {
		return []byte{}, nil
	}
	return io.ReadAll(body)
}

// recordLimits applies 429 retry information to the shared limiter before
// retrying, so concurrent requests see the same cooldown.
func (c *Client) recordLimits(ctx context.Context, method, path string, resp *gohttp.Response, body []byte) (time.Duration, error) {
	info := ratelimit.ParseHeaders(resp.Header)
	if resp.StatusCode == 401 || resp.StatusCode == 403 ||
		(resp.StatusCode == 429 && !strings.EqualFold(info.Scope, "shared")) {
		c.markInvalidRequest()
	}
	var delay time.Duration
	if resp.StatusCode == 429 {
		var response struct {
			RetryAfter float64 `json:"retry_after"`
			Global     bool    `json:"global"`
		}
		delay = info.ResetAfter
		if err := json.Unmarshal(body, &response); err == nil && response.RetryAfter > 0 {
			delay = time.Duration(response.RetryAfter * float64(time.Second))
		} else if value, err := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64); err == nil && value > 0 {
			delay = time.Duration(value * float64(time.Second))
		}
		if delay <= 0 {
			delay = time.Second
		}
		info.Global = info.Global || response.Global || strings.EqualFold(info.Scope, "global")
		info.Remaining = 0
		info.ResetAfter = delay
	}
	_, routeAware := c.Limiter.(ratelimit.RouteLimiter)
	if !routeAware || resp.StatusCode == 429 || resp.Header.Get("X-RateLimit-Remaining") != "" || info.Global {
		if err := c.updateRoute(ctx, method, path, info); err != nil {
			return delay, err
		}
	}
	return delay, nil
}

func (c *Client) waitRetry(ctx context.Context, bucket string, delay time.Duration, retries int) error {
	if c.MaxRetries > 0 && retries >= c.MaxRetries {
		return &RateLimitError{Bucket: bucket, RetryAfter: delay, Retries: retries}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
	return c.requestWithDecoder(ctx, method, path, body, authenticate, func(respBody []byte) error {
		if v != nil && len(respBody) > 0 {
			return json.Unmarshal(respBody, v)
		}
		return nil
	})
}

// requestRaw performs an authenticated request and returns the response body
// verbatim instead of JSON-decoding it. Discord returns CSV for a small
// number of routes such as GET /invites/{code}/target-users.
func (c *Client) requestRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	var raw []byte
	err := c.requestWithDecoder(ctx, method, path, body, true, func(respBody []byte) error {
		raw = respBody
		return nil
	})
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *Client) requestWithDecoder(ctx context.Context, method, path string, body any, authenticate bool, decode func([]byte) error) error {
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

	retries := 0
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

		release, err := c.acquireRoute(ctx, method, bucket)
		if err != nil {
			return err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			release()
			return err
		}

		respBody, err := readResponseBody(resp.Body)
		resp.Body.Close()
		waitDuration, limitErr := c.recordLimits(ctx, method, bucket, resp, respBody)
		release()
		if err != nil {
			return err
		}
		if limitErr != nil {
			return limitErr
		}

		if resp.StatusCode == 429 {
			if err := c.waitRetry(ctx, bucket, waitDuration, retries); err != nil {
				return err
			}
			retries++
			continue // Retry the request
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
				// Some endpoints (for example invite target-user CSV
				// validation) return a bare object with Discord's standard
				// error envelope keys absent; keep the body visible instead
				// of reporting an empty message.
				if apiErr.Message == "" && apiErr.Code == 0 {
					apiErr.Message = strings.TrimSpace(string(respBody))
				}
				return &apiErr
			}
			return &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    string(respBody),
			}
		}

		return decode(respBody)
	}
}
