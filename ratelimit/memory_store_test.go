package ratelimit

import "testing"

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	_, ok := store.Get("b1")
	if ok {
		t.Error("expected not ok for missing bucket")
	}

	state := BucketState{Remaining: 10}
	store.Put("b1", state)

	got, ok := store.Get("b1")
	if !ok {
		t.Error("expected ok for existing bucket")
	}
	if got.Remaining != 10 {
		t.Errorf("expected remaining 10, got %d", got.Remaining)
	}
}
