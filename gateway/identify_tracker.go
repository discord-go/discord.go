package gateway

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrIdentifyRateLimit is returned when the global IDENTIFY rate limit is exceeded.
var ErrIdentifyRateLimit = errors.New("gateway: identify rate limit exceeded (1000 per 24h)")

// IdentifyTracker tracks the number of IDENTIFY calls to prevent exceeding
// the global limit of 1000 calls per 24 hours.
type IdentifyTracker struct {
	mu       sync.Mutex
	attempts []time.Time
}

// NewIdentifyTracker creates a new IdentifyTracker.
func NewIdentifyTracker() *IdentifyTracker {
	return &IdentifyTracker{
		attempts: make([]time.Time, 0, 1000),
	}
}

// Wait checks if an IDENTIFY call can be made. It returns an error if the
// limit is exceeded.
// Alternatively, it could block, but erroring is safer to avoid long deadlocks.
func (t *IdentifyTracker) Wait(ctx context.Context) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-24 * time.Hour)

	// Prune old attempts
	var newAttempts []time.Time
	for _, attempt := range t.attempts {
		if attempt.After(windowStart) {
			newAttempts = append(newAttempts, attempt)
		}
	}
	t.attempts = newAttempts

	if len(t.attempts) >= 1000 {
		return ErrIdentifyRateLimit
	}

	t.attempts = append(t.attempts, now)
	return nil
}
