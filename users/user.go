package users

import (
	"github.com/discord-go/discord.go/snowflake"
)

// User represents a Discord user.
type User struct {
	ID                   snowflake.ID          `json:"id,string"`
	Username             string                `json:"username"`
	Discriminator        string                `json:"discriminator"`
	GlobalName           *string               `json:"global_name,omitempty"`
	Avatar               *string               `json:"avatar"`
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data,omitempty"`
	Bot                  bool                  `json:"bot,omitempty"`
	System               bool                  `json:"system,omitempty"`
	MfaEnabled           bool                  `json:"mfa_enabled,omitempty"`
	Banner               *string               `json:"banner,omitempty"`
	AccentColor          *int                  `json:"accent_color,omitempty"`
	Locale               string                `json:"locale,omitempty"`
	Verified             bool                  `json:"verified,omitempty"`
	Email                *string               `json:"email,omitempty"`
	Flags                Flag                  `json:"flags,omitempty"`
	PremiumType          PremiumType           `json:"premium_type,omitempty"`
	PublicFlags          Flag                  `json:"public_flags,omitempty"`
	Collectibles         *Collectibles         `json:"collectibles,omitempty"`
	PrimaryGuild         *PrimaryGuild         `json:"primary_guild,omitempty"`
}

type AvatarDecorationData struct {
	Asset     string       `json:"asset"`
	SKUID     snowflake.ID `json:"sku_id,string"`
	ExpiresAt *int         `json:"expires_at,omitempty"`
}

type Collectibles struct {
	Nameplate *Nameplate `json:"nameplate,omitempty"`
}

type Nameplate struct {
	Asset string       `json:"asset"`
	SKUID snowflake.ID `json:"sku_id,string"`
	Label string       `json:"label,omitempty"`
}

type PrimaryGuild struct {
	IdentityGuildID snowflake.ID `json:"identity_guild_id,string,omitempty"`
	IdentityEnabled bool         `json:"identity_enabled"`
	Tag             string       `json:"tag,omitempty"`
	Badge           string       `json:"badge,omitempty"`
}
