package ratelimit

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// These microbenchmarks remove the global token wait only inside the test so
// they measure limiter overhead, not Discord's 50-request-per-second policy.
func benchmarkUnrestrictedLimiter() *limiter {
	l := NewLimiter(NewMemoryStore()).(*limiter)
	l.globalLimiter = rate.NewLimiter(rate.Inf, 1)
	return l
}

func BenchmarkLimiter_Wait_Overhead(b *testing.B) {
	l := benchmarkUnrestrictedLimiter()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Wait(ctx, "bench-bucket")
	}
}

func BenchmarkLimiter_Acquire_KnownBucket(b *testing.B) {
	l := benchmarkUnrestrictedLimiter()
	const path = "/channels/123/messages"
	l.UpdateRoute("GET", path, Info{Bucket: "example", Remaining: 1<<31 - 1, ResetAfter: time.Hour})
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			release, err := l.AcquireRoute(ctx, "GET", path)
			if err != nil {
				b.Error(err)
				return
			}
			release()
		}
	})
}

func BenchmarkLimiter_Acquire_UnknownRoute(b *testing.B) {
	l := benchmarkUnrestrictedLimiter()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		release, err := l.AcquireRoute(ctx, "GET", "/channels/123/messages")
		if err != nil {
			b.Fatal(err)
		}
		release()
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

func BenchmarkLimiter_UpdateRoute(b *testing.B) {
	l := NewLimiter(NewMemoryStore()).(RouteLimiter)
	info := Info{Bucket: "bucket-id", Remaining: 100, ResetAfter: time.Hour}
	const path = "/channels/123456789012345678/messages/987654321012345678"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.UpdateRoute("GET", path, info)
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

var benchmarkMajorResource string

func BenchmarkMajorResource(b *testing.B) {
	for _, route := range []string{"/channels/123456789012345678/messages/987654321012345678", "/webhooks/123456789012345678/secret-token/messages/987654321012345678"} {
		b.Run(route, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkMajorResource = majorResource(route)
			}
		})
	}
}
