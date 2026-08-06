package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter manages Discord API rate limits, blocking requests when buckets are exhausted.
type Limiter interface {
	// Wait blocks until a request can be made for the given bucket.
	Wait(ctx context.Context, bucket string) error

	// Update updates the rate limit information from a response.
	Update(bucket string, info Info)
}

type limiter struct {
	store Store

	globalMu    sync.RWMutex
	globalReset time.Time

	globalTokensMu sync.Mutex
	reqTimestamps  [50]time.Time
	reqIdx         int

	bucketLocks sync.Map // map[string]*sync.Mutex
	routeToHash sync.Map // map[string]string
}

// NewLimiter creates a new rate limiter using the provided store.
func NewLimiter(store Store) Limiter {
	if store == nil {
		store = NewMemoryStore()
	}
	return &limiter{
		store: store,
	}
}

func (l *limiter) getBucketLock(bucket string) *sync.Mutex {
	v, _ := l.bucketLocks.LoadOrStore(bucket, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (l *limiter) Wait(ctx context.Context, bucket string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// 1. Check global rate limit before locking bucket
	l.globalMu.RLock()
	globalReset := l.globalReset
	l.globalMu.RUnlock()

	if now := time.Now(); globalReset.After(now) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(globalReset.Sub(now)):
		}
	}

	// Enforce 50 requests/second global limit
	for {
		l.globalTokensMu.Lock()
		tNow := time.Now()
		oldest := l.reqTimestamps[l.reqIdx]
		var waitDuration time.Duration
		if !oldest.IsZero() && tNow.Sub(oldest) < time.Second {
			waitDuration = time.Second - tNow.Sub(oldest)
		}
		if waitDuration <= 0 {
			l.reqTimestamps[l.reqIdx] = tNow
			l.reqIdx = (l.reqIdx + 1) % len(l.reqTimestamps)
			l.globalTokensMu.Unlock()
			break
		}
		l.globalTokensMu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}

	if bucket == "" {
		return nil
	}

	// 2. Lock the specific bucket
	storeKey := bucket
	if hash, ok := l.routeToHash.Load(bucket); ok {
		storeKey = hash.(string)
	}

	mu := l.getBucketLock(storeKey)
	mu.Lock()
	defer mu.Unlock()

	// 3. Check global limit again in case it was updated while waiting for bucket lock
	l.globalMu.RLock()
	globalReset = l.globalReset
	l.globalMu.RUnlock()

	if now := time.Now(); globalReset.After(now) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(globalReset.Sub(now)):
		}
	}

	// 4. Check bucket limits
	state, ok := l.store.Get(storeKey)
	if !ok {
		return nil // Unknown bucket, proceed
	}

	now := time.Now()
	if state.Remaining <= 0 && state.Reset.After(now) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(state.Reset.Sub(now)):
		}
	}

	// 5. Decrement optimistic remaining count
	if state.Remaining > 0 {
		state.Remaining--
	}
	l.store.Put(storeKey, state)

	return nil
}

func (l *limiter) Update(bucket string, info Info) {
	if info.Global {
		l.globalMu.Lock()
		if info.Reset.After(l.globalReset) {
			l.globalReset = info.Reset
		}
		l.globalMu.Unlock()
	}

	if bucket != "" {
		storeKey := bucket
		if info.Bucket != "" {
			storeKey = "bucket:" + info.Bucket
			l.routeToHash.Store(bucket, storeKey)
		}

		mu := l.getBucketLock(storeKey)
		mu.Lock()
		defer mu.Unlock()

		state := BucketState{
			Remaining: info.Remaining,
			Reset:     info.Reset,
		}

		if info.ResetAfter > 0 {
			state.Reset = time.Now().Add(info.ResetAfter)
		}

		l.store.Put(storeKey, state)
	}
}
