# Advanced Examples

## Overview

Advanced examples cover OAuth2 browser authorization and Gateway sharding. These topics matter when a bot becomes a service used by multiple guilds or users.

## Architecture

OAuth2 is an HTTP flow separate from bot-token Gateway access. Sharding divides Gateway sessions while keeping command registration and application state coordinated.

## Quick Start

Read [`oauth2.md`](oauth2.md) and [`sharding.md`](sharding.md) for complete programs and operational requirements.

## Best Practices

Use cryptographically random OAuth2 state values, store secrets server-side, scope OAuth requests narrowly, and test sharding with a disposable application before production rollout.

## Related Pages

- [`low-level/oauth2/`](../../low-level/oauth2/README.md)
- [`low-level/gateway/shards.md`](../../low-level/gateway/shards.md)
