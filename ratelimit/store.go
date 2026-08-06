package ratelimit

import "time"

// BucketState represents the current state of a rate limit bucket.
type BucketState struct {
	Remaining int
	Reset     time.Time
}

// Store defines an interface for saving and retrieving rate limit bucket states.
type Store interface {
	// Get retrieves the current state of a bucket.
	// Returns true if the bucket state exists, false otherwise.
	Get(bucketID string) (BucketState, bool)

	// Put saves the state of a bucket.
	Put(bucketID string, state BucketState)
}
