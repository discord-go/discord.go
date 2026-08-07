package rest

import "context"

type contextKey string

const reasonKey = contextKey("audit_log_reason")

// WithReason returns a new context with the given audit log reason attached.
// This reason will be sent in the X-Audit-Log-Reason header for requests
// that support it.
//
// Discord limits audit log reasons to 512 characters. Reasons longer than
// 512 characters will be rejected by the API with a 400 error.
func WithReason(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, reasonKey, reason)
}

// ReasonFromContext retrieves the audit log reason from the context, if any.
func ReasonFromContext(ctx context.Context) (string, bool) {
	reason, ok := ctx.Value(reasonKey).(string)
	return reason, ok
}
