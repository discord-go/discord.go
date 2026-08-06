package storage

import (
	"context"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore(0)
	if err := store.Set(context.Background(), "guild:1", map[string]string{"prefix": "?"}); err != nil {
		t.Fatal(err)
	}
	var value map[string]string
	if err := store.Get(context.Background(), "guild:1", &value); err != nil {
		t.Fatal(err)
	}
	if value["prefix"] != "?" {
		t.Fatalf("stored value = %#v", value)
	}
	keys, err := store.Keys(context.Background(), "guild:")
	if err != nil || len(keys) != 1 {
		t.Fatalf("keys = %#v, %v", keys, err)
	}
}
