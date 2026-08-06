# Keyv-Compatible Persistence

## Overview

Keyv is a JavaScript key-value abstraction. It exposes asynchronous `get`,
`set`, `delete`, and `clear` operations and delegates persistence to a driver
such as Redis, SQLite, PostgreSQL, or MongoDB. It is not a Go package, so a Go
bot cannot import Keyv directly. The equivalent integration point in this
repository is `storage.Store`.

This guide adapts the Keyv shape into that interface:

- values are JSON-marshaled by the adapter, just as Keyv supports JSON values;
- missing values become `storage.ErrNotFound` instead of JavaScript
  `undefined`;
- every backend operation receives a `context.Context`;
- namespaces are represented by a key prefix;
- `Keys` is implemented by a driver-level prefix scan, not by pretending that
  Keyv's `clear` method is a key iterator.

The result is a backend-independent store that can be attached with
`bot.WithStore` and used for guild settings, feature flags, moderation state,
or other application-owned data. It should not be used as the Discord gateway
cache; `bot.WithCache` and `gateway.ShardManager.SetCache` serve that separate
purpose.

## Architecture

```text
command or service
        |
        v
storage.Store
        |
        v
KeyvStore: namespace + JSON encode/decode + error mapping
        |
        v
Keyv-like driver: get / set / delete / prefix keys
        |
        v
Redis, SQL, document store, or a service that owns a Keyv instance
```

The adapter contract is deliberately small. A real driver can be backed by a
Go Redis or SQL client, or it can call a separate Node service that owns Keyv.
The application code does not change in either case.

| `storage.Store` method | Keyv-style operation | Adapter behavior |
| --- | --- | --- |
| `Get(ctx, key, &target)` | `get(key)` | Decode JSON into `target`; map an absent key to `storage.ErrNotFound`. |
| `Set(ctx, key, value)` | `set(key, value)` | Encode `value` as JSON before writing. |
| `Delete(ctx, key)` | `delete(key)` | Make deletion idempotent, matching `MemoryStore`. |
| `Keys(ctx, prefix)` | driver scan or index | Return sorted, namespace-relative keys. |

Keyv's `clear` operation has no direct `Store` equivalent. If an application
needs to clear one namespace, enumerate `Keys(ctx, "")` and delete the returned
keys, or add a backend-specific administrative method outside the `Store`
interface. Do not implement a broad clear operation behind an ordinary user
command.

## Prerequisites

- Go `1.26.4` or a compatible newer toolchain, as declared by
  [`go.mod`](../../../go.mod).
- A checkout of this module, because the example imports `discord.go/storage`
  and `discord.go/bot`.
- A Discord application and bot token if the bot portion of the example is
  run. Keep `DISCORD_TOKEN` out of source control.
- The `Guilds` intent enabled for the slash-command example.
- A real Keyv driver or a Go implementation of the `KeyvBackend` contract for
  production. The runnable example uses a concurrency-safe in-memory driver
  so it has no additional dependency.

Keyv itself is appropriate when the storage owner is a Node process. If the
Go bot and Node service are separate processes, define the backend contract at
that service boundary and preserve the same not-found, JSON, timeout, and
namespace rules.

## Quick Start

The following complete program adapts a Keyv-like backend, attaches it to
`bot.Bot`, stores a guild setting, and exposes `/storage-check`. The included
backend is only a runnable stand-in for a Keyv driver; replace
`memoryKeyv` with a real implementation without changing `KeyvStore` or the
bot handler.

Run it from the repository root after placing the program in a Go package:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run keyv-example.go
```

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/storage"
)

// KeyvBackend is the small driver boundary needed by this adapter. A real
// implementation can wrap Redis, SQL, or an RPC client for a Node Keyv service.
type KeyvBackend interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte) error
	Delete(context.Context, string) error
	Keys(context.Context, string) ([]string, error)
}

type KeyvStore struct {
	backend  KeyvBackend
	namespace string
}

var _ storage.Store = (*KeyvStore)(nil)

func NewKeyvStore(backend KeyvBackend, namespace string) *KeyvStore {
	return &KeyvStore{backend: backend, namespace: strings.Trim(namespace, ":")}
}

func (s *KeyvStore) scoped(key string) string {
	if s.namespace == "" {
		return key
	}
	return s.namespace + ":" + key
}

func (s *KeyvStore) unscoped(key string) string {
	if s.namespace == "" {
		return key
	}
	return strings.TrimPrefix(key, s.namespace+":")
}

func (s *KeyvStore) Get(ctx context.Context, key string, target any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	data, err := s.backend.Get(ctx, s.scoped(key))
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return storage.ErrNotFound
	}
	return json.Unmarshal(data, target)
}

func (s *KeyvStore) Set(ctx context.Context, key string, value any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.backend.Set(ctx, s.scoped(key), data)
}

func (s *KeyvStore) Delete(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	return s.backend.Delete(ctx, s.scoped(key))
}

func (s *KeyvStore) Keys(ctx context.Context, prefix string) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	keys, err := s.backend.Keys(ctx, s.scoped(prefix))
	if err != nil {
		return nil, err
	}
	for i := range keys {
		keys[i] = s.unscoped(keys[i])
	}
	sort.Strings(keys)
	return keys, nil
}

type memoryKeyv struct {
	mu     sync.RWMutex
	values map[string][]byte
}

func newMemoryKeyv() *memoryKeyv {
	return &memoryKeyv{values: make(map[string][]byte)}
}

func (m *memoryKeyv) Get(ctx context.Context, key string) ([]byte, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	value, ok := m.values[key]
	m.mu.RUnlock()
	if !ok {
		return nil, storage.ErrNotFound
	}
	return append([]byte(nil), value...), nil
}

func (m *memoryKeyv) Set(ctx context.Context, key string, value []byte) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	m.values[key] = append([]byte(nil), value...)
	m.mu.Unlock()
	return nil
}

func (m *memoryKeyv) Delete(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.values, key)
	m.mu.Unlock()
	return nil
}

func (m *memoryKeyv) Keys(ctx context.Context, prefix string) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	keys := make([]string, 0)
	for key := range m.values {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	m.mu.RUnlock()
	sort.Strings(keys)
	return keys, nil
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

type guildSettings struct {
	Prefix string `json:"prefix"`
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	store := NewKeyvStore(newMemoryKeyv(), "discord")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := store.Set(ctx, "guild:demo", guildSettings{Prefix: "?"}); err != nil {
		log.Fatal(err)
	}

	router := bot.NewRouter()
	router.Command("storage-check", "Read a setting through storage.Store", func(ctx *bot.InteractionContext) {
		var settings guildSettings
		if err := store.Get(ctx.Context(), "guild:demo", &settings); err != nil {
			if err := ctx.ReplyEphemeral("storage read failed"); err != nil {
				log.Printf("reply: %v", err)
			}
			log.Printf("storage: %v", err)
			return
		}
		if err := ctx.Reply(fmt.Sprintf("Stored prefix: %s", settings.Prefix)); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	client := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
		bot.WithStore(store),
	)
	client.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Practical Levels

### Basic: one namespace and JSON settings

Use one `Store` instance and stable keys such as `guild:<guild-id>`. Keep the
value a versioned struct rather than a map when the shape is part of your
application contract.

```go
type Settings struct {
	Version int    `json:"version"`
	Prefix  string `json:"prefix"`
}

var settings Settings
err := store.Get(ctx, "guild:"+guildID, &settings)
if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
}
```

### Intermediate: TTL and cache policy

Keyv drivers commonly offer TTL. `storage.Store` has no TTL parameter, so
choose one of these designs explicitly:

- configure a TTL on the backend for data that is safe to recompute;
- store `expires_at` in the JSON value and reject expired records in the
  application;
- expose a separate typed cache for ephemeral data and keep `Store` durable.

Do not silently add a TTL to settings or moderation records. A namespace also
does not provide a transaction or a distributed lock.

### Advanced: shared storage across shards and replicas

Use a shared backend when more than one process needs the same state. Include a
schema or version field, make writes idempotent, and avoid read-modify-write
updates unless the backend provides an atomic operation. Use a separate key
prefix for each data family, for example `guild:`, `user:`, and `oauth:`.

If the driver has no reliable `Keys` operation, maintain an explicit index or
use a domain-specific repository instead of implementing an expensive full
scan.

## Best Practices

- Keep `KeyvStore` behind `storage.Store`; handlers should not know the driver.
- Pass the interaction or request context into storage calls so cancellation
  reaches the backend.
- Use `errors.Is(err, storage.ErrNotFound)` for misses; do not compare error
  strings.
- Namespace keys and validate IDs before concatenating them.
- Marshal only JSON-safe values and use pointers when `Get` must populate a
  target.
- Make `Delete` idempotent, matching `storage.MemoryStore` behavior.
- Sort keys before returning them so callers receive deterministic results.
- Treat backend errors as operational failures and report them without logging
  user data or credentials.
- Add metrics for latency, errors, hit/miss counts, and key families.
- Test TTL, serialization failures, cancellation, driver outages, and two
  concurrent writers before selecting a production backend.

## Common Mistakes

### Treating a missing value as a zero value

Wrong:

```go
var settings Settings
_ = store.Get(ctx, key, &settings)
use(settings) // may use defaults after silently losing a storage error
```

Correct:

```go
var settings Settings
if err := store.Get(ctx, key, &settings); err != nil {
	if errors.Is(err, storage.ErrNotFound) {
		settings = Settings{Version: 1, Prefix: "!"}
	} else {
		return err
	}
}
use(settings)
```

### Assuming every Keyv provider can list keys

`get`, `set`, and `delete` are the common Keyv operations. `storage.Store.Keys`
requires a prefix scan, which may be unavailable or expensive in a driver.
Implement it with a bounded backend query or maintain an index. Never expose a
prefix supplied by an untrusted user as an unrestricted administrative scan.

### Sharing a process-local backend between shards

An in-memory Keyv driver is local to one process. It does not synchronize two
bot replicas and is lost at restart. Use a shared durable backend for state
that must survive either event.

### Confusing namespaces with authorization

`discord:guild:123` prevents accidental key collisions; it does not decide
whether the caller may edit guild `123`. Apply `bot.GuildOnly` and permission
middleware before changing settings.

## API Walkthrough

- [`storage.Store`](../../../storage/store.go) defines `Get`, `Set`, `Delete`,
  and `Keys`, all with a context.
- [`storage.ErrNotFound`](../../../storage/store.go) is the canonical missing
  record error.
- [`storage.NewMemoryStore`](../../../storage/store.go) is the reference
  behavior for JSON encoding, idempotent deletion, sorted keys, and optional
  process-local TTL.
- [`bot.WithStore`](../../../bot/bot.go) attaches a store to `bot.Bot`.
- [`bot.New`](../../../bot/bot.go) creates the bot used by the Quick Start.
- [`bot.NewRouter`](../../../bot/router.go) and `Router.Command` register the
  example command.

The adapter's `Get` must unmarshal into the caller's target, not return a
`json.RawMessage` and force every handler to know the backend format. The
adapter's `Set` must marshal once at the boundary so all backends share the
same JSON semantics.

## Examples

A guild settings repository can centralize key construction and defaults:

```go
func guildKey(id snowflake.ID) string { return "guild:" + id.String() }

func loadSettings(ctx context.Context, store storage.Store, id snowflake.ID) (Settings, error) {
	var value Settings
	if err := store.Get(ctx, guildKey(id), &value); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Settings{Version: 1, Prefix: "!"}, nil
		}
		return Settings{}, err
	}
	return value, nil
}
```

For a Node-owned Keyv service, implement `KeyvBackend` with an authenticated
RPC client. Keep the service response distinction explicit: a `404` maps to
`storage.ErrNotFound`, invalid JSON maps to a decode error, and transport
timeouts return the original context or network error.

## Related Links

- [`storage.Store`](../../../storage/store.go)
- [`storage` tests](../../../storage/store_test.go)
- [`bot.WithStore`](../../../bot/bot.go)
- [Examples overview](../README.md)
- [Keyv documentation](https://keyv.org/docs/)
- [Keyv repository](https://github.com/jaredwray/keyv)
- [Discord.js Keyv guide that inspired this adapter](https://discordjs.dev/docs/packages/guide/legacy/keyv/keyv)
