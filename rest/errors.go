package rest

import (
	"fmt"
	"time"
)

// RateLimitError is returned when a request exhausts the client's retry
// budget. It carries the bucket, the final retry-after duration, and the
// number of retries already performed. A caller that receives it should
// wait at least RetryAfter before attempting the request again.
type RateLimitError struct {
	Bucket     string
	RetryAfter time.Duration
	Retries    int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited after %d retries (bucket %s): retry after %s", e.Retries, e.Bucket, e.RetryAfter)
}

// APIError represents an error returned by the Discord API.
type APIError struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	Errors     map[string]any `json:"errors,omitempty"`
	HTTPStatus int            `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("discord api error: %d (http %d): %s", e.Code, e.HTTPStatus, e.Message)
}

// CaptchaError represents a CAPTCHA challenge from the Discord API.
type CaptchaError struct {
	APIError
	CaptchaKey     any    `json:"captcha_key,omitempty"`
	CaptchaSitekey string `json:"captcha_sitekey,omitempty"`
	CaptchaService string `json:"captcha_service,omitempty"`
	CaptchaRqdata  string `json:"captcha_rqdata,omitempty"`
	CaptchaRqtoken string `json:"captcha_rqtoken,omitempty"`
}

func (e *CaptchaError) Error() string {
	return fmt.Sprintf("captcha required: %s (service: %s, sitekey: %s)", e.Message, e.CaptchaService, e.CaptchaSitekey)
}
