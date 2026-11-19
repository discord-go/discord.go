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

## Current inventory

The 2026-11-19 inventory found **191 exact matches out of 250** specification
operations (76.4%) and 194 statically detected REST operations. The previous
2026-11-18 run found 166/250; the increase comes from scanner fixes, not new
endpoints:

- local integer-to-string helpers such as `rest.itoa` are now treated as path
  placeholders;
- function parameters such as `code string` are now treated as placeholders,
  so `DeleteInvite` is detected;
- the scanner now recognizes `RequestNoAuth`, `RequestMultipartNoAuth`,
  `RequestMultipartFormNamedFile`, `request`, `requestMultipart`,
  `requestMultipartForm`, `requestRaw`, and `requestWithDecoder`, which covers
  webhook, interaction, and no-auth routes.

These are scanner matches, not a percentage of implemented functionality.

## Review backlog

The remaining 59 unmatched operations fall into these categories.

### Detection limitations

- **OAuth2 routes:** `GET /oauth2/@me`, `GET /oauth2/applications/@me`,
  `GET /oauth2/keys`, and `GET /oauth2/userinfo` live in the `oauth2` package,
  which the scanner does not walk. They are not missing from the library.
- **Pin path alias:** the REST helpers use
  `/channels/{}/messages/pins` and `/channels/{}/messages/pins/{}`, while the
  preview specification lists `/channels/{}/pins` and `/channels/{}/pins/{}`.
  Compare against the current Discord documentation before changing either
  side; this is an alias, not a missing endpoint.
- **Dynamic routes:** routes assembled through helpers outside the recognized
  set, or through values that are not statically resolvable, are reported as
  missing. Add a scanner test when a real route is found this way.

### Genuine missing endpoints

These have no matching helper in `rest` after reviewing the source. They are
candidates for implementation, not confirmed regressions:

- guild templates (`GET/POST/PUT/PATCH/DELETE`);
- guild integration deletion (`DELETE /guilds/{}/integrations/{}`);
- scheduled-event exceptions and user counts;
- guild join requests (`GET /guilds/{}/requests`, `PATCH .../requests/{}`);
- new-member welcome (`GET /guilds/{}/new-member-welcome`);
- channel and guild permission member/role lookups;
- thread search (`GET /channels/{}/threads/search`);
- sticker packs (`GET /sticker-packs/{}`) and guild sticker creation;
- guild widget and widget image;
- incident actions (`PUT /guilds/{}/incident-actions`);
- application lookup by ID and attachment upload;
- group-DM recipient add/remove.

### Out of scope

- Lobby operations (`/lobbies/...`) and partner-SDK operations
  (`/partner-sdk/...`) are partner-only APIs. Confirm authentication and
  product scope before adding them to a general bot library.

### Model matching

Exact-name schema comparison misses renamed and shared types. Compare selected
fields and nullability manually before asserting parity. The current report
matches 6 of 542 schemas by exact name; that number is not a model-coverage
metric.

## Regression checks

`scripts/api_coverage/main_test.go` protects the scanner patterns that keep the
inventory honest:

- `TestRestOperationsFindsRepresentativeRoutes` runs the scanner over the real
  `rest` package and asserts representative routes, including the poll-answer
  route, invite deletion, webhook token access, and interaction callbacks.
- `TestRestOperationsFindsLocalItoaRoute` covers `rest.itoa`.
- `TestRestOperationsFindsStringParameterRoute` covers raw string parameters.
- `TestRestOperationsFindsNoAuthRoute` covers `RequestNoAuth`.
- `TestRestOperationsFindsEscapedReaction` covers `url.PathEscape` and query
  helpers.

The scheduled job fails if the inventory command fails. The report depends on
Discord's preview specification and network availability; a failed download is
not evidence of a change in API coverage. A drop in the exact-match count
without a corresponding source change is a scanner regression and should be
investigated before the report is accepted.
