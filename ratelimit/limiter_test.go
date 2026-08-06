package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLimiter_NoLimits(t *testing.T) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()

	start := time.Now()
	err := l.Wait(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Fatal("expected no delay")
	}
}

func TestLimiter_EmptyBucket(t *testing.T) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()

	err := l.Wait(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLimiter_GlobalLimit(t *testing.T) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()

	l.Update("", Info{
		Global: true,
		Reset:  time.Now().Add(50 * time.Millisecond),
	})

	start := time.Now()
	err := l.Wait(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("expected global limit to delay request")
	}
}

func TestLimiter_BucketLimit(t *testing.T) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()

	l.Update("b1", Info{
		Remaining: 0,
		Reset:     time.Now().Add(50 * time.Millisecond),
	})

	start := time.Now()
	err := l.Wait(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("expected bucket limit to delay request")
	}
}

func TestLimiter_BucketLimitResetAfter(t *testing.T) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()

	l.Update("b1", Info{
		Remaining:  0,
		ResetAfter: 50 * time.Millisecond,
	})

	start := time.Now()
	err := l.Wait(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("expected bucket limit to delay request via ResetAfter")
	}
}

func TestLimiter_ContextCanceled_Global(t *testing.T) {
	l := NewLimiter(NewMemoryStore())

	l.Update("", Info{
		Global: true,
		Reset:  time.Now().Add(1 * time.Second),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := l.Wait(ctx, "b1")
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestLimiter_ContextCanceled_Bucket(t *testing.T) {
	l := NewLimiter(NewMemoryStore())

	l.Update("b1", Info{
		Remaining: 0,
		Reset:     time.Now().Add(1 * time.Second),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := l.Wait(ctx, "b1")
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestLimiter_DecrementRemaining(t *testing.T) {
	store := NewMemoryStore()
	l := NewLimiter(store)
	ctx := context.Background()

	l.Update("b1", Info{
		Remaining: 2,
		Reset:     time.Now().Add(1 * time.Second),
	})

	l.Wait(ctx, "b1")
	state, _ := store.Get("b1")
	if state.Remaining != 1 {
		t.Errorf("expected 1 remaining, got %d", state.Remaining)
	}

	l.Wait(ctx, "b1")
	state, _ = store.Get("b1")
	if state.Remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", state.Remaining)
	}
}

func TestLimiter_GlobalUpdateOlderIgnored(t *testing.T) {
	l := NewLimiter(NewMemoryStore())

	now := time.Now()
	l.Update("", Info{
		Global: true,
		Reset:  now.Add(1 * time.Second),
	})

	l.Update("", Info{
		Global: true,
		Reset:  now,
	})

	lim := l.(*limiter)
	lim.globalMu.RLock()
	defer lim.globalMu.RUnlock()
	if !lim.globalReset.Equal(now.Add(1 * time.Second)) {
		t.Fatal("expected global reset to remain unchanged")
	}
}

func TestLimiter_ConcurrentRequests(t *testing.T) {
	store := NewMemoryStore()
	l := NewLimiter(store)

	l.Update("b1", Info{
		Remaining: 5,
		Reset:     time.Now().Add(100 * time.Millisecond),
	})

	var wg sync.WaitGroup
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Wait(ctx, "b1")
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	if duration < 100*time.Millisecond {
		t.Fatalf("expected duration >= 100ms, got %v", duration)
	}
}

func TestLimiter_WaitGlobalUpdateWhileWaiting(t *testing.T) {
	store := NewMemoryStore()
	l := NewLimiter(store)

	lim := l.(*limiter)
	mu := lim.getBucketLock("b1")
	mu.Lock()

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		err := l.Wait(ctx, "b1")
		if err != context.Canceled {
			t.Errorf("expected context canceled, got %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	l.Update("", Info{
		Global: true,
		Reset:  time.Now().Add(1 * time.Second),
	})

	cancel()
	mu.Unlock()
	wg.Wait()
}
