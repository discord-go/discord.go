package backoff

import (
	"math"
	"math/rand/v2"
	"time"
)

// Calculate returns an exponential backoff duration for the given attempt.
//
// The formula is: min(base * 2^attempt + jitter, max)
//
// The jitter is a random duration between 0 and base, added to avoid
// thundering-herd problems when many clients retry simultaneously.
//
// attempt should be 0-indexed (i.e. the first retry is attempt 0).
// base is the starting backoff duration before any exponential scaling.
// max is the upper-bound cap on the returned duration.
//
// If attempt is negative it is treated as 0. If base is zero or negative
// the returned duration will only contain jitter (capped by max). If max
// is zero or negative the function returns 0.
func Calculate(attempt int, base, max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}

	if attempt < 0 {
		attempt = 0
	}

	if base <= 0 {
		base = time.Millisecond
	}

	// Calculate 2^attempt safely, capping at a reasonable value to avoid overflow.
	exp := math.Pow(2, float64(attempt))
	backoff := time.Duration(float64(base) * exp)

	// Guard against overflow: if the multiplication wrapped or exceeded max,
	// clamp to max.
	if backoff <= 0 || backoff > max {
		backoff = max
	}

	// Add jitter: random value in [0, base).
	jitter := time.Duration(rand.Int64N(int64(base)))
	if jitter > max-backoff {
		backoff = max
	} else {
		backoff += jitter
	}

	// Final clamp.
	if backoff > max {
		backoff = max
	}

	return backoff
}
