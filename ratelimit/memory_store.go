package ratelimit

import "sync"

// MemoryStore is a thread-safe, in-memory implementation of Store.
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]BucketState
}

// NewMemoryStore creates and returns a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: make(map[string]BucketState),
	}
}

// Get retrieves a bucket state from memory.
func (s *MemoryStore) Get(bucketID string) (BucketState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.buckets[bucketID]
	return state, ok
}

// Put saves a bucket state to memory.
func (s *MemoryStore) Put(bucketID string, state BucketState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets[bucketID] = state
}
