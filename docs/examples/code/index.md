# Runnable code

These examples use Go 1.22 or newer. From a repository checkout, set
`DISCORD_TOKEN` in your environment and run the desired example from the
repository root:

| Example | Command | Demonstrates |
| --- | --- | --- |
| Ping | `go run ./docs/examples/code/ping` | Slash and prefix commands |
| Moderation | `go run ./docs/examples/code/moderation` | Permission checks and moderation REST calls |
| Welcome | `go run ./docs/examples/code/welcome` | Guild member events and welcome messages |
| Tickets | `go run ./docs/examples/code/tickets` | Private channels and permission overwrites |
| Slash commands | `go run ./docs/examples/code/slash_commands` | Command registration and options |
| Components | `go run ./docs/examples/code/v2_components` | Components V2 interactions |
| Shard supervisor | `go run ./docs/examples/code/shard_supervisor` | Multiple worker processes with explicit shard IDs |

Use a test server and grant the bot only the permissions needed for the example.
The ping example's prefix command requires the Message Content intent. The
welcome example requires the Guild Members privileged intent and a
`WELCOME_CHANNEL_ID`. The tickets example needs Manage Channels and Manage Roles
to create private channels. Read the source and linked guide for each example's
setup and behavior. The shard supervisor example needs deployment-level Identify
and ownership coordination; see [Gateway Sharding](../advanced/sharding.md)
before using it.

For larger starting points, see the [bot templates](../templates/index.md).
CI compiles the runnable programs with the root module and separately tests the
nested template modules.
