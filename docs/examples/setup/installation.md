# Installation

## Overview

discord.go is a Go module, so installation means creating a module, selecting a compatible Go toolchain, and importing the packages used by your bot. This is the discord.go version of the Discord.js Installation topic: there is no global CLI or `node_modules` directory to configure.

## Architecture

The root `go.mod` declares module path `discord.go` and Go version `1.26.4`. A consumer module normally imports packages such as `discord.go/bot` and resolves them through a release or a local `replace` directive. Go records selected dependency versions in `go.mod` and checksums in `go.sum`.

## Prerequisites

- Go `1.26.4` or a compatible newer Go toolchain.
- Git if installing from a checkout.
- A bot token and installed test application for running the result.
- Network access to download module dependencies on the first build.

## Quick Start

From the repository root, the shortest complete run is:

```bash
go test ./...
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/ping
```

For a separate application module, use a local replacement while developing against this checkout:

```bash
mkdir my-bot
cd my-bot
go mod init example.com/my-bot
go mod edit -replace discord.go=/home/mlinux/Desktop/discord.go
go get discord.go
```

Save this complete `main.go` in that module:

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
	router.Command("version", "Report the bot version", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("This bot uses discord.go."); err != nil {
			log.Printf("reply: %v", err)
		}
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Use the repository's declared module path when importing local packages. If your application lives inside this repository, no `replace` directive is needed. If it lives elsewhere, use a published version when available; use `go mod edit -replace` only for local development or a deliberate fork. Run `go mod tidy` after changing imports.

## Common Patterns

- Run `go test ./...` after installing or upgrading.
- Keep application code in its own module or under `examples/` rather than editing library packages for a bot feature.
- Use `go run .` for the current application package.
- Use `go list -m all` to inspect selected module versions.
- Keep `go.mod` and `go.sum` under source control, but never commit secrets.

## Best Practices

- Pin intentional dependency upgrades through normal Go module commands.
- Use a reproducible Go toolchain in CI.
- Prefer the public package APIs over reaching into `internal/` packages.
- Run `gofmt` and `go vet ./...` before review.
- Keep local `replace` directives out of deployable application modules unless the deployment intentionally builds from that checkout.

## Common Mistakes

### Incorrect

```go
import "github.com/discordjs/discord.js"
```

### Correct

```go
import "github.com/discord-go/discord.go/bot"
```

### Incorrect

```bash
go run main.go
```

### Correct

```bash
```

Running the package with `go run .` includes all Go files in the package and catches module/import problems earlier.

## API Walkthrough

- `go mod init` creates the application module.
- `go mod edit -replace` points `discord.go` imports at a local checkout.
- `go get` adds the module requirement.
- `bot`, `intents`, and `interactions` are public packages used by normal applications.
- `go run .` builds and executes the complete main package.
- `go test ./...` compiles packages and runs their tests without connecting to Discord.

## Examples

- [Basic Client](../basic-client.md) is the smallest source-backed bot.
- [Project Setup](../commands/project-setup.md) shows a maintainable application layout.
- [`go.mod`](../../../go.mod) documents this checkout's module and toolchain.

## Related Pages

- [App Setup](app-setup.md)
- [Adding Your App](adding-your-app.md)
- [Linter](linter.md)
- [Commands](../commands/README.md)
