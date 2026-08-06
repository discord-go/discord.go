## Summary

Describe the change and the Discord API or developer-experience problem it solves.

## Verification

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] Nested example modules tested when changed
- [ ] Documentation updated

## API Review

- [ ] Low-level implementation exists before high-level wrappers
- [ ] Request/response serialization is covered
- [ ] Nullable and optional fields are preserved
- [ ] No secrets or generated artifacts are included

## Compatibility

- [ ] Existing public APIs remain compatible, or the breaking change is explicitly justified
- [ ] Discord API version and official documentation links are included for REST/Gateway changes
- [ ] Nullable, optional, and snowflake fields preserve Discord wire semantics
- [ ] Multipart field names and attachment retention behavior are tested where applicable
- [ ] High-level wrappers use the new low-level API instead of duplicating routes

## Documentation

- [ ] Relevant `docs/low-level/` page updated
- [ ] Relevant `docs/high-level/` page updated
- [ ] Runnable example updated when developer experience changes
