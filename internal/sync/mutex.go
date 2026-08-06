package sync

import "sync"

// KeyedMutex provides per-key mutual exclusion. Each unique key gets its own
// lock, so operations on different keys can proceed concurrently while
// operations on the same key are serialized.
//
// This is useful for per-bucket locking in the rate limiter, where each
// rate limit bucket needs its own independent lock.
//
// A zero-value KeyedMutex is ready to use.
type KeyedMutex struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
}

// lockEntry tracks a per-key mutex and its current reference count.
type lockEntry struct {
	mu      sync.Mutex
	waiters int
}

// Lock acquires the lock for the given key. If another goroutine holds the
// lock for the same key, Lock blocks until it is available. Locks on
// different keys do not contend with each other.
func (km *KeyedMutex) Lock(key string) {
	km.mu.Lock()
	if km.locks == nil {
		km.locks = make(map[string]*lockEntry)
	}
	entry, ok := km.locks[key]
	if !ok {
		entry = &lockEntry{}
		km.locks[key] = entry
	}
	entry.waiters++
	km.mu.Unlock()

	entry.mu.Lock()
}

// Unlock releases the lock for the given key. If no other goroutine is
// waiting for this key, the internal entry is cleaned up to avoid unbounded
// memory growth.
//
// Unlock panics if the key was not previously locked.
func (km *KeyedMutex) Unlock(key string) {
	km.mu.Lock()
	entry, ok := km.locks[key]
	if !ok {
		km.mu.Unlock()
		panic("sync: unlock of unlocked KeyedMutex key: " + key)
	}
	entry.waiters--
	if entry.waiters == 0 {
		delete(km.locks, key)
	}
	km.mu.Unlock()

	entry.mu.Unlock()
}

// Len returns the number of keys that currently have active locks.
// This is primarily useful for testing and diagnostics.
func (km *KeyedMutex) Len() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return len(km.locks)
}
