<p align="center">
  <img src="docs/public/logo.svg" alt="discord.go" width="180">
</p>

<h1 align="center">discord.go</h1>

<p align="center">
  A typed Go library for building Discord applications.<br>
  Low-level Discord API foundation and high-level bot facade.
</p>

<p align="center">
  <a href="https://github.com/discord-go/discord.go/actions/workflows/ci.yml"><img src="https://github.com/discord-go/discord.go/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/discord-go/discord.go"><img src="https://pkg.go.dev/badge/github.com/discord-go/discord.go.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/discord-go/discord.go"><img src="https://goreportcard.com/badge/github.com/discord-go/discord.go" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/Discord%20API-v10-5865f2.svg" alt="Discord API v10">
  <img src="https://img.shields.io/badge/Go-1.26+-00add8.svg" alt="Go Version">
  <img src="https://img.shields.io/badge/version-v0.11.2--coffee-blue.svg" alt="Version">
  <img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="License">
</p>

---

## Why choose discord.go?

- **Two-layer architecture.** A typed low-level protocol/REST foundation and a
  high-level `bot` facade with commands, middleware, collectors, and lifecycle
  management. Use the facade for application code, drop to the low-level API
  for protocol work.
- **Gateway v10 with sharding.** Dispatch, heartbeat, reconnect, resume,
  zlib-stream compression, multi-shard support with concurrency buckets, and
  automatic shard count detection.
- **Voice with DAVE.** WebSocket, UDP, RTP, AES-256-GCM transport encryption,
  Opus frame handling, and DAVE/MLS end-to-end encryption support.
- **Components V2.** Full support for Discord's Components V2 message model
  including containers, media galleries, separators, and section components.
- **Typed throughout.** Every REST resource, gateway event, and interaction
  type has a typed Go model. Nullable fields use pointers where omission
  differs from `null`.
- **Production-oriented.** Rate limiting with bucket hashing, invalid-request
  tracking to prevent Cloudflare bans, audit-log reasons, multipart uploads,
  and context-aware cancellation on all REST calls.

## Installation

```bash
go get github.com/discord-go/discord.go
```

Requires Go 1.26 or newer. The library targets Discord API v10.

## Quick Start

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
    router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
        if err := ctx.Reply("Pong!"); err != nil {
            log.Printf("reply: %v", err)
        }
    })

    client := bot.New(token,
        bot.WithIntents(intents.Guilds),
        bot.WithRouter(router),
    )
    if err := client.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Documentation

The complete documentation site is at
[discord-go.github.io/discord.go](https://discord-go.github.io/discord.go/).

Start with the [high-level guides](docs/high-level/) for application code or
the [low-level guides](docs/low-level/) for protocol and REST work. Runnable
examples are organized under [docs/examples/](docs/examples/).

API reference is available on
[pkg.go.dev](https://pkg.go.dev/github.com/discord-go/discord.go).

See the [FAQ](docs/faq.md) and [Troubleshooting](docs/examples/troubleshooting.md)
pages for common questions and solutions.

## Development

```bash
go test ./...
go test -race ./...
go vet ./...
```

Nested example modules can be tested independently from their directories.
See [`CONTRIBUTING.md`](CONTRIBUTING.md) before opening a pull request.

## Security

Never commit bot tokens, OAuth client secrets, webhook tokens, database
credentials, or generated coverage and build artifacts. Report security issues
privately via [GitHub's security advisory system](SECURITY.md) rather than
opening a public issue. See the [security policy](SECURITY.md) for details.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
