# Interactions

## Overview

The `interactions` package models incoming Discord interactions and the JSON
used to answer them. It includes `Interaction`, command definitions and option
types, callback data, response callback types, a slash-command builder,
Ed25519 request verification, and an `http.Handler` that receives and verifies
interaction webhooks automatically.

## Architecture

`Interaction` is the envelope: it carries IDs, type, token, version, optional
guild/channel/member/user/message data, permissions, locales, and raw `Data`.
`InteractionType` distinguishes ping, commands, components, autocomplete, and
modal submit. Decode command `Data` into
`ApplicationCommandInteractionData`, whose option `Value` is `interface{}` in
the repository. With the standard JSON decoder, numbers in an interface become
`float64`; represent snowflake-like values as strings or decode into a typed
field when precision matters.

`InteractionResponse` contains an `InteractionCallbackType` and optional
`InteractionCallbackData`. Callback types cover pong, immediate and deferred
messages, update-message, autocomplete, modal, premium-required, and launch
activity responses. Callback data can contain content, embeds, allowed
mentions, flags, components, attachments, choices, modal fields, thread name,
tags, and polls.

## Quick Start

```go
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/interactions"
)

func main() {
	command := interactions.NewSlashCommandBuilder("hello", "Say hello").
		AddStringOption("name", "Who to greet", false).
		SetContexts(int(interactions.InteractionContextTypeGuild)).
		Build()

	response := interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: &interactions.InteractionCallbackData{Content: "hello"},
	}
	encoded, _ := json.Marshal(response)

	seed := make([]byte, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	body := []byte(`{"type":1}`)
	timestamp := "1700000000"
	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)
	valid := interactions.VerifySignature(hex.EncodeToString(privateKey.Public().(ed25519.PublicKey)), timestamp, hex.EncodeToString(signature), body)
	fmt.Println(command.Name, string(encoded), valid)
}
```

## Creating Commands

`NewSlashCommandBuilder(name, description)` creates a chat-input command with
the corresponding type. Setters mutate the builder and return it. Convenience
methods add string, integer, boolean, user, channel, role, and mentionable
options; `AddStringOptionWithChoices` attaches predefined choices, and
`AddOption` accepts an arbitrary `ApplicationCommandOption`. `SetIntegrationTypes`
uses 0 for guild install and 1 for user install. `SetContexts` uses 0 for
guild, 1 for bot DM, and 2 for private channel. `Build` returns a value copy.
The builder performs no name, description, or option-count validation.

## Using Responses

Use callback type 4 for an immediate message, 5 to defer a command response,
6 to defer a component update, 7 to update a component message, 8 for
autocomplete choices, and 9 for a modal. A modal commonly uses
`components.ModalBuilder` and action rows containing text inputs. Set
`messages.FlagEphemeral` in callback data when the response should be
ephemeral; flags are bit values, not callback types.

## Common Patterns

Verify an interaction before unmarshaling or acting on its body. The signature
message is the UTF-8 timestamp concatenated directly with the raw request
body. `VerifySignature` expects hex accepted by `encoding/hex`, returns false
for invalid lengths or encoding, and never returns an error. It verifies the
cryptographic signature but does not check timestamp freshness.

Use `VerifyRequest` for incoming HTTP requests. It calls `VerifySignature`
and additionally rejects timestamps older or newer than `MaxTimestampSkew`
(5 minutes) relative to the verifier's clock, preventing replay attacks:

```go
valid := interactions.VerifyRequest(publicKeyHex, timestamp, signatureHex, body, time.Now())
```

## HTTP Server

The `interactions.Server` is an `http.Handler` that receives Discord
interaction webhooks, verifies their Ed25519 signature and timestamp
freshness, and dispatches them to a handler. It auto-responds to pings
(type 1) with a pong and defaults to a deferred response if the handler
returns nil.

```go
package main

import (
	"net/http"

	"github.com/discord-go/discord.go/interactions"
)

func main() {
	publicKey := "your-app-public-key"
	srv := interactions.NewServer(publicKey, func(i *interactions.Interaction) *interactions.InteractionResponse {
		switch i.Type {
		case interactions.InteractionTypeApplicationCommand:
			return &interactions.InteractionResponse{
				Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
				Data: &interactions.InteractionCallbackData{Content: "Hello!"},
			}
		}
		return nil // auto-defer
	})
	http.Handle("/interactions", srv)
	http.ListenAndServe(":8080", nil)
}
```

## Best Practices

Acknowledge an interaction within Discord's deadline, then use the interaction
token for follow-ups. Preserve raw `Data` until the interaction type is known.
Validate option names and choice values before registration. Treat tokens,
signature headers, and custom IDs as secrets or untrusted input respectively.
Use `interactions.Server` or `VerifyRequest` (not `VerifySignature` alone)
for incoming HTTP requests to prevent replay attacks.

## Common Mistakes

Do not verify a re-encoded JSON body; whitespace and field ordering affect the
signature. Do not claim `Value` preserves arbitrary numeric precision: it is an
`interface{}`. Do not send a bot Authorization header to an interaction
webhook operation that should use [`rest.RequestNoAuth`](../rest/requests.md).

## API Walkthrough

The exported API is `Interaction`, `InteractionType` and constants,
`ApplicationCommand`, `ApplicationCommandType` and constants,
`ApplicationCommandOption`, `ApplicationCommandOptionType` and constants,
`ApplicationCommandOptionChoice`, `ApplicationCommandInteractionData`,
`ApplicationCommandInteractionDataOption`, `InteractionCallbackData`,
`InteractionCallbackType` and constants, `InteractionResponse`,
`InteractionContextType` and constants, `SlashCommandBuilder` and all its
builder methods, `VerifySignature`, `VerifyRequest`, `MaxTimestampSkew`,
`Server`, `NewServer`, and `InteractionHandler`.

## Examples

The Quick Start program is complete and runnable. It builds a command,
serializes an immediate response, and signs/verifies a request locally.

## Related APIs

- [`../components/`](../components/README.md)
- [`../messages/`](../messages/README.md)
- [`../rest/`](../rest/README.md)
