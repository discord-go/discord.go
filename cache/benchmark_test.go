package cache

import "testing"

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
