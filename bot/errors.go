package bot

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrMissingToken indicates that Start was called without a bot token.
	ErrMissingToken = errors.New("bot: token is required")
	// ErrInvalidToken indicates that the token does not look like a Discord
	// bot token (three dot-separated segments).
	ErrInvalidToken = errors.New("bot: token format is invalid")
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
	Stack []byte
}

func (e *HandlerPanicError) Error() string {
	if len(e.Stack) > 0 {
		return fmt.Sprintf("handler panic during %s: %v\n%s", e.Event, e.Value, e.Stack)
	}
	return fmt.Sprintf("handler panic during %s: %v", e.Event, e.Value)
}

// redactToken replaces any occurrence of the bot token in a string with
// "[REDACTED]". This prevents tokens from leaking into logs when an error
// message contains the Authorization header or a URL that includes the token.
func (b *Bot) redactToken(s string) string {
	b.mu.RLock()
	token := b.token
	b.mu.RUnlock()
	if token == "" {
		return s
	}
	return strings.ReplaceAll(s, token, "[REDACTED]")
}

// isValidBotToken performs a lightweight structural check on a Discord bot
// token. Discord bot tokens consist of three dot-separated segments. This
// does not verify the token with Discord — it only catches obvious typos
// early so users get a clear error instead of an opaque identify failure.
func isValidBotToken(token string) bool {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}
