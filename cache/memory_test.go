package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryCache_GetSetDeleteClear(t *testing.T) {
	c := NewMemoryCache()

	// Test Get empty
	if _, ok := c.Get("test"); ok {
		t.Error("Expected false, got true")
	}

	// Test Set and Get
	c.Set("test", "value")
	val, ok := c.Get("test")
	if !ok {
		t.Error("Expected true, got false")
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}

	// Test Delete
	c.Delete("test")
	if _, ok := c.Get("test"); ok {
		t.Error("Expected false, got true")
	}

	// Test Clear
	c.Set("test2", "value2")
	c.Clear()
	if _, ok := c.Get("test2"); ok {
		t.Error("Expected false after Clear, got true")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	c := NewMemoryCache(WithTTL(50 * time.Millisecond))
	c.Set("test", "value")

	// Should exist immediately
	if _, ok := c.Get("test"); !ok {
		t.Error("Expected true immediately after Set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not exist after TTL
	if _, ok := c.Get("test"); ok {
		t.Error("Expected false after TTL expiration")
	}

	// Test CleanUp
	c.Set("test2", "value2")
	time.Sleep(100 * time.Millisecond)
	c.CleanUp()
	c.mu.RLock()
	if len(c.items) != 0 {
		t.Error("Expected items to be empty after CleanUp")
	}
	c.mu.RUnlock()
}

func TestMemoryCache_MaxSize(t *testing.T) {
	c := NewMemoryCache(WithMaxSize(2))
	c.Set("1", "one")
	c.Set("2", "two")
	c.Set("3", "three")

	c.mu.RLock()
	l := len(c.items)
	c.mu.RUnlock()

	if l > 2 {
		t.Errorf("Expected max size 2, got %d", l)
	}

	// Check that we can update existing key without exceeding size
	c.Set("3", "three-new")
	c.mu.RLock()
	l2 := len(c.items)
	c.mu.RUnlock()

	if l2 > 2 {
		t.Errorf("Expected max size 2 after update, got %d", l2)
	}
}

func TestMemoryCache_Concurrency(t *testing.T) {
	c := NewMemoryCache()
	var wg sync.WaitGroup
	workers := 100

	// Concurrent Set
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key%d", id), id)
		}(i)
	}
	wg.Wait()

	// Concurrent Get and Set
	wg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			c.Get(fmt.Sprintf("key%d", id))
		}(i)
		go func(id int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key%d", id), id*10)
		}(i)
	}
	wg.Wait()

	// Concurrent Delete
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			c.Delete(fmt.Sprintf("key%d", id))
		}(i)
	}
	wg.Wait()

	// Test concurrent CleanUp and Clear just to ensure no deadlocks or panics
	wg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key_extra_%d", id), id)
		}(i)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				c.CleanUp()
			} else {
				c.Clear()
			}
		}(i)
	}
	wg.Wait()
}
