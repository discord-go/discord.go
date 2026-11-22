package ratelimit

import (
	"context"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter manages Discord API rate limits, blocking requests when buckets are exhausted.
type Limiter interface {
	// Wait blocks until a request can be made for the given bucket.
	Wait(ctx context.Context, bucket string) error

	// Update updates the rate limit information from a response.
	Update(bucket string, info Info)
}

// StateReader is implemented by Limiters that can report the current state
// of a bucket. The concrete limiter returned by NewLimiter implements it;
// custom limiters may opt in.
type StateReader interface {
	// State returns the current state of a bucket, if known.
	State(bucket string) (BucketState, bool)
}

// RouteLimiter lets a REST client distinguish HTTP methods while retaining
// compatibility with implementations of Limiter that accept only a path.
type RouteLimiter interface {
	WaitRoute(ctx context.Context, method, path string) error
	UpdateRoute(method, path string, info Info)
}

type limiter struct {
	store Store

	globalMu    sync.RWMutex
	globalReset time.Time

	// globalLimiter enforces Discord's 50 requests/second global limit
	// using a token bucket. This is more efficient than a mutex-protected
	// ring buffer because Wait does not hold a lock while sleeping —
	// concurrent goroutines can reserve tokens in parallel.
	globalLimiter *rate.Limiter

	bucketLocks sync.Map // map[string]*sync.Mutex
	routeToHash sync.Map // map[string]string
	lastRoute   sync.Map // map[path]route, for the legacy State method
}

// NewLimiter creates a new rate limiter using the provided store.
func NewLimiter(store Store) Limiter {
	if store == nil {
		store = NewMemoryStore()
	}
	return &limiter{
		store:         store,
		globalLimiter: rate.NewLimiter(rate.Limit(50), 50),
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

	// Enforce 50 requests/second global limit using a token bucket.
	// rate.Limiter.Wait blocks until a token is available, without
	// holding a global mutex during the wait.
	if err := l.globalLimiter.Wait(ctx); err != nil {
		return err
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

	// 3. Check global limit again in case it was updated while waiting for bucket lock
	l.globalMu.RLock()
	globalReset = l.globalReset
	l.globalMu.RUnlock()

	if now := time.Now(); globalReset.After(now) {
		mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(globalReset.Sub(now)):
		}
		return l.waitBucket(ctx, storeKey)
	}
	return l.waitBucketLocked(ctx, storeKey, mu)
}

func (l *limiter) waitBucket(ctx context.Context, key string) error {
	mu := l.getBucketLock(key)
	mu.Lock()
	return l.waitBucketLocked(ctx, key, mu)
}

func (l *limiter) waitBucketLocked(ctx context.Context, key string, mu *sync.Mutex) error {
	for {
		l.globalMu.RLock()
		globalReset := l.globalReset
		l.globalMu.RUnlock()
		if delay := time.Until(globalReset); delay > 0 {
			mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			mu.Lock()
			continue
		}
		state, ok := l.store.Get(key)
		if !ok {
			mu.Unlock()
			return nil
		}
		now := time.Now()
		if state.Remaining > 0 || !state.Reset.After(now) {
			if state.Remaining > 0 {
				state.Remaining--
			}
			l.store.Put(key, state)
			mu.Unlock()
			return nil
		}
		delay := state.Reset.Sub(now)
		mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		mu.Lock()
	}
}

// State returns the current state of a bucket, if known.
func (l *limiter) State(bucket string) (BucketState, bool) {
	if bucket == "" {
		return BucketState{}, false
	}
	storeKey := bucket
	if route, ok := l.lastRoute.Load(bucket); ok {
		storeKey = route.(string)
	}
	if hash, ok := l.routeToHash.Load(storeKey); ok {
		storeKey = hash.(string)
	}
	return l.store.Get(storeKey)
}

func (l *limiter) Update(bucket string, info Info) {
	l.update(bucket, majorResource(bucket), info)
}

// WaitRoute waits for the method and path's rate limit bucket.
func (l *limiter) WaitRoute(ctx context.Context, method, path string) error {
	return l.Wait(ctx, strings.ToUpper(method)+" "+path)
}

// UpdateRoute records response limits for the method and path.
func (l *limiter) UpdateRoute(method, path string, info Info) {
	route := strings.ToUpper(method) + " " + path
	l.update(route, majorResource(path), info)
	l.lastRoute.Store(path, route)
}

func majorResource(path string) string {
	if i := strings.IndexByte(path, ' '); i >= 0 {
		path = path[i+1:]
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	switch parts[0] {
	case "channels", "guilds":
		return parts[0] + "/" + parts[1]
	case "webhooks":
		if len(parts) > 2 {
			return parts[0] + "/" + parts[1] + "/" + parts[2]
		}
		return parts[0] + "/" + parts[1]
	}
	return ""
}

func (l *limiter) update(bucket, major string, info Info) {
	reset := info.Reset
	if info.ResetAfter > 0 {
		reset = time.Now().Add(info.ResetAfter)
	}
	if info.Global {
		l.globalMu.Lock()
		if reset.After(l.globalReset) {
			l.globalReset = reset
		}
		l.globalMu.Unlock()
	}

	if bucket != "" {
		storeKey := bucket
		if info.Bucket != "" {
			storeKey = "bucket:" + info.Bucket + ":" + major
			l.routeToHash.Store(bucket, storeKey)
		} else if hash, ok := l.routeToHash.Load(bucket); ok {
			storeKey = hash.(string)
		}

		mu := l.getBucketLock(storeKey)
		mu.Lock()
		defer mu.Unlock()

		state := BucketState{
			Remaining: info.Remaining,
			Reset:     reset,
		}
		// In-flight responses can arrive out of order. Keep the lower
		// remaining count until the known window has expired.
		if previous, ok := l.store.Get(storeKey); ok && previous.Reset.After(time.Now()) {
			if previous.Remaining < state.Remaining {
				state.Remaining = previous.Remaining
			}
			if previous.Reset.After(state.Reset) {
				state.Reset = previous.Reset
			}
		}

		l.store.Put(storeKey, state)
	}
}
