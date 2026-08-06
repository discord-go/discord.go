# Guilds

## Overview

The `guilds` package is the model layer for Discord servers. It covers the
full `Guild` object plus previews, unavailable notices, feature flags, roles
and emojis embedded in guild responses, integrations, onboarding, welcome
screens, scheduled events, AutoMod rules, stage instances, templates, vanity
URLs, and widgets. Fetching and mutation are provided by
[`../rest/`](../rest/README.md).

## Architecture

`Guild` is intentionally wide because Discord uses one shape for several
endpoints. Nullable IDs and optional counts are pointers. `MaxEmojis` and
`MaxStickers` derive limits from `PremiumTier`; they are convenience methods,
not a statement that the current guild inventory is valid. `Feature` is a
string enum whose constants include community, discoverability, banners, role
icons, stickers, AutoMod, welcome screens, and monetization features.

The moderation models are `AutoModerationRule`,
`AutoModerationTriggerMetadata`, `AutoModerationAction`, and
`AutoModerationActionMetadata`. Onboarding uses `Onboarding`,
`OnboardingPrompt`, and `OnboardingPromptOption`. Events use
`ScheduledEvent`, `ScheduledEventEntityMetadata`, and
`ScheduledEventRecurrenceRule`. `StageInstance`, `Template`, `WelcomeScreen`,
`Widget`, and their nested types represent the remaining resource families.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/guilds"
)

func main() {
	var guild guilds.Guild
	if err := json.Unmarshal([]byte(`{"id":"1","name":"Example","premium_tier":2,"roles":[],"emojis":[],"features":["COMMUNITY"]}`), &guild); err != nil {
		panic(err)
	}
	fmt.Println(guild.Name, guild.Features[0], guild.MaxEmojis(), guild.MaxStickers())
}
```

## Creating And Using Guilds

Most models are constructed with struct literals or decoded from REST and
Gateway JSON. `GuildPreview` contains public counts and assets without the
full member and channel inventory. `UnavailableGuild` represents an outage
notice. `Integration` contains an external account, optional user, optional
application, scopes, sync state, and expiry settings; its custom unmarshaller
handles the API's nullable fields. `IntegrationApplication.UnmarshalJSON` is
available for the nested application response.

`Template` includes timestamps and a serialized source guild and has a custom
unmarshaller. `WelcomeScreenChannel.UnmarshalJSON`, `Widget.UnmarshalJSON`,
`WidgetChannel.UnmarshalJSON`, and `WidgetSettings.UnmarshalJSON` handle
nullable snowflakes in those responses.

## Using AutoMod And Onboarding

Trigger metadata can contain keyword filters, regex patterns, presets, allow
lists, mention limits, and raid protection. Actions may target a channel,
timeout duration, or custom message. Exempt roles and channels are
`snowflake.IDs`. Onboarding prompts carry channel and role choices, emoji
metadata, required state, and single-select behavior.

## Using Scheduled Events And Stages

Scheduled event times are strings on the response model because Discord can
return ISO-8601 values and the package preserves them exactly. Recurrence
rules use `Start`, optional `End`, frequency, interval, weekday/month fields,
and count. A stage instance identifies its guild and channel, topic, privacy,
and optional scheduled-event ID.

## Common Patterns

Use pointer fields to distinguish "not included" from zero. Treat `Features`
as an open string list even though known constants are provided. For changes,
send the dedicated REST parameter structs instead of reusing response models.
Use `MaxEmojis` and `MaxStickers` for UI hints, then handle API errors when a
write races another inventory change.

## Best Practices

Check `Unavailable` on a guild delete event. Do not retain a full guild as the
source of truth for member or channel data unless the required Gateway intents
are enabled. Preserve unknown feature strings and optional fields when
round-tripping models.

## Common Mistakes

`Guild.Region` is deprecated and may be nil. Counts prefixed with
`Approximate` are estimates. A nil `ChannelID` on a scheduled event is valid
for external-location events; inspect `EntityMetadata`. A template's source
guild is a serialized snapshot, not the live guild.

## API Walkthrough

The exported declarations are `Guild`, `GuildPreview`, `UnavailableGuild`,
`Feature` and all feature constants, `Guild.MaxEmojis`, `Guild.MaxStickers`,
and `Guild.UnmarshalJSON`; `AutoModerationRule`, `AutoModerationTriggerMetadata`,
`AutoModerationAction`, `AutoModerationActionMetadata`; `Integration`,
`IntegrationAccount`, `IntegrationApplication` and its unmarshaller;
`Onboarding`, `OnboardingPrompt`, `OnboardingPromptOption`; `ScheduledEvent`,
`ScheduledEventEntityMetadata`, `ScheduledEventRecurrenceRule`;
`StageInstance`; `Template` and its unmarshaller; `VanityURL`; `WelcomeScreen`,
`WelcomeScreenChannel` and its unmarshaller; `Widget`, `WidgetChannel`,
`WidgetSettings` and their unmarshalers.

## Examples

The Quick Start program is complete and runnable. REST resource methods are
grouped in [`../rest/endpoints.md`](../rest/endpoints.md), and typed Gateway
wrappers are in [`../events/`](../events/README.md).

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../events/`](../events/README.md)
- [`../roles/`](../roles/README.md)
- [`../emojis/`](../emojis/README.md)
