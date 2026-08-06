// Package storage contains small persistence primitives that can back bot
// settings, statistics, moderation records, and other application state.
// Database-specific adapters can implement Store without changing bot code.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("storage: record not found")

// Store is a JSON-oriented persistence boundary for application data.
type Store interface {
	Get(ctx context.Context, key string, target any) error
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, prefix string) ([]string, error)
}

type memoryRecord struct {
	data      json.RawMessage
	expiresAt time.Time
}

// MemoryStore is a concurrency-safe Store suitable for tests and small bots.
type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]memoryRecord
	TTL     time.Duration
}

func NewMemoryStore(ttl time.Duration) *MemoryStore {
	return &MemoryStore{records: make(map[string]memoryRecord), TTL: ttl}
}

func (s *MemoryStore) Get(ctx context.Context, key string, target any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	s.mu.RLock()
	record, ok := s.records[key]
	s.mu.RUnlock()
	if !ok || (!record.expiresAt.IsZero() && time.Now().After(record.expiresAt)) {
		return ErrNotFound
	}
	return json.Unmarshal(record.data, target)
}

func (s *MemoryStore) Set(ctx context.Context, key string, value any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	record := memoryRecord{data: data}
	if s.TTL > 0 {
		record.expiresAt = time.Now().Add(s.TTL)
	}
	s.mu.Lock()
	s.records[key] = record
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.records, key)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Keys(ctx context.Context, prefix string) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	now := time.Now()
	s.mu.RLock()
	keys := make([]string, 0)
	for key, record := range s.records {
		if strings.HasPrefix(key, prefix) && (record.expiresAt.IsZero() || now.Before(record.expiresAt)) {
			keys = append(keys, key)
		}
	}
	s.mu.RUnlock()
	sort.Strings(keys)
	return keys, nil
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
