// Package sync provides a KeyedMutex that locks per-key.
//
// This is an internal package and MUST NOT be imported by external consumers.
// It is useful for per-bucket locking in the rate limiter, where each rate
// limit bucket needs its own independent lock.
package sync
