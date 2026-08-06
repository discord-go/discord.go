# Partials And Cache

## Overview

Discord events can contain optional objects or only the IDs needed to identify
them. In `discord.go`, these shapes appear as pointers such as `Interaction.User`,
`Interaction.Member`, `Interaction.Channel`, and `Interaction.Guild`; there is
no separate Discord.js-style partials manager. Treat absent pointers as normal,
then use the optional cache and REST fetch helpers when a complete resource is
needed.

## Tutorial: Cache-Aside Fetch

1. Create a bounded `cache.MemoryCache` with a TTL.
2. Attach it using `bot.WithCache`.
3. Try `CachedUser`, `CachedMember`, or another typed lookup.
4. On a miss, call `FetchUser`, `FetchMember`, or a matching method with a
   timeout.
5. Treat the returned object as a snapshot and invalidate or refetch after
   mutations.

The built-in cache implements guild, channel, user, role, message, and member
typed interfaces. A custom cache can implement only the interfaces its
application needs.

## Complete Runnable Example

Copy to `examples/partials-cache/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/user-cache` with a user option twice.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("user-cache", "Inspect a cached user", func(ctx *bot.InteractionContext) {
		userID := ctx.GetUserID("user")
		if userID == 0 {
			_ = ctx.ReplyEphemeral("A user is required.")
			return
		}
		if user, ok := ctx.Bot.CachedUser(userID); ok {
			_ = ctx.Reply(fmt.Sprintf("Cache hit: %s", user.Username))
			return
		}
		requestCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
		defer cancel()
		user, err := ctx.Bot.FetchUser(requestCtx, userID)
		if err != nil {
			log.Printf("fetch user: %v", err)
			_ = ctx.ReplyEphemeral("The user could not be fetched.")
			return
		}
		_ = ctx.Reply(fmt.Sprintf("Cache miss; fetched %s", user.Username))
	}, interactions.ApplicationCommandOption{
		Type:        interactions.ApplicationCommandOptionTypeUser,
		Name:        "user",
		Description: "The user to fetch",
		Required:    true,
	})

	store := cache.NewMemoryCache(cache.WithTTL(10*time.Minute), cache.WithMaxSize(10000))
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithCache(store), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Optional Data Rules

Always nil-check `Member`, `User`, `Channel`, and `Guild` pointers. A DM has no
guild or member. A cache hit is not authoritative permission data, and a cache
miss does not prove that a resource does not exist. Use `FetchMember` or another
REST call when the operation requires fresh data.

## Common Mistakes

- Dereferencing an optional event pointer without checking it.
- Configuring a cache but ignoring the hit boolean.
- Using an unlimited cache for a large, long-running bot.
- Assuming Gateway hydration covers data for unrequested intents.
- Using stale cached permissions for a destructive action.

## Expected Result

The first `/user-cache` call fetches and stores the selected user. A later call
can use the typed cache hit until the TTL expires.

## Related Pages

- [Gateway Intents](gateway-intents.md)
- [Permissions](permissions.md)
- [Common Errors](common-errors.md)
