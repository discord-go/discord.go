// Package bucket provides route hash calculation for Discord rate limit buckets.
//
// This is an internal package and MUST NOT be imported by external consumers.
// Discord's rate limiting groups certain routes into the same bucket; this
// package computes a deterministic hash from the HTTP method and path so the
// rate limiter can key on it.
package bucket
