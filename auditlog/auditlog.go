package auditlog

import (
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// AuditLog represents a Discord audit log object.
type AuditLog struct {
	ApplicationCommands  []interactions.ApplicationCommand `json:"application_commands"`
	AuditLogEntries      []AuditLogEntry                   `json:"audit_log_entries"`
	AutoModerationRules  []guilds.AutoModerationRule       `json:"auto_moderation_rules"`
	GuildScheduledEvents []guilds.ScheduledEvent           `json:"guild_scheduled_events"`
	Integrations         []guilds.Integration              `json:"integrations"`
	Threads              []channels.Channel                `json:"threads"`
	Users                []users.User                      `json:"users"`
	Webhooks             []channels.Webhook                `json:"webhooks"`
}

// AuditLogEvent represents the type of an audit log event.
type AuditLogEvent int

const (
	GUILD_UPDATE                                AuditLogEvent = 1
	CHANNEL_CREATE                              AuditLogEvent = 10
	CHANNEL_UPDATE                              AuditLogEvent = 11
	CHANNEL_DELETE                              AuditLogEvent = 12
	CHANNEL_OVERWRITE_CREATE                    AuditLogEvent = 13
	CHANNEL_OVERWRITE_UPDATE                    AuditLogEvent = 14
	CHANNEL_OVERWRITE_DELETE                    AuditLogEvent = 15
	MEMBER_KICK                                 AuditLogEvent = 20
	MEMBER_PRUNE                                AuditLogEvent = 21
	MEMBER_BAN_ADD                              AuditLogEvent = 22
	MEMBER_BAN_REMOVE                           AuditLogEvent = 23
	MEMBER_UPDATE                               AuditLogEvent = 24
	MEMBER_ROLE_UPDATE                          AuditLogEvent = 25
	MEMBER_MOVE                                 AuditLogEvent = 26
	MEMBER_DISCONNECT                           AuditLogEvent = 27
	BOT_ADD                                     AuditLogEvent = 28
	ROLE_CREATE                                 AuditLogEvent = 30
	ROLE_UPDATE                                 AuditLogEvent = 31
	ROLE_DELETE                                 AuditLogEvent = 32
	INVITE_CREATE                               AuditLogEvent = 40
	INVITE_UPDATE                               AuditLogEvent = 41
	INVITE_DELETE                               AuditLogEvent = 42
	WEBHOOK_CREATE                              AuditLogEvent = 50
	WEBHOOK_UPDATE                              AuditLogEvent = 51
	WEBHOOK_DELETE                              AuditLogEvent = 52
	EMOJI_CREATE                                AuditLogEvent = 60
	EMOJI_UPDATE                                AuditLogEvent = 61
	EMOJI_DELETE                                AuditLogEvent = 62
	MESSAGE_DELETE                              AuditLogEvent = 72
	MESSAGE_BULK_DELETE                         AuditLogEvent = 73
	MESSAGE_PIN                                 AuditLogEvent = 74
	MESSAGE_UNPIN                               AuditLogEvent = 75
	INTEGRATION_CREATE                          AuditLogEvent = 80
	INTEGRATION_UPDATE                          AuditLogEvent = 81
	INTEGRATION_DELETE                          AuditLogEvent = 82
	STAGE_INSTANCE_CREATE                       AuditLogEvent = 83
	STAGE_INSTANCE_UPDATE                       AuditLogEvent = 84
	STAGE_INSTANCE_DELETE                       AuditLogEvent = 85
	STICKER_CREATE                              AuditLogEvent = 90
	STICKER_UPDATE                              AuditLogEvent = 91
	STICKER_DELETE                              AuditLogEvent = 92
	GUILD_SCHEDULED_EVENT_CREATE                AuditLogEvent = 100
	GUILD_SCHEDULED_EVENT_UPDATE                AuditLogEvent = 101
	GUILD_SCHEDULED_EVENT_DELETE                AuditLogEvent = 102
	THREAD_CREATE                               AuditLogEvent = 110
	THREAD_UPDATE                               AuditLogEvent = 111
	THREAD_DELETE                               AuditLogEvent = 112
	APPLICATION_COMMAND_PERMISSION_UPDATE       AuditLogEvent = 121
	AUTO_MODERATION_RULE_CREATE                 AuditLogEvent = 140
	AUTO_MODERATION_RULE_UPDATE                 AuditLogEvent = 141
	AUTO_MODERATION_RULE_DELETE                 AuditLogEvent = 142
	AUTO_MODERATION_BLOCK_MESSAGE               AuditLogEvent = 143
	AUTO_MODERATION_FLAG_TO_CHANNEL             AuditLogEvent = 144
	AUTO_MODERATION_USER_COMMUNICATION_DISABLED AuditLogEvent = 145
	CREATOR_MONETIZATION_REQUEST_CREATED        AuditLogEvent = 150
	CREATOR_MONETIZATION_TERMS_ACCEPTED         AuditLogEvent = 151
)

// OptionalAuditEntryInfo represents optional information in an audit log entry.
type OptionalAuditEntryInfo struct {
	ApplicationID                 snowflake.ID `json:"application_id,omitempty,string"`
	AutoModerationRuleName        string       `json:"auto_moderation_rule_name,omitempty"`
	AutoModerationRuleTriggerType string       `json:"auto_moderation_rule_trigger_type,omitempty"`
	ChannelID                     snowflake.ID `json:"channel_id,omitempty,string"`
	Count                         string       `json:"count,omitempty"`
	DeleteMemberDays              string       `json:"delete_member_days,omitempty"`
	ID                            snowflake.ID `json:"id,omitempty,string"`
	MembersRemoved                string       `json:"members_removed,omitempty"`
	MessageID                     snowflake.ID `json:"message_id,omitempty,string"`
	RoleName                      string       `json:"role_name,omitempty"`
	Type                          string       `json:"type,omitempty"`
	IntegrationType               string       `json:"integration_type,omitempty"`
}

// AuditLogEntry represents a single entry in the audit log.
type AuditLogEntry struct {
	TargetID   *string                 `json:"target_id"`
	Changes    []AuditLogChange        `json:"changes,omitempty"`
	UserID     *snowflake.ID           `json:"user_id,string"`
	ID         snowflake.ID            `json:"id,string"`
	ActionType AuditLogEvent           `json:"action_type"`
	Options    *OptionalAuditEntryInfo `json:"options,omitempty"`
	Reason     string                  `json:"reason,omitempty"`
}

// AuditLogChange represents a change in an audit log entry.
type AuditLogChange struct {
	NewValue interface{} `json:"new_value,omitempty"`
	OldValue interface{} `json:"old_value,omitempty"`
	Key      string      `json:"key"`
}
