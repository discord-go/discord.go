#!/usr/bin/env bash
set -euo pipefail

version=${1:?Usage: verify-release.sh vX.Y.Z[-suffix]}
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo 'Expected a Go module version such as v0.16.1-wid' >&2
  exit 2
fi
go_bin="$(go env GOROOT)/bin/go"
workdir=$(mktemp -d)
cleanup() {
  chmod -R u+w "$workdir" 2>/dev/null || true
  rm -rf "$workdir"
}
trap cleanup EXIT
cd "$workdir"
cat > go.mod <<MOD
module example.com/discordgo-release-check

go 1.22.0

require github.com/discord-go/discord.go $version
MOD
cat > consumer_test.go <<'GO'
package consumer

import (
    _ "github.com/discord-go/discord.go/bot"
    "github.com/discord-go/discord.go/dave-go/mls-go/group"
    "github.com/discord-go/discord.go/ratelimit/redislimiter"
    "github.com/discord-go/discord.go/rest"
    _ "github.com/discord-go/discord.go/voice"
    _ "github.com/discord-go/discord.go/voiceplayer"
    "testing"
)

var _ = group.NewGroup
var _ = redislimiter.New
var _ = (*rest.Client).GetGuildTemplate

func TestImports(t *testing.T) {}
GO
export GOWORK=off
export GOMODCACHE="$workdir/modcache"
export GOTOOLCHAIN=local
# The module is fetched directly from its Git repository. Skip the public
# checksum database for this module path so a sumdb outage or a restricted
# network does not mask the packaging result we are checking.
export GOPRIVATE=github.com/discord-go/*
# Bypass stale negative proxy results immediately after new tags are pushed.
GOPROXY=direct "$go_bin" mod tidy
GOPROXY=direct "$go_bin" test ./...
"$go_bin" list -m all
