# Permissions

## Overview

`permissions.Permission` is a `uint64` bitfield. Each exported permission
constant occupies one bit, so a role or member permission set is represented
compactly and serializes as a Discord string when used in the REST parameter
types that carry `json:",string"`. The package also resolves channel
overwrites using Discord's precedence rules.

## Architecture

`Build(permissions ...Permission)` ORs flags into one value. `NewBuilder`,
`Builder.Add`, `Builder.Remove`, and `Builder.Build` provide the same operation
with fluent mutation. On a `Permission` value, `Has` tests whether any supplied
bit is present and `HasAll` tests whether every supplied bit is present.
Pointer methods `Add` and `Remove` mutate the receiver.

`Calculate` first grants all permissions to the guild owner, then combines the
base @everyone permissions with member role permissions. Administrator also
grants all bits. It applies the @everyone overwrite, aggregates matching role
overwrites, and finally applies the member-specific overwrite. `Overwrite.Type`
is 0 for roles and 1 for members.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/permissions"
)

func main() {
	p := permissions.NewBuilder(permissions.ViewChannel).
		Add(permissions.SendMessages, permissions.ReadMessageHistory).
		Remove(permissions.SendMessages).
		Build()
	fmt.Println(p.Has(permissions.ViewChannel), p.HasAll(permissions.ViewChannel|permissions.ReadMessageHistory), p.Has(permissions.SendMessages))
}
```

## Creating A Permission Set

The constants cover invites, moderation, channel management, messages,
threads, voice, expressions, application commands, polls, soundboard, and
newer application permissions. `ManageGuildExpressions` is an alias of
`ManageEmojisAndStickers` in this repository. The zero value means no bits;
there is no validation against a particular Discord API version.

## Using Resolve

Call `Calculate(memberID, guildID, guildOwnerID, baseRolePermissions,
memberRoleIDs, memberRolePermissions, overwrites)`. The function ORs every
permission in `memberRolePermissions`; the caller is responsible for supplying
the permission values for the member's `memberRoleIDs` because the function
does not pair or validate those two slices. Missing overwrites simply do
nothing.
The owner and Administrator fast paths return `^Permission(0)`, including bits
that do not currently have a named constant.

## Common Patterns

Use `HasAll` for authorization requirements such as "view and send"; use
`Has` only for "at least one of these alternatives." Use pointers when passing
permissions to `rest.ModifyRoleParams` so omission is distinct from an
explicit zero.

## Best Practices

Calculate permissions from fresh role and overwrite data. Treat the result as
a snapshot because role edits and channel edits can invalidate it. Keep
permission checks server-side as well; a local bitfield is not a substitute
for Discord's response.

## Common Mistakes

`Has` does not mean all bits. `Remove` clears only supplied bits. Do not use
signed integers for permission values, and do not assume `Administrator`
needs channel overwrites applied: `Calculate` returns early for it.

## API Walkthrough

The exported API is `Permission`, all named permission constants, `Build`,
`NewBuilder`, `Builder.Add`, `Builder.Remove`, `Builder.Build`,
`Permission.Add`, `Permission.Remove`, `Permission.Has`, `Permission.HasAll`,
`Overwrite`, and `Calculate`.

## Examples

The Quick Start program is runnable and demonstrates both builder mutation and
the distinction between `Has` and `HasAll`.

## Related APIs

- [`../roles/`](../roles/README.md)
- [`../channels/`](../channels/README.md)
- [`../rest/`](../rest/README.md)
