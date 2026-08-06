package sync

import (
	gosync "sync"
	"testing"
	"time"
)

func TestKeyedMutex_LockUnlock(t *testing.T) {
	var km KeyedMutex

	km.Lock("a")
	km.Unlock("a")
	// Should not deadlock.
}

func TestKeyedMutex_DifferentKeysNonBlocking(t *testing.T) {
	var km KeyedMutex

	km.Lock("a")
	defer km.Unlock("a")

	// Locking a different key should not block.
	done := make(chan struct{})
	go func() {
		km.Lock("b")
		km.Unlock("b")
		close(done)
	}()

	select {
	case <-done:
		// OK: different key did not block.
	case <-time.After(time.Second):
		t.Fatal("locking different key blocked unexpectedly")
	}
}

func TestKeyedMutex_SameKeyBlocks(t *testing.T) {
	var km KeyedMutex

	km.Lock("a")

	blocked := make(chan struct{})
	released := make(chan struct{})
	go func() {
		close(blocked)
		km.Lock("a") // Should block until we unlock.
		km.Unlock("a")
		close(released)
	}()

	<-blocked
	// Give the goroutine time to hit the Lock call.
	time.Sleep(50 * time.Millisecond)

	select {
	case <-released:
		t.Fatal("same key lock did not block")
	default:
		// Good: it's still blocking.
	}

	km.Unlock("a")

	select {
	case <-released:
		// OK: unlocking allowed the goroutine to proceed.
	case <-time.After(time.Second):
		t.Fatal("goroutine did not acquire lock after unlock")
	}
}

func TestKeyedMutex_CleansUpEntries(t *testing.T) {
	var km KeyedMutex

	km.Lock("a")
	km.Lock("b")

	if km.Len() != 2 {
		t.Errorf("Len() = %d, want 2", km.Len())
	}

	km.Unlock("a")
	km.Unlock("b")

	if km.Len() != 0 {
		t.Errorf("Len() = %d after unlocking all, want 0", km.Len())
	}
}

func TestKeyedMutex_UnlockPanicsOnUnknownKey(t *testing.T) {
	var km KeyedMutex

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on unlock of unknown key")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", r, r)
		}
		if msg != "sync: unlock of unlocked KeyedMutex key: unknown" {
			t.Errorf("unexpected panic message: %q", msg)
		}
	}()

	km.Unlock("unknown")
}

func TestKeyedMutex_ConcurrentSameKey(t *testing.T) {
	var km KeyedMutex
	var counter int
	var wg gosync.WaitGroup

	const n = 100

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			km.Lock("counter")
			counter++
			km.Unlock("counter")
		}()
	}

	wg.Wait()

	if counter != n {
		t.Errorf("counter = %d, want %d (race condition detected)", counter, n)
	}
}

func TestKeyedMutex_ConcurrentDifferentKeys(t *testing.T) {
	var km KeyedMutex
	var wg gosync.WaitGroup

	keys := []string{"a", "b", "c", "d", "e"}
	counters := make(map[string]int)
	var mu gosync.Mutex

	const perKey = 50

	wg.Add(len(keys) * perKey)
	for _, key := range keys {
		for range perKey {
			go func(k string) {
				defer wg.Done()
				km.Lock(k)
				mu.Lock()
				counters[k]++
				mu.Unlock()
				km.Unlock(k)
			}(key)
		}
	}

	wg.Wait()

	for _, key := range keys {
		if counters[key] != perKey {
			t.Errorf("counter[%s] = %d, want %d", key, counters[key], perKey)
		}
	}
}

func TestKeyedMutex_ZeroValueReady(t *testing.T) {
	// A zero-value KeyedMutex should be ready to use without initialization.
	var km KeyedMutex
	km.Lock("x")
	km.Unlock("x")

	if km.Len() != 0 {
		t.Errorf("Len() = %d after lock/unlock, want 0", km.Len())
	}
}

func TestKeyedMutex_ReuseSameKey(t *testing.T) {
	var km KeyedMutex

	// Lock and unlock the same key multiple times.
	for range 10 {
		km.Lock("reuse")
		km.Unlock("reuse")
	}

	if km.Len() != 0 {
		t.Errorf("Len() = %d after reuse, want 0", km.Len())
	}
}

func TestKeyedMutex_Len(t *testing.T) {
	var km KeyedMutex

	if km.Len() != 0 {
		t.Errorf("initial Len() = %d, want 0", km.Len())
	}

	km.Lock("x")
	if km.Len() != 1 {
		t.Errorf("Len() = %d after locking 1 key, want 1", km.Len())
	}

	km.Lock("y")
	if km.Len() != 2 {
		t.Errorf("Len() = %d after locking 2 keys, want 2", km.Len())
	}

	km.Unlock("x")
	if km.Len() != 1 {
		t.Errorf("Len() = %d after unlocking 1 key, want 1", km.Len())
	}

	km.Unlock("y")
	if km.Len() != 0 {
		t.Errorf("Len() = %d after unlocking all, want 0", km.Len())
	}
}
