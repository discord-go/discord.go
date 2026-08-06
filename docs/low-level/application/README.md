# Applications

## Overview

The `application` package models application metadata returned by Discord. It
contains `Application`, `Team`, `TeamMember`, installation parameters,
integration configuration, and activity-instance location data. It does not
register commands or manage entitlements; those operations are grouped in
[`../rest/`](../rest/README.md).

## Architecture

`Application` is a broad response model. It includes identity, description,
icons, bot and owner users, verification key, team, redirect URIs, install
parameters, integration types, role-connection URLs, tags, approximate counts,
and event-webhook fields. Most optional response fields use pointers or
`omitempty` so a partial response can be represented.

`Team` owns `TeamMember` values. A member has `MembershipState` (invited = 1,
accepted = 2), permissions, team ID, user, and role. `InstallParams` carries
OAuth scopes and a permission string. `IntegrationTypeConfig` can contain
OAuth2 install parameters. `ActivityInstance` groups an instance ID,
application ID, free-form activity data, participants, and locations.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/application"
)

func main() {
	var app application.Application
	if err := json.Unmarshal([]byte(`{"id":"42","name":"Example","description":"demo","install_params":{"scopes":["bot"],"permissions":"0"}}`), &app); err != nil {
		panic(err)
	}
	fmt.Println(app.ID, app.Name, app.InstallParams.Scopes[0], app.InstallParams.Permissions)
}
```

## Creating And Using Applications

All fields are exported and can be populated with struct literals. Use
`InstallParams` when displaying or validating an install URL, and inspect
`IntegrationTypesConfig` when an application supports different install
paths. `TeamMember.User` is a full `users.User` model. `ActivityLocation` uses
strings for guild and channel IDs because that is the repository's model for
this response shape.

## Common Patterns

Use `Application.VerifyKey` with interaction signature verification only after
converting the stored key to the format expected by
[`../interactions/`](../interactions/README.md). Use `Application.Team` only
when the response includes team data; it is nil for many applications. Use
`ActivityInstance.Participants` as a snapshot, not a live membership list.

## Best Practices

Do not expose `VerifyKey` or OAuth client secrets unnecessarily. Preserve
unknown map data in `ActivityInstance.Activity` when forwarding responses.
Check optional pointers before use, especially `Bot`, `Owner`, `Team`, and
`InstallParams`.

## Common Mistakes

The package has no `NewApplication` constructor and no API calls. It does not
turn a permission string into a `permissions.Permission` automatically. Team
membership state is an integer enum, not a boolean.

## API Walkthrough

The complete exported API is `Application`, `Team`, `TeamMember`,
`MembershipState` and its two constants, `InstallParams`,
`IntegrationTypeConfig`, `ActivityInstance`, and `ActivityLocation`.

## Examples

The Quick Start program is runnable without a token. Continue with
[`../rest/endpoints.md`](../rest/endpoints.md) for current-application and
command methods, or [`../oauth2/`](../oauth2/README.md) for bearer access.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../oauth2/`](../oauth2/README.md)
- [`../interactions/`](../interactions/README.md)
