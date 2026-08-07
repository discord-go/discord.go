package ratelimit

import (
	"context"
	"testing"
	"time"
)

func BenchmarkLimiter_Wait_NoLimits(b *testing.B) {
	l := NewLimiter(NewMemoryStore())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Wait(ctx, "bench-bucket")
	}
}

func BenchmarkLimiter_Update(b *testing.B) {
	l := NewLimiter(NewMemoryStore())
	info := Info{
		Remaining: 100,
		Reset:     time.Now().Add(time.Second),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Update("bench-bucket", info)
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	store := NewMemoryStore()
	store.Put("key", BucketState{Remaining: 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get("key")
	}
}

func BenchmarkMemoryStore_Put(b *testing.B) {
	store := NewMemoryStore()
	state := BucketState{Remaining: 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Put("key", state)
	}
}
