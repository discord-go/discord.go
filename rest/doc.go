// Package rest provides a client for interacting with the Discord REST API.
package rest

// StringPtr returns a pointer to the given string. It is a convenience
// helper for the many REST parameter fields that use *string (for example,
// EditMessageParams.Content, ModifyChannelParams.Name, and ModifyChannelParams.Topic)
// so callers can write rest.StringPtr("hello") instead of taking the address
// of a local variable.
func StringPtr(s string) *string {
	return &s
}
