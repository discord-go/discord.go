package bot

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

// Validate runs application validation before a command handler. A non-nil
// error is returned to the user as an ephemeral message.
func Validate(check func(*InteractionContext) error) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if check == nil {
				next(ctx)
				return
			}
			if err := check(ctx); err != nil {
				replyEphemeral(ctx, err.Error())
				return
			}
			next(ctx)
		}
	}
}

// RequirePermissions creates middleware that checks that the invoking member
// has all specified Discord permissions before executing the command.
func RequirePermissions(perms permissions.Permission) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.Member == nil {
				replyEphemeral(ctx, "You must use this command in a server.")
				return
			}
			if !ctx.Member.Permissions.HasAll(perms) {
				replyEphemeral(ctx, "You do not have permission to use this command.")
				return
			}
			next(ctx)
		}
	}
}

// RequireAnyPermissions allows a command when the member has at least one of
// the supplied permission bits.
func RequireAnyPermissions(perms permissions.Permission) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.Member == nil {
				replyEphemeral(ctx, "You must use this command in a server.")
				return
			}
			if !ctx.Member.Permissions.Has(perms) {
				replyEphemeral(ctx, "You do not have permission to use this command.")
				return
			}
			next(ctx)
		}
	}
}

// RequireBotPermissions checks the permissions Discord granted to the bot for
// the invoking interaction.
func RequireBotPermissions(perms permissions.Permission) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			actual, err := strconv.ParseUint(ctx.AppPermissions, 10, 64)
			if err != nil || !permissions.Permission(actual).HasAll(perms) {
				replyEphemeral(ctx, "The bot does not have permission to complete this command.")
				return
			}
			next(ctx)
		}
	}
}

// RequireRole creates middleware that ensures the invoking member has a
// specific role before the command executes.
func RequireRole(roleID snowflake.ID) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.Member == nil {
				replyEphemeral(ctx, "You must use this command in a server.")
				return
			}
			for _, role := range ctx.Member.Roles {
				if role == roleID {
					next(ctx)
					return
				}
			}
			replyEphemeral(ctx, "You do not have the required role to use this command.")
		}
	}
}

// RequireAnyRole allows a command when the member has at least one role.
func RequireAnyRole(roleIDs ...snowflake.ID) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.Member == nil {
				replyEphemeral(ctx, "You must use this command in a server.")
				return
			}
			for _, memberRole := range ctx.Member.Roles {
				for _, requiredRole := range roleIDs {
					if memberRole == requiredRole {
						next(ctx)
						return
					}
				}
			}
			replyEphemeral(ctx, "You do not have a required role to use this command.")
		}
	}
}

// RequireOwners restricts a command to a configured owner allowlist.
func RequireOwners(ownerIDs ...snowflake.ID) Middleware {
	allowed := make(map[snowflake.ID]struct{}, len(ownerIDs))
	for _, id := range ownerIDs {
		allowed[id] = struct{}{}
	}
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			userID := snowflake.ID(0)
			if ctx.User != nil {
				userID = ctx.User.ID
			} else if ctx.Member != nil && ctx.Member.User != nil {
				userID = ctx.Member.User.ID
			}
			if _, ok := allowed[userID]; !ok {
				replyEphemeral(ctx, "Only a bot owner can use this command.")
				return
			}
			next(ctx)
		}
	}
}

// PrefixGuildOnly restricts a prefix command to guild messages.
func PrefixGuildOnly() PrefixMiddleware {
	return func(next PrefixHandler) PrefixHandler {
		return func(ctx *MessageContext, args []string) {
			if ctx.GuildID == 0 {
				_, _ = ctx.Reply("This command can only be used in a server.")
				return
			}
			next(ctx, args)
		}
	}
}

// RequirePrefixPermissions checks cached member permissions for a prefix
// command. Configure WithCache for this middleware to resolve members.
func RequirePrefixPermissions(perms permissions.Permission) PrefixMiddleware {
	return func(next PrefixHandler) PrefixHandler {
		return func(ctx *MessageContext, args []string) {
			if ctx.GuildID == 0 || ctx.Author == nil || ctx.Bot == nil {
				_, _ = ctx.Reply("You do not have permission to use this command.")
				return
			}
			member, ok := ctx.Bot.CachedMember(ctx.GuildID, ctx.Author.ID)
			if !ok || !member.Permissions.HasAll(perms) {
				_, _ = ctx.Reply("You do not have permission to use this command.")
				return
			}
			next(ctx, args)
		}
	}
}

// GuildOnly restricts a command to guild channels.
func GuildOnly() Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.GuildID == nil {
				replyEphemeral(ctx, "This command can only be used in a server.")
				return
			}
			next(ctx)
		}
	}
}

// OwnerOnly restricts a command to the current guild owner.
func OwnerOnly() Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			if ctx.Member == nil || ctx.GuildID == nil || ctx.Guild == nil || ctx.Member.User == nil {
				replyEphemeral(ctx, "This command can only be used by the server owner.")
				return
			}
			if ctx.Member.User.ID != ctx.Guild.OwnerID {
				replyEphemeral(ctx, "Only the server owner can use this command.")
				return
			}
			next(ctx)
		}
	}
}

// Cooldown allows one invocation per user during the supplied duration. The
// returned middleware is safe to share across commands and bot instances.
func Cooldown(duration time.Duration) Middleware {
	if duration <= 0 {
		return func(next CommandHandler) CommandHandler { return next }
	}
	var mu sync.Mutex
	lastUsed := make(map[string]time.Time)
	return func(next CommandHandler) CommandHandler {
		return func(ctx *InteractionContext) {
			key := cooldownKey(ctx)
			now := time.Now()
			mu.Lock()
			last, exists := lastUsed[key]
			if !exists || now.Sub(last) >= duration {
				lastUsed[key] = now
				mu.Unlock()
				next(ctx)
				return
			}
			remaining := duration - now.Sub(last)
			mu.Unlock()
			replyEphemeral(ctx, fmt.Sprintf("Please wait %s before using this command again.", remaining.Round(time.Second)))
		}
	}
}

// PrefixCooldown limits a prefix command per author and guild.
func PrefixCooldown(duration time.Duration) PrefixMiddleware {
	if duration <= 0 {
		return func(next PrefixHandler) PrefixHandler { return next }
	}
	var mu sync.Mutex
	lastUsed := make(map[string]time.Time)
	return func(next PrefixHandler) PrefixHandler {
		return func(ctx *MessageContext, args []string) {
			userID := snowflake.ID(0)
			if ctx.Author != nil {
				userID = ctx.Author.ID
			}
			key := userID.String() + ":" + ctx.GuildID.String() + ":" + ctx.Content
			now := time.Now()
			mu.Lock()
			last, exists := lastUsed[key]
			if !exists || now.Sub(last) >= duration {
				lastUsed[key] = now
				mu.Unlock()
				next(ctx, args)
				return
			}
			remaining := duration - now.Sub(last)
			mu.Unlock()
			_, _ = ctx.Reply(fmt.Sprintf("Please wait %s before using this command again.", remaining.Round(time.Second)))
		}
	}
}

func cooldownKey(ctx *InteractionContext) string {
	userID := snowflake.ID(0)
	if ctx.User != nil {
		userID = ctx.User.ID
	} else if ctx.Member != nil && ctx.Member.User != nil {
		userID = ctx.Member.User.ID
	}
	guildID := snowflake.ID(0)
	if ctx.GuildID != nil {
		guildID = *ctx.GuildID
	}
	return userID.String() + ":" + guildID.String() + ":" + ctx.CommandName()
}

func replyEphemeral(ctx *InteractionContext, content string) {
	if err := ctx.ReplyEphemeral(content); err != nil && ctx.Bot != nil {
		ctx.Bot.reportError(err)
	}
}
