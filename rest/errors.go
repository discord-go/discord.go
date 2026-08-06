package rest

import "fmt"

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
