# Storage

## Overview

`storage.Store` is the persistence boundary for application-owned state. It
is intentionally independent of Discord resources and can hold settings,
moderation records, statistics, tickets, or jobs. `MemoryStore` provides a
concurrency-safe JSON-backed implementation for tests and small processes;
production adapters can implement the same four methods over SQL, Redis, or a
document database.

## Architecture

The interface is `Get(ctx, key, target) error`, `Set(ctx, key, value) error`,
`Delete(ctx, key) error`, and `Keys(ctx, prefix) ([]string, error)`. Values are
serialized to JSON on `Set` and decoded from JSON on `Get`, which means the
store keeps a snapshot rather than a pointer to mutable caller memory.
`ErrNotFound` is returned for missing or expired records.

`NewMemoryStore(ttl time.Duration)` initializes an empty store. A positive TTL
is applied to every record at write time. Zero or negative TTL means no
expiration. `Keys` returns live keys whose names have the requested prefix in
sorted order; it does not delete expired records as a side effect.

## Quick Start

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/discord-go/discord.go/storage"
)

type Settings struct {
	Prefix string `json:"prefix"`
	Enabled bool `json:"enabled"`
}

func main() {
	ctx := context.Background()
	store := storage.NewMemoryStore(0)
	if err := store.Set(ctx, "guild:42", Settings{Prefix: "!", Enabled: true}); err != nil {
		panic(err)
	}
	var settings Settings
	if err := store.Get(ctx, "guild:42", &settings); err != nil {
		panic(err)
	}
	keys, _ := store.Keys(ctx, "guild:")
	fmt.Println(settings.Prefix, settings.Enabled, keys)
	if err := store.Get(ctx, "missing", &settings); !errors.Is(err, storage.ErrNotFound) {
		panic("expected a missing-record error")
	}
}
```

## Creating A Store

The constructor is the only constructor in the package. `MemoryStore.TTL` is
exported so it can be inspected, but changing it after records are written
does not recalculate existing expiration times. A zero `MemoryStore` value is
not initialized with a records map; use `NewMemoryStore`.

## Using A Store

All methods check a non-nil context for cancellation before touching the store.
They accept a nil context for compatibility, but callers should pass a real
context. `Set` returns JSON marshal errors, `Get` returns JSON unmarshal errors,
and `Delete` succeeds even when the key is absent. Keys are opaque strings;
prefix matching is a simple `strings.HasPrefix` test.

## Common Patterns

Namespace keys such as `guild:<id>:settings` and `user:<id>:preferences`.
Store versioned structs so future migrations can distinguish old records. Use
`errors.Is(err, storage.ErrNotFound)` rather than comparing error text.

## Best Practices

Keep records small, avoid storing secrets without encryption, and treat writes
as replacements. Use a durable implementation before relying on the store for
moderation or financial decisions. Honor context cancellation in custom Store
implementations just as `MemoryStore` does.

## Common Mistakes

`MemoryStore` is not durable or distributed. `Keys` does not return values and
does not remove expired records. Passing a non-pointer target to `Get` causes a
JSON decoding error, and a nil target is not a useful destination.

## API Walkthrough

The complete exported API is `ErrNotFound`, `Store`, `MemoryStore`,
`NewMemoryStore`, and `MemoryStore.Get`, `Set`, `Delete`, and `Keys`.

## Examples

The Quick Start program is runnable and demonstrates JSON snapshot semantics,
prefix listing, and the package error value.

## Related APIs

- [`../cache/`](../cache/README.md)
- [`../json/`](../json/README.md)
- [`../models/`](../models/README.md)
