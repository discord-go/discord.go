package bot

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingToken indicates that Start was called without a bot token.
	ErrMissingToken = errors.New("bot: token is required")
	// ErrBotAlreadyRunning indicates that Start was called for an active bot.
	ErrBotAlreadyRunning = errors.New("bot: already running")
	// ErrBotNotRunning indicates that Wait was called before Start.
	ErrBotNotRunning = errors.New("bot: not running")
	// ErrInteractionAlreadyResponded prevents accidental double acknowledgements.
	ErrInteractionAlreadyResponded = errors.New("bot: interaction already responded")
)

// HandlerPanicError describes a panic recovered from a user handler.
type HandlerPanicError struct {
	Event string
	Value interface{}
}

func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("handler panic during %s: %v", e.Event, e.Value)
}
