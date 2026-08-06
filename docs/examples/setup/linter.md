# Linter

## Overview

The Discord.js guide uses a linter to keep JavaScript consistent. In a Go project, formatting and baseline static analysis are built into the toolchain. `gofmt` is mandatory formatting, `go vet` catches suspicious constructs, and tests provide the behavioral check. Add a third-party linter only when the project agrees on its version and configuration.

## Architecture

Formatting changes source text deterministically. `go vet` analyzes packages and reports likely bugs without running the bot. `go test` compiles every package and runs tests. None of these commands need a Discord token; the executable itself only connects when `b.Run()` is reached.

## Prerequisites

- Go `1.26.4` or newer.
- A Go module containing the bot.
- `gofmt`, `go vet`, and `go test` available through the Go installation.
- A token and installed application only for the final runtime check.

## Quick Start

Create `main.go` with the complete program below, then run the checks before starting the bot:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
```

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("check", "Run a health check", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("The bot passed its runtime check."); err != nil {
			log.Printf("reply: %v", err)
		}
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	b.OnError(func(err error) {
		log.Printf("runtime error: %v", err)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Run `gofmt` on changed `.go` files. Run `go vet ./...` from the module root so imports and all packages are checked. Run `go test ./...` in CI and before release. A linter should inspect code; it should not contain a bot token or make network calls merely to analyze a package.

## Common Patterns

- Add `gofmt -w .` or an equivalent file list to the development workflow.
- Run `go vet ./...` and `go test ./...` on every pull request.
- Use `router.Validate()` in a startup test for statically assembled commands.
- Keep handlers small enough that static analysis and tests can exercise them.
- Add a context timeout around REST calls in code that is not using context helpers.

## Best Practices

- Make formatting a CI gate rather than a personal preference.
- Review lint suppressions individually and document the reason.
- Use `go test -race ./...` when shared state is touched by concurrent handlers.
- Keep generated code and generated command definitions separate from hand-written handlers.
- Never suppress a warning by removing error handling from an interaction response.

## Common Mistakes

### Incorrect

```go
_, _ = ctx.Reply("done")
```

### Correct

```go
if _, err := ctx.Reply("done"); err != nil {
	log.Printf("reply: %v", err)
}
```

### Incorrect

```bash
gofmt main.go
```

### Correct

```bash
```

The `-w` flag writes formatting changes, and the package-wide commands check more than the file currently open in an editor.

## API Walkthrough

- `gofmt` is the canonical Go formatter.
- `go vet` performs standard static checks.
- `go test` compiles packages and executes tests.
- `router.Validate` checks command names, descriptions, handlers, and option definitions.
- `b.OnError` reports runtime failures after the static checks have passed.

## Examples

- [Slash Commands](../slash-commands.md) demonstrates validation and middleware.
- [Gateway](../gateway.md) demonstrates error handling for event decoding.
- [Full Template](../full-template.md) shows a layout suitable for larger applications.

## Related Pages

- [Installation](installation.md)
- [Project Setup](../commands/project-setup.md)
- [Handling Commands](../commands/handling-commands.md)
- [Go command documentation](https://go.dev/doc/cmd)
