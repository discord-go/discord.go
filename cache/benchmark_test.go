package cache

import (
	"testing"
	"time"
)

func BenchmarkMemoryCache_Get(b *testing.B) {
	c := NewMemoryCache()
	c.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get("key")
	}
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	c := NewMemoryCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", "value")
	}
}

func BenchmarkMemoryCache_GetGuild(b *testing.B) {
	c := NewMemoryCache()
	c.SetGuild("123", "guild")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetGuild("123")
	}
}

func BenchmarkMemoryCache_SetGuild(b *testing.B) {
	c := NewMemoryCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetGuild("123", "guild")
	}
}

// BenchmarkMemoryCache_GetParallel models concurrent reads of a cached value.
func BenchmarkMemoryCache_GetParallel(b *testing.B) {
	c := NewMemoryCache()
	c.Set("test", "value")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Get("test")
		}
	})
}

// BenchmarkMemoryCache_MixedParallel models a read-heavy cache with TTL.
func BenchmarkMemoryCache_MixedParallel(b *testing.B) {
	c := NewMemoryCache(WithTTL(time.Hour))
	c.Set("test", "value")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var operations int
		for pb.Next() {
			operations++
			if operations%20 == 0 {
				c.Set("test", "value")
			} else {
				_, _ = c.Get("test")
			}
		}
	})
}
