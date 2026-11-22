// Package ratelimit provides rate limiting for Discord API requests.
//
// The default limiter uses a 50-request-per-second token bucket with a burst
// of 50, and tracks route limits reported by Discord. Share one limiter
// between REST clients using the same bot token. It coordinates only within one
// process; separate processes need their own shared coordination strategy.
// Unauthenticated requests can also share an IP-based limit.
// A custom [Limiter] may be supplied to a REST client when needed.
//
// See https://docs.discord.com/developers/topics/rate-limits.
package ratelimit
