package gateway

import "errors"

var (
	ErrReconnectRequested = errors.New("reconnect requested by gateway")
	ErrInvalidSession     = errors.New("invalid session")
	ErrFatalClose         = errors.New("fatal close code received")
)
