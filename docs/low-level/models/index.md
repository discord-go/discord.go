# Resource Models

## Overview

The model packages are transport-independent Go representations of Discord
objects. They expose exported fields with JSON tags and a small number of
custom marshalers where Discord's wire format needs special handling. Models
do not fetch data, enforce permissions, or maintain freshness; REST and
Gateway code provide those boundaries.

## Architecture

[`application/`](../application/README.md) contains applications, teams,
installation parameters, and activity instances. [`channels/`](../channels/README.md)
contains channel variants, threads, invites, forum tags, and overwrites.
[`guilds/`](../guilds/README.md) covers guild snapshots, features, AutoMod,
onboarding, integrations, events, stages, templates, widgets, and welcome
screens. [`users/`](../users/README.md) covers users, members, presences,
flags, decorations, collectibles, and primary guild data. [`roles/`](../roles/README.md)
contains role permissions and presence-based tags. [`emojis/`](../emojis/README.md)
contains emojis and sticker families. [`webhook/`](../webhook/README.md)
contains full webhook objects. [`snowflake/`](../snowflake/README.md) provides
IDs used across all of them.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

func main() {
	guild := guilds.Guild{ID: snowflake.ID(42), Name: "Example"}
	user := users.User{ID: snowflake.ID(7), Username: "reader"}
	data, err := json.Marshal(struct {
		Guild guilds.Guild `json:"guild"`
		User users.User `json:"user"`
	}{guild, user})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
```

## Creating Models

Use struct literals for tests and decode API data into pointers. Nullable IDs,
timestamps, and optional nested objects are pointers in the relevant models;
preserve that distinction when sending partial updates. Snowflakes serialize as
strings through their field tags, and `snowflake.IDs` serializes arrays as
string arrays.

## Using Models

Decode a Gateway envelope and select a model from its event name, or use REST
methods that return a model directly. Models may embed other packages, such as
`events.MessageCreate` embedding `messages.Message` or `auditlog.AuditLog`
containing guild and channel values. Custom unmarshalers handle component
interfaces, nullable snowflakes, role ID arrays, and presence-based role tags.

## Common Patterns

Treat every response as a snapshot. Merge partial message updates explicitly,
and use REST reads when a cache entry is absent or stale. Keep model packages
in domain code and keep transport setup in [`../rest/`](../rest/README.md) or
[`../gateway/`](../gateway/README.md).

## Best Practices

Check unmarshal errors, especially when IDs are supplied by external clients.
Use pointer checks before reading optional data. Preserve unknown strings in
feature and enum-like fields so forward-compatible API values are not lost.

## Common Mistakes

A model constructor does not exist for most packages because zero values are
valid partial objects. A decoded model is not automatically cached. Do not
assume all fields are present on every endpoint or event variant.

## API Walkthrough

The package directories are the API split: applications, channels, guilds,
users, roles, emojis, webhook, and snowflake. Their individual README pages
list each exported type, method, constant, constructor, and unmarshaller.

## Examples

The Quick Start program is complete and runnable. Follow the linked package
guides for focused examples and exact field behavior.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../events/`](../events/README.md)
- [`../cache/`](../cache/README.md)
