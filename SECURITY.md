# Security Policy

## Reporting a Vulnerability

Do not open a public GitHub issue for security vulnerabilities.

To report a vulnerability, use GitHub's private vulnerability reporting:

1. Go to the [repository Security tab](https://github.com/discord-go/discord.go/security/advisories/new).
2. Click **Report a vulnerability**.
3. Provide a description, affected package or file, reproduction steps, and impact assessment.

You can also email the maintainer directly. Include enough detail to reproduce
or verify the issue, and redact any live credentials (bot tokens, OAuth
secrets, webhook tokens) from your report.

## Response Time

- **Acknowledgement:** within 48 hours.
- **Initial assessment:** within 7 days.
- **Fix or mitigation:** depends on severity, but target is 30 days for
  high-severity issues.

## Scope

Security issues in the `discord.go` library and its first-party packages are
in scope, including:

- Interaction signature verification (`interactions` package)
- Token handling and credential storage
- OAuth2 flows (`oauth2` package)
- Gateway authentication (`gateway` package)
- REST authentication (`rest` package)
- Voice encryption (`voice` package)

## Out of Scope

- Vulnerabilities in third-party dependencies (report to the upstream project)
- Issues in bots built with this library (contact the bot author)
- Social engineering or physical attacks

## Best Practices for Users

- Never commit bot tokens, OAuth client secrets, or webhook tokens to version
  control. Load them from environment variables or a secrets manager.
- Use `interactions.VerifyRequest` (not `VerifySignature`) to validate incoming
  interaction webhooks. `VerifyRequest` enforces both the Ed25519 signature and
  timestamp freshness to prevent replay attacks.
- Set a timeout on your HTTP server's read/write operations.
- Use HTTPS for all webhook endpoints.
