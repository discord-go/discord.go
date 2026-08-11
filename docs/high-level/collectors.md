# Collectors And Jobs

## Overview

Collectors are one-shot waits for a matching interaction or message. They are
useful for button confirmations, select menus, modal workflows, and temporary
conversation steps. The bot also exposes `Every` for a cancellable periodic job
owned by the current bot run.

## Architecture

`AwaitInteraction` and `AwaitMessage` register an in-memory collector with a
filter and a one-value result channel. Gateway dispatch publishes each typed
context to matching collectors; the first match removes that collector. A
context cancellation removes the collector and returns the context error.

`Every` derives a job context from the bot run context, ticks at the requested
interval, recovers job panics into the bot error path, and cancels jobs during
`Stop` or a run termination.

## Quick Start

This complete program sends a button from `/choose`, waits for the invoking user
to click it, and then reports the result. The wait is bounded by a timeout.

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("choose", "Start a one-click flow", func(ctx *bot.InteractionContext) {
		button := components.NewButtonBuilder().
			SetCustomID("choose:yes").SetLabel("Choose yes").
			SetStyle(components.ButtonStylePrimary).Build()
		row := components.NewActionRowBuilder().AddComponents(button).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("reply: %v", err)
			return
		}

		go func(owner bot.InteractionContext) {
			waitCtx, cancel := context.WithTimeout(owner.Context(), 30*time.Second)
			defer cancel()
			selected, err := owner.Bot.AwaitInteraction(waitCtx, func(next *bot.InteractionContext) bool {
				return next.CustomID() == "choose:yes" && next.User != nil && owner.User != nil && next.User.ID == owner.User.ID
			})
			if err != nil {
				log.Printf("choice: %v", err)
				return
			}
			_, _ = selected.Followup("You chose yes")
		}(*ctx)
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

The example copies the context value before starting the goroutine so the
handler does not depend on a mutable local variable. In production, also scope
by guild and triggering message when those fields are available.

## Creating/Configuration

Call `AwaitInteraction(ctx context.Context, filter InteractionFilter)` or
`AwaitMessage(ctx context.Context, filter MessageFilter)`. A nil filter accepts
the next event of that type, but a nil filter is rarely safe in a shared bot.
Use `context.WithTimeout` or `context.WithCancel` so an abandoned workflow does
not wait forever.

Call `Every(ctx, interval, job)` after the bot has started, normally from
`OnReady`, when the job should inherit the active run context. The returned
function cancels that job and is safe to call repeatedly.

## Using

### Basic: wait for one event

Build a filter around `CustomID`, `CommandName`, `Content`, or `Author.ID`.
Handle `context.DeadlineExceeded` as an expected timeout, not necessarily an
application error.

### Intermediate: build a wizard

After one result, register the next wait with a new context and a filter that
checks the same user and message. Avoid a single global collector for all users.

### Advanced: collect multiple interactions

Use `CollectInteractions` when a flow needs to receive multiple interactions
over time, such as live polls, reaction roles, or multi-step wizards. It
returns a channel and a stop function; the channel receives every matching
interaction until the context is cancelled or `stop()` is called.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

ch, stop := b.CollectInteractions(ctx, func(ic *bot.InteractionContext) bool {
    return ic.IsButton() && ic.CustomID() == "vote:yes"
})
defer stop()

for ic := range ch {
    log.Printf("vote from %s", ic.User().ID)
}
```

### Advanced: periodic work

Use `Every` for cleanup, metrics, or refresh work that belongs to the bot run.
Call the returned cancel function for feature-specific shutdown; `Stop` cancels
all remaining jobs.

## Common Patterns

- Filter on user, guild, channel, message, and custom ID as applicable.
- Use one collector per workflow step and one timeout for the whole workflow.
- Keep collection handlers short; persist state before sending the next prompt.
- Let the job context cancel HTTP calls made by an `Every` job.
- Record collector timeout counts separately from gateway errors.

## Best Practices

### Always bound a collector

Why: a user can close Discord or abandon a prompt.

Pros: no leaked collector registrations and predictable memory use.

Cons: timeout handling becomes part of the user experience.

### Prefer routes for durable behavior

Why: collectors are process-local and one-shot, while router routes are stable
dispatch definitions.

Pros: routes survive unrelated workflows and scale better across users.

Cons: routes require explicit state and authorization when the interaction ID
contains data.

### Cancel jobs deliberately

Why: `Every` runs concurrently and its callback can overlap with shutdown.

Pros: jobs stop with the bot and can be stopped early by feature code.

Cons: the callback must honor its context; cancellation cannot interrupt code
that ignores context.

## Common Mistakes

Incorrect: accepting the first user's click for a private workflow.

```go
selected, _ := b.AwaitInteraction(ctx, func(i *bot.InteractionContext) bool {
		return i.CustomID() == "confirm"
})
```

Correct: include the owner and the relevant message or guild.

```go
selected, _ := b.AwaitInteraction(ctx, func(i *bot.InteractionContext) bool {
	return i.CustomID() == "confirm" &&
		i.User != nil && i.User.ID == ownerID
})
```

Incorrect: starting a periodic job with a nil or zero interval.

```go
b.Every(context.Background(), 0, job)
```

Correct: use a positive interval and retain the cancellation function.

```go
cancelJob := b.Every(ctx, time.Minute, job)
defer cancelJob()
```

## API Walkthrough

- `InteractionFilter` is `func(*InteractionContext) bool`.
- `MessageFilter` is `func(*MessageContext) bool`.
- `ReactionFilter` is `func(*ReactionContext) bool`.
- `AwaitInteraction(context.Context, InteractionFilter) (*InteractionContext, error)`
  waits for one matching interaction.
- `CollectInteractions(context.Context, InteractionFilter) (<-chan *InteractionContext, func())`
  collects multiple matching interactions into a channel until the context
  is cancelled or the stop function is called. Useful for live polls,
  reaction roles, and multi-step wizards.
- `AwaitMessage(context.Context, MessageFilter) (*MessageContext, error)` waits
  for one matching message. Both return the context error on cancellation.
- `AwaitReaction(context.Context, ReactionFilter) (*ReactionContext, error)`
  waits for one matching reaction. Useful for confirmation prompts,
  reaction-based menus, and pagination.
- `ErrCollectorClosed` describes a closed collector; current waits primarily
  return the supplied context error when cancellation wins the select.
- `Every(context.Context, time.Duration, func(context.Context)) func()` starts
  a lifecycle-managed ticker job. Invalid intervals or nil jobs return a no-op
  cancellation function.

## Reaction Collector

`AwaitReaction` works like the other collectors: it registers a filter,
publishes each `MESSAGE_REACTION_ADD` event to matching collectors, and
returns the first match or the context error.

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

reaction, err := b.AwaitReaction(ctx, func(rc *bot.ReactionContext) bool {
    return rc.MessageID == targetMessageID && rc.Emoji.Name == "✅"
})
if err != nil {
    log.Printf("no reaction received: %v", err)
    return
}
log.Printf("confirmed by %s", reaction.UserID)
```

## Paginator

The `Paginator` creates a message with prev/next/stop buttons and edits
it in-place as the user navigates. Pages are provided as a slice of
`PaginatorPage` structs containing content and optional embeds.

```go
pages := []bot.PaginatorPage{
    {Content: "Page 1"},
    {Content: "Page 2"},
    {Content: "Page 3"},
}
p := b.NewPaginator(channelID, pages,
    bot.WithPaginatorTimeout(10*time.Minute),
    bot.WithPaginatorUser(userID), // restrict to one user
)
if err := p.Send(ctx); err != nil {
    log.Printf("paginator error: %v", err)
}
```

The paginator blocks until the timeout expires, the user clicks stop,
or the context is cancelled. Buttons are removed on exit. Because `Send`
blocks, it must be launched in a goroutine from interaction handlers —
acknowledge the interaction first, then call `Send` in a `go` statement.

For convenience, `SendAsync` wraps this pattern: it launches `Send` in a
goroutine and returns immediately, delivering any error to an optional
callback.

```go
ctx.ReplyEphemeral("Opening paginator…")
p.SendAsync(ctx, func(err error) {
    log.Printf("paginator error: %v", err)
})
```

## Examples

- [Collectors example](../examples/more-to-know/collectors.md)
- [Buttons](buttons.md)
- [Modals](modals.md)

## Related APIs

- [`lifecycle.md`](lifecycle.md) for job cancellation during shutdown.
- [`buttons.md`](buttons.md) and [`modals.md`](modals.md) for interaction flows.
- [`errors.md`](errors.md) for recovered job panics.
