package cache

import "time"

// Options contains configuration for the cache.
type Options struct {
	// TTL is the time-to-live for cache entries. A value of 0 means no expiration.
	TTL time.Duration
	// MaxSize is the maximum number of items the cache can hold. 0 means unlimited.
	MaxSize int
}

// Option is a functional option for configuring a Cache.
type Option func(*Options)

// WithTTL sets the TTL for the cache.
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithMaxSize sets the maximum size for the cache.
func WithMaxSize(size int) Option {
	return func(o *Options) {
		o.MaxSize = size
	}
}

// DefaultOptions returns the default cache options.
func DefaultOptions() *Options {
	return &Options{
		TTL:     0,
		MaxSize: 0, // 0 = unlimited
	}
}
