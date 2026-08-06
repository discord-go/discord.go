# Persistence Examples

## Overview

These guides explain how Discord.js guide users of Keyv or Sequelize can build the same application boundary with `storage.Store`.

## Architecture

Bot handlers should depend on a repository interface, not a concrete database. The storage layer serializes application records and can be backed by memory for tests or a durable database in production.

## Quick Start

Read [`keyv.md`](keyv.md) for a complete memory-backed program and [`sequelize.md`](sequelize.md) for a relational repository design.

## Best Practices

Use stable keys, explicit schemas, context timeouts, idempotent upserts, and migrations. Do not store tokens or raw Discord interaction data without a retention policy.

## Related Pages

- [`low-level/storage/`](../../low-level/storage/README.md)
- [`high-level/caching.md`](../../high-level/caching.md)
