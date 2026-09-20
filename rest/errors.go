package rest

import (
	"encoding/json"
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
	// R3VOKE PATCH (2026-09-20): append Discord's field-level error
	// details when present. The Errors map was already decoded from the
	// response body but dropped from the rendered message, which reduced
	// 50035 "Invalid Form Body" to a top-level string with no pointer to
	// the offending field — making command-tree sync failures
	// undiagnosable from logs alone.
	msg := fmt.Sprintf("discord api error: %d (http %d): %s", e.Code, e.HTTPStatus, e.Message)
	if len(e.Errors) > 0 {
		if details, err := json.Marshal(e.Errors); err == nil {
			msg += ": " + string(details)
		}
	}
	return msg
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
