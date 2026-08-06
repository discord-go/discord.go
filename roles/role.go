package roles

import (
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

// Role represents a set of permissions attached to a group of users.
// https://discord.com/developers/docs/topics/permissions#role-object
type Role struct {
	ID           snowflake.ID           `json:"id,string"`
	Name         string                 `json:"name"`
	Color        int                    `json:"color"`
	Colors       *RoleColors            `json:"colors,omitempty"`
	Hoist        bool                   `json:"hoist"`
	Icon         *string                `json:"icon,omitempty"`
	UnicodeEmoji *string                `json:"unicode_emoji,omitempty"`
	Position     int                    `json:"position"`
	Permissions  permissions.Permission `json:"permissions,string"`
	Managed      bool                   `json:"managed"`
	Mentionable  bool                   `json:"mentionable"`
	Tags         *RoleTags              `json:"tags,omitempty"`
	Flags        int                    `json:"flags,omitempty"`
}

type RoleColors struct {
	PrimaryColor   int `json:"primary_color"`
	SecondaryColor int `json:"secondary_color"`
	TertiaryColor  int `json:"tertiary_color"`
}
