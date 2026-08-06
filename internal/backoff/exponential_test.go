package backoff

import (
	"testing"
	"time"
)

func TestCalculate_BasicExponential(t *testing.T) {
	base := 100 * time.Millisecond
	max := 10 * time.Second

	// Attempt 0: backoff should be in [base, 2*base) due to jitter.
	for range 50 {
		d := Calculate(0, base, max)
		if d < base || d >= 2*base {
			t.Errorf("attempt 0: got %v, want in [%v, %v)", d, base, 2*base)
		}
	}

	// Attempt 1: backoff should be in [2*base, 3*base).
	for range 50 {
		d := Calculate(1, base, max)
		if d < 2*base || d >= 3*base {
			t.Errorf("attempt 1: got %v, want in [%v, %v)", d, 2*base, 3*base)
		}
	}

	// Attempt 2: backoff should be in [4*base, 5*base).
	for range 50 {
		d := Calculate(2, base, max)
		if d < 4*base || d >= 5*base {
			t.Errorf("attempt 2: got %v, want in [%v, %v)", d, 4*base, 5*base)
		}
	}
}

func TestCalculate_CapsAtMax(t *testing.T) {
	base := 100 * time.Millisecond
	max := 500 * time.Millisecond

	// A very high attempt should be capped at max.
	for range 50 {
		d := Calculate(100, base, max)
		if d > max {
			t.Errorf("attempt 100: got %v, want <= %v", d, max)
		}
		if d <= 0 {
			t.Errorf("attempt 100: got %v, want > 0", d)
		}
	}
}

func TestCalculate_NegativeAttempt(t *testing.T) {
	base := 100 * time.Millisecond
	max := 10 * time.Second

	// Negative attempt should be treated as attempt 0.
	for range 20 {
		d := Calculate(-5, base, max)
		if d < base || d >= 2*base {
			t.Errorf("negative attempt: got %v, want in [%v, %v)", d, base, 2*base)
		}
	}
}

func TestCalculate_ZeroMax(t *testing.T) {
	d := Calculate(0, 100*time.Millisecond, 0)
	if d != 0 {
		t.Errorf("zero max: got %v, want 0", d)
	}
}

func TestCalculate_NegativeMax(t *testing.T) {
	d := Calculate(0, 100*time.Millisecond, -1*time.Second)
	if d != 0 {
		t.Errorf("negative max: got %v, want 0", d)
	}
}

func TestCalculate_ZeroBase(t *testing.T) {
	max := 10 * time.Second

	// Zero base gets normalized to 1ms.
	d := Calculate(0, 0, max)
	if d <= 0 || d > max {
		t.Errorf("zero base: got %v, want in (0, %v]", d, max)
	}
}

func TestCalculate_NegativeBase(t *testing.T) {
	max := 10 * time.Second

	// Negative base gets normalized to 1ms.
	d := Calculate(0, -100*time.Millisecond, max)
	if d <= 0 || d > max {
		t.Errorf("negative base: got %v, want in (0, %v]", d, max)
	}
}

func TestCalculate_JitterVaries(t *testing.T) {
	base := 1 * time.Second
	max := 10 * time.Second

	// Run many iterations and check that we get at least 2 distinct values
	// (jitter should introduce randomness).
	seen := make(map[time.Duration]bool)
	for range 100 {
		d := Calculate(0, base, max)
		seen[d] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected jitter to produce varied results, got %d distinct values", len(seen))
	}
}

func TestCalculate_AlwaysPositive(t *testing.T) {
	base := 50 * time.Millisecond
	max := 30 * time.Second

	for attempt := range 20 {
		d := Calculate(attempt, base, max)
		if d <= 0 {
			t.Errorf("attempt %d: got %v, want > 0", attempt, d)
		}
		if d > max {
			t.Errorf("attempt %d: got %v, want <= %v", attempt, d, max)
		}
	}
}
