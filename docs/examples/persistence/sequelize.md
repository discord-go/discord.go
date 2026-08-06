# Sequelize-Model Persistence

## Overview

Sequelize is a JavaScript ORM. It maps model operations such as `findOne`,
`upsert`, `destroy`, and `findAll` to a relational database. It is not a Go
dependency and cannot be imported by a Go bot. To adapt a Sequelize-backed
service to this repository, put a small repository boundary in front of the
model and make that boundary implement `storage.Store`.

This guide uses a single JSON key-value table because that is the shape of the
current interface. It is useful for settings, feature flags, and small
application records. It is not a replacement for a relational domain model:
use typed repositories and transactions when a feature needs joins, unique
constraints, counters, or multi-row invariants.

The adapter maps:

- `findOne({ where: { key } })` to `Store.Get`;
- `upsert({ key, value })` to `Store.Set`;
- `destroy({ where: { key } })` to `Store.Delete`;
- an ordered `findAll` key query to `Store.Keys`.

The value column should contain JSON text or a native JSON column. The key
column must be unique. Store Discord snowflakes as strings at the application
boundary so they remain exact across JavaScript number limits and database
drivers.

## Architecture

```text
bot handler
    |
    v
storage.Store
    |
    v
SequelizeStore: JSON encode/decode + missing-row mapping
    |
    v
SequelizeModel boundary
    |
    v
Node service with Sequelize, or a Go SQL/ORM implementation of the same model
```

A representative table is:

| column | purpose |
| --- | --- |
| `key` | Primary key, for example `guild:123:settings`. |
| `value` | JSON or text containing the application record. |
| `updated_at` | Database-managed update timestamp for operations and audits. |

The `storage.Store` interface intentionally does not expose transactions,
partial updates, or database-specific query expressions. Keep those features
behind a domain repository. The adapter should make the simple contract
correct rather than leak ORM objects into interaction handlers.

## Prerequisites

- Go `1.26.4` or a compatible newer toolchain, as declared by
  [`go.mod`](../../../go.mod).
- This repository, for `discord.go/storage` and `discord.go/bot`.
- A Discord bot token and an application with the `Guilds` intent if running
  the bot portion.
- A relational database and a migration for the key-value table in production.
- If Sequelize remains in a Node service, an authenticated RPC or HTTP
  transport that implements the model boundary used below.

The runnable program uses an in-memory model so it compiles with the current
module's dependencies. It demonstrates the adapter and bot integration without
inventing a database driver that is not present in `go.mod`.

## Quick Start

The complete program below defines the model boundary, maps it to
`storage.Store`, attaches the result with `bot.WithStore`, and serves a
`/storage-check` slash command. `memoryModel` stands in for calls to a
Sequelize model. A production implementation replaces only `memoryModel`.

Run it from the repository root after placing the program in a Go package:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run sequelize-example.go
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

// SequelizeModel describes the operations a Sequelize model or a Go wrapper
// around that model must provide. RawMessage is the database JSON column.
type SequelizeModel interface {
	FindOne(context.Context, string) (json.RawMessage, error)
	Upsert(context.Context, string, json.RawMessage) error
	Destroy(context.Context, string) (bool, error)
	FindKeys(context.Context, string) ([]string, error)
}

type SequelizeStore struct {
	model SequelizeModel
}

var _ storage.Store = (*SequelizeStore)(nil)

func NewSequelizeStore(model SequelizeModel) *SequelizeStore {
	return &SequelizeStore{model: model}
}

func (s *SequelizeStore) Get(ctx context.Context, key string, target any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	data, err := s.model.FindOne(ctx, key)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return storage.ErrNotFound
	}
	return json.Unmarshal(data, target)
}

func (s *SequelizeStore) Set(ctx context.Context, key string, value any) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.model.Upsert(ctx, key, data)
}

func (s *SequelizeStore) Delete(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	_, err := s.model.Destroy(ctx, key)
	return err
}

func (s *SequelizeStore) Keys(ctx context.Context, prefix string) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	keys, err := s.model.FindKeys(ctx, prefix)
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)
	return keys, nil
}

// memoryModel has the same observable behavior as a model backed by a unique
// key column and a JSON value column.
type memoryModel struct {
	mu     sync.RWMutex
	values map[string]json.RawMessage
}

func newMemoryModel() *memoryModel {
	return &memoryModel{values: make(map[string]json.RawMessage)}
}

func (m *memoryModel) FindOne(ctx context.Context, key string) (json.RawMessage, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	value, ok := m.values[key]
	m.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return append(json.RawMessage(nil), value...), nil
}

func (m *memoryModel) Upsert(ctx context.Context, key string, value json.RawMessage) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	m.values[key] = append(json.RawMessage(nil), value...)
	m.mu.Unlock()
	return nil
}

func (m *memoryModel) Destroy(ctx context.Context, key string) (bool, error) {
	if err := contextErr(ctx); err != nil {
		return false, err
	}
	m.mu.Lock()
	_, existed := m.values[key]
	delete(m.values, key)
	m.mu.Unlock()
	return existed, nil
}

func (m *memoryModel) FindKeys(ctx context.Context, prefix string) ([]string, error) {
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

	store := NewSequelizeStore(newMemoryModel())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := store.Set(ctx, "guild:demo", guildSettings{Prefix: "?"}); err != nil {
		log.Fatal(err)
	}

	router := bot.NewRouter()
	router.Command("storage-check", "Read a relational record through storage.Store", func(ctx *bot.InteractionContext) {
		var settings guildSettings
		if err := store.Get(ctx.Context(), "guild:demo", &settings); err != nil {
			log.Printf("storage: %v", err)
			_ = ctx.ReplyEphemeral("storage read failed")
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

### Basic: one model and one JSON column

Create a table with a unique string key and a JSON or text value. Keep the
adapter responsible for serialization and make the model responsible for
queries. This lets the rest of the bot use the stable `storage.Store` API.

The equivalent Sequelize operations are conceptually:

```js
const row = await Settings.findOne({ where: { key } });
await Settings.upsert({ key, value: JSON.stringify(value) });
await Settings.destroy({ where: { key } });
await Settings.findAll({
  attributes: ['key'],
  where: { key: { [Op.like]: `${prefix}%` } },
  order: [['key', 'ASC']],
});
```

Treat this JavaScript fragment as a model mapping, not as Go code. The Go
adapter's `SequelizeModel` boundary is where an RPC client or a Go SQL layer
implements these operations.

### Intermediate: migrations and versioned records

Use migrations to create the table and indexes. Do not use a destructive
`sync({ force: true })` equivalent in a deployed bot. Add a `version` field to
values when the JSON shape can evolve, and migrate records on read or with an
explicit job.

Keep database errors distinguishable. A unique-key violation should be handled
as a conflict in a domain repository, while a connection failure should be
returned as an operational error. `Store.Set` is an upsert boundary and does
not promise compare-and-swap semantics.

### Advanced: relational domain repositories

Use a typed model for tags, economy balances, or audit records when you need
constraints and queries. That model can coexist with `SequelizeStore`:

- `Store` holds small opaque settings and feature flags;
- a typed repository owns transactions and invariants;
- handlers depend on application services, not on Sequelize instances;
- migrations and indexes are deployed independently of bot startup.

For a multi-process bot, use database transactions or atomic SQL updates for
counters. A `Get` followed by a `Set` can lose a concurrent update.

## Best Practices

- Make `key` a primary key or add a unique index; `upsert` without uniqueness
  is not a safe replacement for an update.
- Store snowflakes as strings in JSON and database columns.
- Use parameterized ORM conditions; never concatenate user input into SQL or
  Sequelize literals.
- Add an index that supports the `Keys` prefix query, or replace `Keys` with a
  domain-specific listing method when scans are too broad.
- Keep migrations separate from application startup and never erase production
  data as a startup convenience.
- Pass bounded contexts to model methods and configure database pool limits.
- Preserve `storage.ErrNotFound` for an absent row and keep `Delete`
  idempotent, as `MemoryStore` does.
- Use transactions for multi-row state and atomic updates for counters.
- Redact database URLs, credentials, OAuth tokens, and user content from logs.
- Test the adapter with a real database in addition to the in-memory model;
  SQL null, JSON, collation, and transaction behavior vary by provider.

## Common Mistakes

### Running schema synchronization destructively

Wrong for a deployed service:

```js
await Settings.sync({ force: true });
```

Use migrations and an explicit test database instead. A bot restart must not
drop settings.

### Treating `Store.Set` as an atomic merge

Wrong:

```go
var value Counter
_ = store.Get(ctx, key, &value)
value.Count++
_ = store.Set(ctx, key, value)
```

Two handlers can read the same count and overwrite one another. Use an atomic
database update or a domain repository transaction for counters.

### Returning a model object from the adapter

`storage.Store.Get` promises JSON decoding into the caller's target. Returning a
Sequelize instance, driver row, or `json.RawMessage` from every handler leaks
the persistence choice and makes a backend migration expensive.

### Using a broad prefix scan for user input

`Keys(ctx, "")` can return every record. Restrict prefixes in application code,
paginate at the model boundary if needed, and reserve administrative scans for
trusted maintenance paths.

## API Walkthrough

- [`storage.Store`](../../../storage/store.go) defines the four methods the
  adapter must implement.
- [`storage.ErrNotFound`](../../../storage/store.go) is the canonical missing
  row result.
- [`storage.MemoryStore`](../../../storage/store.go) is a small reference
  implementation and useful test double.
- [`bot.WithStore`](../../../bot/bot.go) exposes the adapter to `bot.Bot`.
- [`bot.New`](../../../bot/bot.go) and [`bot.NewRouter`](../../../bot/router.go)
  are the current bot and command constructors used by the example.

The `SequelizeModel` boundary in the Quick Start is intentionally narrower than
Sequelize. It captures only what `storage.Store` needs. Add typed methods to a
separate repository rather than expanding this adapter with ORM-specific
filters, associations, or transactions.

## Examples

A typed tag service can use the same lifecycle and interaction rules while
keeping its model separate:

```go
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
}

func saveTag(ctx context.Context, store storage.Store, tag Tag) error {
	if tag.Name == "" || len(tag.Name) > 64 {
		return fmt.Errorf("invalid tag name")
	}
	return store.Set(ctx, "tag:"+tag.Name, tag)
}
```

For a real Sequelize model, map `null` rows to a nil `json.RawMessage` so the
adapter returns `storage.ErrNotFound`. Map a successful `destroy` with zero
affected rows to `nil`, matching `storage.MemoryStore`, unless a domain method
specifically needs to report whether a row existed.

## Related Links

- [`storage.Store`](../../../storage/store.go)
- [`storage` tests](../../../storage/store_test.go)
- [`bot.WithStore`](../../../bot/bot.go)
- [Keyv-compatible guide](keyv.md)
- [Examples overview](../README.md)
- [Sequelize documentation](https://sequelize.org/docs/v6/)
- [Sequelize model querying](https://sequelize.org/docs/v6/core-concepts/model-querying-basics/)
- [Sequelize migrations](https://sequelize.org/docs/v6/other-topics/migrations/)
- [Discord.js Sequelize guide that inspired this adapter](https://discordjs.dev/docs/packages/guide/legacy/sequelize)
