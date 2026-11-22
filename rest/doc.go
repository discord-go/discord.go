// Package rest provides a client for the Discord REST API.
//
// New configures a default HTTP client and rate limiter when they are nil.
// Requests accept a context for cancellation. Set Client.MaxRetries to bound
// retries after rate-limit responses; zero leaves retries unbounded.
package rest

// StringPtr returns a pointer to the given string. It is a convenience
// helper for the many REST parameter fields that use *string (for example,
// EditMessageParams.Content, ModifyChannelParams.Name, and ModifyChannelParams.Topic)
// so callers can write rest.StringPtr("hello") instead of taking the address
// of a local variable.
func StringPtr(s string) *string {
	return &s
}
