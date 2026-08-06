# Handling Commands

## Overview

Handling Commands is where a definition becomes user-visible behavior. A handler must acknowledge an interaction quickly with `Reply`, `ReplyEphemeral`, `Defer`, `Update`, or `ShowModalBuilder`. After a deferral, use `Followup` or edit the original response. This page also shows the separate prefix-command path.

## Architecture

Discord allows one initial interaction response. `InteractionContext` records whether that response was accepted and returns `bot.ErrInteractionAlreadyResponded` for a second initial response. Deferred responses reserve the interaction while work continues. Prefix commands are message events routed by `bot.Router`; their handler receives `MessageContext` and parsed arguments.

## Prerequisites

- A bot token and installed application.
- `Guilds` enabled for slash commands.
- `GuildMessages` and the privileged `MessageContent` intent for the prefix example.
- Go `1.26.4` or newer.

## Quick Start

This complete program handles a fast slash response, a deferred response, and a `!say` prefix command:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
```

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("fast", "Reply immediately", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Immediate response."); err != nil {
			log.Printf("fast reply: %v", err)
		}
	})
	router.Command("slow", "Demonstrate a deferred response", func(ctx *bot.InteractionContext) {
		if err := ctx.Defer(); err != nil {
			log.Printf("defer: %v", err)
			return
		}
		time.Sleep(250 * time.Millisecond)
		if _, err := ctx.Followup("The deferred work finished."); err != nil {
			log.Printf("follow-up: %v", err)
		}
	}, interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeString, Name: "note", Description: "Optional note"})
	router.Prefix("say", func(ctx *bot.MessageContext, args []string) {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text == "" {
			text = "nothing to say"
		}
		if _, err := ctx.Reply(fmt.Sprintf("You said: %s", text)); err != nil {
			log.Printf("prefix reply: %v", err)
		}
	}).MinArgs(1).Usage("<text>")

	b := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent),
		bot.WithRouter(router),
	)
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Use `Reply` for work that is already ready, `ReplyEphemeral` for private results, and `Defer` before bounded work that may exceed the initial interaction deadline. A follow-up is a webhook message associated with the interaction. For component handlers, use `Update` or `DeferUpdate` when the original message should change. Prefix commands parse quoted arguments and pass everything after the command name to the handler.

## Common Patterns

- Return immediately when an initial response or deferral fails.
- Use `ctx.Followup` after `ctx.Defer`, never a second `ctx.Reply`.
- Use `ctx.EditReply` to replace a deferred original response.
- Add `MinArgs`, `Usage`, `Validate`, and `Aliases` to prefix commands.
- Apply `bot.GuildOnly` and permissions middleware before administrative work.

## Best Practices

- Acknowledge interactions within Discord's response window.
- Put deadlines on REST and database work after acknowledgement.
- Validate user input before including it in a message or using it in an action.
- Keep prefix commands optional; slash commands provide discoverability and typed options.
- Log failed responses so a transient REST failure is observable.

## Common Mistakes

### Incorrect

```go
router.Command("slow", "Slow work", func(ctx *bot.InteractionContext) {
	time.Sleep(5 * time.Second)
	_ = ctx.Reply("done")
})
```

### Correct

```go
router.Command("slow", "Slow work", func(ctx *bot.InteractionContext) {
	if err := ctx.Defer(); err != nil {
		return
	}
	// Perform bounded work here.
	_, _ = ctx.Followup("done")
})
```

### Incorrect

```go
if err := ctx.Reply("first"); err == nil {
	_ = ctx.Reply("second")
}
```

### Correct

```go
if err := ctx.Reply("first"); err != nil {
	log.Printf("initial response: %v", err)
	return
}
_, _ = ctx.Followup("second")
```

## API Walkthrough

- `ctx.Reply` sends the public initial response.
- `ctx.ReplyEphemeral` sends an invoker-only initial response.
- `ctx.Defer` acknowledges and shows a thinking state.
- `ctx.Followup` sends a message after acknowledgement.
- `ctx.EditReply` edits the original response.
- `ctx.UpdateContent` acknowledges a component by editing its source message.
- `ctx.HasResponded`, `Deferred`, and `Replied` expose response state.
- `PrefixCommand.MinArgs`, `Usage`, `Validate`, and `Aliases` shape message commands.

## Examples

- [Buttons](../buttons.md) updates a component message.
- [Modals](../modals.md) acknowledges with a modal and handles its submission.
- [Moderation](../moderation.md) defers before REST actions.
- [Collectors](../collectors.md) scopes follow-up interaction workflows.

## Related Pages

- [Creating Commands](creating-commands.md)
- [Handling Events](handling-events.md)
- [Deploying Commands](deploying-commands.md)
- [Interactions](../../low-level/interactions/README.md)
