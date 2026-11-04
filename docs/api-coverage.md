# Discord API Coverage Inventory

The scheduled **API coverage inventory** workflow downloads Discord's standard
OpenAPI file and compares its REST method/path pairs with requests detected in
this repository's `rest` package. Run the same report locally with:

```sh
go run ./scripts/api_coverage
```

For an offline fixture or a saved specification, pass
`--spec path/to/openapi.json`. The report also shows exact-name matches between
OpenAPI schemas and exported Go structs. That name comparison is only a review
hint: it does not establish field parity, account for intentional aliases, or
verify routes built dynamically, Gateway events, or semantic behavior.

Discord describes its OpenAPI specification as a public preview that may change
without notice and may contain inaccuracies. Use the
[Discord developer documentation](https://discord.com/developers/docs) as the
authority when the specification conflicts with it. This inventory is a way to
find review candidates, not a completeness guarantee.
