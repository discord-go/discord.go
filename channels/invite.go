package channels

import (
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type Invite struct {
	Code                     string       `json:"code"`
	Guild                    *InviteGuild `json:"guild,omitempty"`
	Channel                  *Channel     `json:"channel"`
	Inviter                  *users.User  `json:"inviter,omitempty"`
	TargetType               int          `json:"target_type,omitempty"`
	TargetUser               *users.User  `json:"target_user,omitempty"`
	TargetApplication        *Application `json:"target_application,omitempty"`
	ApproximatePresenceCount int          `json:"approximate_presence_count,omitempty"`
	ApproximateMemberCount   int          `json:"approximate_member_count,omitempty"`
	ExpiresAt                *time.Time   `json:"expires_at,omitempty"`

	// Metadata fields present on guild/channel invite listings and
	// INVITE_CREATE/INVITE_DELETE gateway payloads. Uses is the number of
	// times the invite has been accepted.
	Uses         int          `json:"uses,omitempty"`
	MaxUses      int          `json:"max_uses,omitempty"`
	MaxAge       int          `json:"max_age,omitempty"`
	Temporary    bool         `json:"temporary,omitempty"`
	CreatedAt    *time.Time   `json:"created_at,omitempty"`
	InviteType   int          `json:"type,omitempty"`
	GuildID      snowflake.ID `json:"guild_id,string,omitempty"`
	ChannelID    snowflake.ID `json:"channel_id,string,omitempty"`
	InviterID    snowflake.ID `json:"inviter_id,string,omitempty"`
	TargetUserID snowflake.ID `json:"target_user_id,string,omitempty"`
}

type InviteGuild struct {
	ID                snowflake.ID `json:"id,string"`
	Name              string       `json:"name"`
	Splash            *string      `json:"splash"`
	Banner            *string      `json:"banner"`
	Description       *string      `json:"description"`
	Icon              *string      `json:"icon"`
	Features          []string     `json:"features"`
	VerificationLevel int          `json:"verification_level"`
	VanityURLCode     *string      `json:"vanity_url_code"`
}

type Application struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Icon        *string      `json:"icon"`
	Description string       `json:"description"`
}
