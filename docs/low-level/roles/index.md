# Roles

## Overview

The `roles` package models the role object returned by Discord. `Role` combines
identity, display properties, hierarchy, permissions, managed state, tags, and
flags. It is a data package: role creation, editing, positioning, and deletion
are methods on [`../rest/`](../rest/README.md).

## Architecture

`Role.ID` is a `snowflake.ID`; `Permissions` is a
[`permissions.Permission`](../permissions/README.md) with a string JSON tag.
`Color` is the legacy integer color, while `Colors` can contain primary,
secondary, and tertiary colors. `Icon` and `UnicodeEmoji` are optional display
values. `Position` controls hierarchy, but the API returns the authoritative
order.

`RoleTags` has normal optional ID fields for bot, integration, and subscription
listing ownership. `PremiumSubscriber`, `AvailableForPurchase`, and
`GuildConnections` are presence flags: their JSON values are normally null,
and the custom `UnmarshalJSON` sets the bool when the key is present. They are
not serialized by the struct's own JSON tags.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/roles"
)

func main() {
	var role roles.Role
	err := json.Unmarshal([]byte(`{"id":"10","name":"moderator","permissions":"1024","tags":{"premium_subscriber":null}}`), &role)
	if err != nil {
		panic(err)
	}
	fmt.Println(role.ID, role.Name, role.Permissions, role.Tags.PremiumSubscriber)
}
```

## Creating And Using Roles

`Role` fields are exported for direct construction. `RoleColors` contains the
three integer color values. `RoleTags` can be attached with a pointer when
known. `Managed` identifies integration-managed roles and `Mentionable`
controls whether users can mention the role; neither field grants authority.
Use `Position` only as received or as input to a REST role-position request.

## Common Patterns

Check permissions with `Permission.Has` and `HasAll`, not by comparing the
whole integer. A role can have both a legacy `Color` and newer `Colors`. When
displaying a role icon, pass `Role.Icon` to the CDN URL builder only when it is
non-nil; role IDs and icon hashes are independent values.

## Best Practices

Treat role objects as snapshots. Do not infer a user's effective permissions
from one role; combine all member roles and channel overwrites with
`permissions.Calculate`. Preserve unknown tag fields when forwarding data by
using the model as a decode target and avoiding lossy map transformations.

## Common Mistakes

Presence-based tag fields are not ordinary booleans in the wire format. A
missing `premium_subscriber` key and a present `null` key are represented
differently by the custom unmarshaller. `Role.Permissions` is a bitfield, not
a decimal permission count, and `Managed` roles generally cannot be edited by
normal role endpoints.

## API Walkthrough

The exported declarations are `Role`, `RoleColors`, `RoleTags`, and
`RoleTags.UnmarshalJSON`. `Role` exposes ID, Name, Color, Colors, Hoist, Icon,
UnicodeEmoji, Position, Permissions, Managed, Mentionable, Tags, and Flags.
There are no constructors or role mutation methods in this package.

## Examples

The Quick Start program is complete and runnable. For writes, use the typed
parameter structs and methods in [`../rest/endpoints.md`](../rest/endpoints.md).

## Related APIs

- [`../permissions/`](../permissions/README.md)
- [`../snowflake/`](../snowflake/README.md)
- [`../rest/`](../rest/README.md)
