package rest

import (
	"github.com/discord-go/discord.go/interactions"
)

// CreateCommandParams contains the parameters for creating an application command.
type CreateCommandParams struct {
	Name                     string                                  `json:"name"`
	Type                     *interactions.ApplicationCommandType    `json:"type,omitempty"`
	Description              string                                  `json:"description,omitempty"`
	Options                  []interactions.ApplicationCommandOption `json:"options,omitempty"`
	NameLocalizations        map[string]string                       `json:"name_localizations,omitempty"`
	DescriptionLocalizations map[string]string                       `json:"description_localizations,omitempty"`
	DefaultMemberPermissions *string                                 `json:"default_member_permissions,omitempty"`
	DMPermission             *bool                                   `json:"dm_permission,omitempty"`
	NSFW                     bool                                    `json:"nsfw,omitempty"`
	IntegrationTypes         []int                                   `json:"integration_types,omitempty"`
	Contexts                 []int                                   `json:"contexts,omitempty"`
	Handler                  *int                                    `json:"handler,omitempty"`
	DefaultPermission        *bool                                   `json:"default_permission,omitempty"`
}
