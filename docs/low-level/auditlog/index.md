# Audit Logs

## Overview

The `auditlog` package models the object returned by Discord's guild audit-log
endpoint. It keeps the entry list together with referenced users, threads,
webhooks, integrations, application commands, AutoMod rules, and scheduled
events. It does not fetch logs; use `rest.GetAuditLog` and pass filters through
`rest.AuditLogParams`.

## Architecture

`AuditLog` is the top-level response. Each `AuditLogEntry` has an ID, optional
target ID, optional actor ID, an `AuditLogEvent` action type, optional
`OptionalAuditEntryInfo`, changes, and the audit reason. `AuditLogChange` stores
`NewValue` and `OldValue` as `interface{}` because the value shape depends on
the changed key. The API may omit changes entirely.

`OptionalAuditEntryInfo` contains endpoint-specific values such as channel,
message, role, application, deletion count, AutoMod rule name, and integration
type. Its fields are strings or snowflakes according to Discord's response;
do not assume every field is populated for every action.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/auditlog"
)

func main() {
	var entry auditlog.AuditLogEntry
	err := json.Unmarshal([]byte(`{"id":"900","user_id":"42","action_type":32,"target_id":"700","reason":"cleanup"}`), &entry)
	if err != nil {
		panic(err)
	}
	fmt.Println(entry.ID, entry.UserID, entry.ActionType == auditlog.ROLE_DELETE, entry.Reason)
}
```

## Creating And Using Entries

The exported action constants cover guild, channel, member, role, invite,
webhook, emoji, message, integration, stage, sticker, scheduled-event,
thread, command-permission, AutoMod, and creator-monetization actions. Use the
constant names rather than numeric literals when filtering application logic.
`TargetID` is a `*string` because Discord can return null and because some
target identifiers are not snowflakes. `UserID` is a nullable snowflake.

## Common Patterns

Decode `Changes` by switching on `change.Key` and then type-asserting or
re-marshaling `NewValue` into a typed value. Keep the raw values when building
an audit viewer because new Discord fields can appear without a library type
change. Use `rest.WithReason` for the write that produced a reason; the reason
is returned on the entry when Discord includes it.

## Best Practices

Treat action types as an open protocol even though the package defines current
constants. Handle nil `Options`, nil `UserID`, and nil `TargetID`. Audit logs
are snapshots and can be paginated or filtered, so do not infer that one page
contains the complete history.

## Common Mistakes

`AuditLogChange.NewValue` is not guaranteed to be a string or object. A nil
`Changes` slice is different from a change whose value is JSON null. Do not use
the target ID as a user ID without checking the action type.

## API Walkthrough

The exported API is `AuditLog`, `AuditLogEntry`, `AuditLogChange`,
`OptionalAuditEntryInfo`, and `AuditLogEvent` with all action constants from
`GUILD_UPDATE` through the creator-monetization actions. There are no
constructors or methods. Transport and reason helpers live in [`../rest/`](../rest/README.md).

## Examples

The Quick Start program is complete and runnable. A real fetch is a single
call such as `restClient.GetAuditLog(ctx, guildID, params)` after REST setup.

## Related APIs

- [`../rest/requests.md`](../rest/requests.md)
- [`../guilds/`](../guilds/README.md)
- [`../events/`](../events/README.md)
