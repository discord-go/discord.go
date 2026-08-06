# Contributing

## Overview

Contributions should improve the public Discord API, the high-level bot API,
documentation, examples, or test coverage without weakening existing behavior.

## Before You Start

1. Search existing packages, REST resources, and issues before adding a new abstraction.
2. Read the relevant Discord Developer Documentation page.
3. Check whether the feature belongs in a low-level package or the `bot` facade.
4. Avoid committing secrets, generated coverage files, binaries, or dependency caches.

## Implementation Rules

- Add low-level protocol and REST capabilities before high-level wrappers.
- Use typed request and response models.
- Preserve nullable fields with pointers where omission differs from `null`.
- Add contract tests for HTTP method, path, query, body, headers, and decoding.
- Use contexts for all network and collector operations.
- Keep public names consistent with the existing package architecture.
- Update the relevant Markdown guide and runnable example.

## Verification

Run the following before submitting a pull request:

```bash
gofmt -w path/to/changed/files.go
```

If a nested example module changes, run its tests from that module directory too.

## Pull Requests

Keep pull requests focused. Explain the Discord API source, wire-level changes,
compatibility considerations, and tests. Do not combine unrelated formatting
or generated-artifact changes with a feature.
