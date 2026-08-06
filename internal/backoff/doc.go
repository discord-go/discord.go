// Package backoff provides a reusable exponential backoff calculator with jitter.
//
// This is an internal package and MUST NOT be imported by external consumers.
// It is used by the Gateway reconnection logic and the rate limiter to compute
// retry delays that avoid thundering-herd problems.
package backoff
