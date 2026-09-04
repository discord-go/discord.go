package interactions

import "github.com/discord-go/discord.go/snowflake"

// ApplicationCommandType represents the type of an application command.
type ApplicationCommandType int

const (
	ApplicationCommandTypeChatInput         ApplicationCommandType = 1
	ApplicationCommandTypeUser              ApplicationCommandType = 2
	ApplicationCommandTypeMessage           ApplicationCommandType = 3
	ApplicationCommandTypePrimaryEntryPoint ApplicationCommandType = 4
)

// ApplicationCommandOptionType represents the type of an option.
type ApplicationCommandOptionType int

const (
	ApplicationCommandOptionTypeSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionTypeSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionTypeString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionTypeInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionTypeBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionTypeUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionTypeChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionTypeRole            ApplicationCommandOptionType = 8
	ApplicationCommandOptionTypeMentionable     ApplicationCommandOptionType = 9
	ApplicationCommandOptionTypeNumber          ApplicationCommandOptionType = 10
	ApplicationCommandOptionTypeAttachment      ApplicationCommandOptionType = 11
)

// ApplicationCommandOption represents a slash command option.
type ApplicationCommandOption struct {
	Type                     ApplicationCommandOptionType     `json:"type"`
	Name                     string                           `json:"name"`
	Description              string                           `json:"description"`
	Required                 bool                             `json:"required,omitempty"`
	Choices                  []ApplicationCommandOptionChoice `json:"choices,omitempty"`
	Autocomplete             bool                             `json:"autocomplete,omitempty"`
	MinValue                 interface{}                      `json:"min_value,omitempty"`
	MaxValue                 interface{}                      `json:"max_value,omitempty"`
	NameLocalizations        map[string]string                `json:"name_localizations,omitempty"`
	DescriptionLocalizations map[string]string                `json:"description_localizations,omitempty"`
	MinLength                int                              `json:"min_length,omitempty"`
	MaxLength                int                              `json:"max_length,omitempty"`
	ChannelTypes             []int                            `json:"channel_types,omitempty"`
	Options                  []ApplicationCommandOption       `json:"options,omitempty"`
}

// ApplicationCommand represents a Discord application command.
type ApplicationCommand struct {
	ID                       snowflake.ID               `json:"id,string,omitempty"`
	Type                     *ApplicationCommandType    `json:"type,omitempty"`
	ApplicationID            snowflake.ID               `json:"application_id,string,omitempty"`
	// GuildID is set for guild-scoped commands; zero for global commands.
	GuildID                  snowflake.ID               `json:"guild_id,string,omitempty"`
	Name                     string                     `json:"name"`
	Description              string                     `json:"description"`
	Options                  []ApplicationCommandOption `json:"options,omitempty"`
	Version                  *snowflake.ID              `json:"version,string,omitempty"`
	IntegrationTypes         []int                      `json:"integration_types,omitempty"`
	Contexts                 []int                      `json:"contexts,omitempty"`
	NameLocalizations        map[string]string          `json:"name_localizations,omitempty"`
	DescriptionLocalizations map[string]string          `json:"description_localizations,omitempty"`
	DefaultMemberPermissions *string                    `json:"default_member_permissions,omitempty"`
	DMPermission             *bool                      `json:"dm_permission,omitempty"`
	DefaultPermission        *bool                      `json:"default_permission,omitempty"`
	NSFW                     bool                       `json:"nsfw,omitempty"`
	Handler                  *int                       `json:"handler,omitempty"`
}
